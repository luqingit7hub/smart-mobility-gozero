// ============================================================
// 【第6步·Stream 异步落库】抢单 Redis 事件 → MySQL 正式接单
//
// 为什么异步：抢单在 Redis Lua 里已完成，MySQL 写入稍慢且可重试，不阻塞司机「抢单成功」响应。
// 流程：GrabOrder Lua XADD → 本消费者 XREADGROUP → MySQL status=2 → MQ 通知乘客（第9步）
// 幂等：OrderUpdateGrabbed 使用 WHERE status=1；已接单且司机一致视为成功
// ============================================================
package pool

import (
	"common/config"
	"common/constants"
	"common/model"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	consumerBatchSize    = 50
	consumerBlockTimeout = 5 * time.Second
	consumerMaxDeliver   = 3
)

// OrderGrabbedConsumer 抢单 Stream 消费者
type OrderGrabbedConsumer struct {
	rdb          *redis.Client
	stream       string
	dlqStream    string
	group        string
	consumerName string
}

// NewOrderGrabbedConsumer 创建消费者（多实例可共用同一 group 负载均衡）
func NewOrderGrabbedConsumer() *OrderGrabbedConsumer {
	host, _ := os.Hostname()
	return &OrderGrabbedConsumer{
		rdb:          config.Rdb,
		stream:       constants.OrderGrabbedStream,
		dlqStream:    constants.OrderGrabbedStreamDLQ,
		group:        constants.OrderGrabbedGroup,
		consumerName: fmt.Sprintf("c-%s-%d", host, os.Getpid()),
	}
}

// StartOrderGrabbedConsumer 在 rpcOrder 启动时 go 调用，阻塞消费直到 ctx 取消
func StartOrderGrabbedConsumer(ctx context.Context) error {
	return NewOrderGrabbedConsumer().Start(ctx)
}

// Start 阻塞运行消费循环
func (c *OrderGrabbedConsumer) Start(ctx context.Context) error {
	if err := c.ensureGroup(ctx); err != nil {
		return fmt.Errorf("创建消费者组失败: %w", err)
	}
	fmt.Printf("[order-consumer] 已启动 stream=%s group=%s consumer=%s\n",
		c.stream, c.group, c.consumerName)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[order-consumer] 已停止")
			return nil
		default:
		}

		streams, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    c.group,
			Consumer: c.consumerName,
			Streams:  []string{c.stream, ">"},
			Count:    consumerBatchSize,
			Block:    consumerBlockTimeout,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) || errors.Is(err, context.Canceled) {
				continue
			}
			time.Sleep(time.Second)
			continue
		}

		for _, s := range streams {
			for _, msg := range s.Messages {
				c.handleOne(ctx, msg)
			}
		}
	}
}

func (c *OrderGrabbedConsumer) ensureGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(ctx, c.stream, c.group, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func (c *OrderGrabbedConsumer) handleOne(ctx context.Context, msg redis.XMessage) {
	orderNo, _ := msg.Values["order_no"].(string)
	driverId := parseStreamInt64(msg.Values["driver_id"])

	if orderNo == "" || driverId <= 0 {
		fmt.Printf("[order-consumer] 消息无效 %s values=%v\n", msg.ID, msg.Values)
		c.toDLQ(ctx, msg, "invalid payload")
		c.ack(ctx, msg.ID)
		return
	}

	if err := c.updateOrderGrabbed(orderNo, driverId); err != nil {
		deliveries := c.deliveryCount(ctx, msg.ID)
		if deliveries >= consumerMaxDeliver {
			fmt.Printf("[order-consumer] order=%s 超过重试次数，转入DLQ: %v\n", orderNo, err)
			c.toDLQ(ctx, msg, err.Error())
			c.ack(ctx, msg.ID)
			return
		}
		fmt.Printf("[order-consumer] order=%s 落库失败 deliver=%d: %v (将重试)\n",
			orderNo, deliveries, err)
		return
	}

	publishOrderGrabbedNotify(orderNo, driverId)
	c.ack(ctx, msg.ID)
}

// updateOrderGrabbed 更新 MySQL；已接单且司机一致时幂等成功
func (c *OrderGrabbedConsumer) updateOrderGrabbed(orderNo string, driverId int64) error {
	order := &model.Order{}
	if err := order.OrderUpdateGrabbed(config.DB, orderNo, driverId); err != nil {
		var exist model.Order
		if findErr := exist.OrderModelFindNumber(config.DB, orderNo); findErr == nil &&
			exist.Status == model.OrderStatusAccepted && exist.DriverId == driverId {
			return nil
		}
		return err
	}
	fmt.Printf("[order-consumer] 落库成功 order=%s driver=%d\n", orderNo, driverId)
	return nil
}

func (c *OrderGrabbedConsumer) ack(ctx context.Context, id string) {
	if err := c.rdb.XAck(ctx, c.stream, c.group, id).Err(); err != nil {
		fmt.Printf("[order-consumer] XACK 失败 id=%s err=%v\n", id, err)
	}
}

func (c *OrderGrabbedConsumer) deliveryCount(ctx context.Context, id string) int64 {
	res, err := c.rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: c.stream,
		Group:  c.group,
		Start:  id,
		End:    id,
		Count:  1,
	}).Result()
	if err != nil || len(res) == 0 {
		return 1
	}
	return res[0].RetryCount
}

func (c *OrderGrabbedConsumer) toDLQ(ctx context.Context, msg redis.XMessage, reason string) {
	values := make(map[string]interface{}, len(msg.Values)+3)
	for k, v := range msg.Values {
		values[k] = v
	}
	values["_original_id"] = msg.ID
	values["_failed_at"] = time.Now().Unix()
	values["_reason"] = reason
	if err := c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: c.dlqStream,
		MaxLen: 10000,
		Approx: true,
		Values: values,
	}).Err(); err != nil {
		fmt.Printf("[order-consumer] 写入DLQ失败: %v\n", err)
	}
}

func parseStreamInt64(v interface{}) int64 {
	switch x := v.(type) {
	case string:
		n, _ := strconv.ParseInt(x, 10, 64)
		return n
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	default:
		return 0
	}
}
