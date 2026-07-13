// Package rmq 【第3步·RabbitMQ】延迟队列 + 实时通知队列的发布与消费。
//
// 在本项目中的作用：
//   - 延迟队列（DLX 死信）：TakeCar 下单时投递，到期后由 rpcOrder 的 HandleOrderDelay 消费（第7步）
//   - 通知队列：抢单落库/取消/完单/推司机时发布，由 apiGateway 消费并推 WebSocket（第9步）
package rmq

import (
	"common/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// 延迟队列拓扑（amap-lq 专用命名，避免与其他项目冲突）
const (
	// DelayQueue 延迟队列：消息带 per-message TTL，过期后进入死信交换机
	DelayQueue = "amap-lq:order:delay"
	// DLXExchange 死信交换机
	DLXExchange = "amap-lq:order:dlx"
	// DelayRoutingKey 死信路由键
	DelayRoutingKey = "order.delay"
	// ConsumeQueue 延迟到期后的真正消费队列
	ConsumeQueue = "amap-lq:order:consume"

	// 订单实时通知（Direct 路由，第 9～10 步使用）
	NotifyExchange    = "amap-lq:order:notify:exchange"
	NotifyQueue       = "amap-lq:order:notify:queue"
	NotifyKeyGrabbed    = "order.grabbed"
	NotifyKeyCancelled  = "order.cancelled"
	NotifyKeyCompleted  = "order.completed"
	NotifyKeyNearby         = "order.nearby"
	NotifyKeyPushedDrivers  = "order.pushed_drivers"
	NotifyKeyTripStarted    = "order.trip_started"

	confirmTimeout   = 5 * time.Second
	consumerMaxRetry = 3
)

var (
	mqMu       sync.Mutex
	mqConn     *amqp.Connection
	mqDeclared bool
)

func mqURL() string {
	return config.DataConfig.RabbitMq.Host
}

// dial 建立连接（懒加载，发布/消费共用）
func dial() (*amqp.Connection, error) {
	mqMu.Lock()
	defer mqMu.Unlock()

	if mqConn != nil && !mqConn.IsClosed() {
		return mqConn, nil
	}
	conn, err := amqp.Dial(mqURL())
	if err != nil {
		return nil, fmt.Errorf("连接 RabbitMQ 失败: %w", err)
	}
	mqConn = conn
	mqDeclared = false
	return conn, nil
}

// declareDelayTopology 声明延迟队列 + 死信交换机 + 消费队列
func declareDelayTopology(ch *amqp.Channel) error {
	if _, err := ch.QueueDeclare(DelayQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    DLXExchange,
		"x-dead-letter-routing-key": DelayRoutingKey,
		// 不设队列级 TTL，每条消息用 Publishing.Expiration 单独指定延迟
	}); err != nil {
		return fmt.Errorf("声明延迟队列失败: %w", err)
	}
	if err := ch.ExchangeDeclare(DLXExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明死信交换机失败: %w", err)
	}
	if _, err := ch.QueueDeclare(ConsumeQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明消费队列失败: %w", err)
	}
	if err := ch.QueueBind(ConsumeQueue, DelayRoutingKey, DLXExchange, false, nil); err != nil {
		return fmt.Errorf("绑定消费队列失败: %w", err)
	}
	return nil
}

// declareNotifyTopology 声明订单通知交换机与队列（抢单/取消/附近新单）
func declareNotifyTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(NotifyExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明通知交换机失败: %w", err)
	}
	if _, err := ch.QueueDeclare(NotifyQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明通知队列失败: %w", err)
	}
	for _, key := range []string{NotifyKeyGrabbed, NotifyKeyCancelled, NotifyKeyCompleted, NotifyKeyNearby, NotifyKeyPushedDrivers, NotifyKeyTripStarted} {
		if err := ch.QueueBind(NotifyQueue, key, NotifyExchange, false, nil); err != nil {
			return fmt.Errorf("绑定通知队列 key=%s 失败: %w", key, err)
		}
	}
	return nil
}

func ensureTopology(ch *amqp.Channel) error {
	mqMu.Lock()
	defer mqMu.Unlock()
	if mqDeclared {
		return nil
	}
	if err := declareDelayTopology(ch); err != nil {
		return err
	}
	if err := declareNotifyTopology(ch); err != nil {
		return err
	}
	mqDeclared = true
	log.Println("[rmq] 拓扑声明完成: delay + notify")
	return nil
}

// publishWithConfirm 发布并等待 Broker Confirm
func publishWithConfirm(ch *amqp.Channel, exchange, routingKey string, pub amqp.Publishing) error {
	if err := ch.Confirm(false); err != nil {
		return fmt.Errorf("开启 Confirm 失败: %w", err)
	}
	ackCh, nackCh := ch.NotifyConfirm(make(chan uint64, 1), make(chan uint64, 1))

	if err := ch.Publish(exchange, routingKey, false, false, pub); err != nil {
		return fmt.Errorf("Publish 失败: %w", err)
	}
	select {
	case <-ackCh:
		return nil
	case <-nackCh:
		return errors.New("Broker 返回 Nack，消息未确认")
	case <-time.After(confirmTimeout):
		return errors.New("等待 Confirm 超时")
	}
}

// PublishDelay 【第4步调用·第7步消费】下单时投递延迟任务（TTL 到期后进入 ConsumeQueue）。
func PublishDelay(msg *OrderDelayMsg, ttlMs int64) error {
	if msg == nil || msg.OrderNo == "" || msg.Action == "" {
		return errors.New("延迟消息参数无效")
	}
	if ttlMs <= 0 {
		return errors.New("延迟时间必须大于 0")
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("打开 Channel 失败: %w", err)
	}
	defer ch.Close()

	if err := ensureTopology(ch); err != nil {
		return err
	}

	pub := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
		Expiration:   strconv.FormatInt(ttlMs, 10), // per-message TTL（毫秒）
	}
	if err := publishWithConfirm(ch, "", DelayQueue, pub); err != nil {
		return err
	}
	log.Printf("[rmq] 延迟消息已确认 order=%s action=%s ttlMs=%d", msg.OrderNo, msg.Action, ttlMs)
	return nil
}

// PublishNotify 发布实时通知到通知交换机
func PublishNotify(routingKey string, evt *OrderNotifyEvent) error {
	if evt == nil {
		return errors.New("通知事件不能为空")
	}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	conn, err := dial()
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("打开 Channel 失败: %w", err)
	}
	defer ch.Close()

	if err := ensureTopology(ch); err != nil {
		return err
	}

	pub := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	}
	if err := publishWithConfirm(ch, NotifyExchange, routingKey, pub); err != nil {
		return err
	}
	log.Printf("[rmq] 通知已发布 key=%s order=%s event=%s", routingKey, evt.OrderNo, evt.Event)
	return nil
}

// PublishOrderGrabbed 司机接单通知（落库成功后调用）
func PublishOrderGrabbed(evt *OrderNotifyEvent) error {
	if evt.Event == "" {
		evt.Event = EventDriverAccepted
	}
	return PublishNotify(NotifyKeyGrabbed, evt)
}

// PublishOrderCancelled 无人接单/取消通知
func PublishOrderCancelled(evt *OrderNotifyEvent) error {
	if evt.Event == "" {
		evt.Event = EventOrderCancelled
	}
	return PublishNotify(NotifyKeyCancelled, evt)
}

// PublishOrderNearby 附近新单通知（推给司机）
func PublishOrderNearby(evt *OrderNotifyEvent) error {
	if evt.Event == "" {
		evt.Event = EventNewOrderNearby
	}
	return PublishNotify(NotifyKeyNearby, evt)
}

// PublishOrderPushedDrivers 延迟推司机完成后通知乘客（推给乘客 WebSocket）
func PublishOrderPushedDrivers(evt *OrderNotifyEvent) error {
	if evt.Event == "" {
		evt.Event = EventOrderPushedDrivers
	}
	return PublishNotify(NotifyKeyPushedDrivers, evt)
}

// PublishOrderCompleted 完单通知（推给乘客）
func PublishOrderCompleted(evt *OrderNotifyEvent) error {
	if evt.Event == "" {
		evt.Event = EventOrderCompleted
	}
	return PublishNotify(NotifyKeyCompleted, evt)
}

// PublishTripStarted 行程开始通知（推给乘客）
func PublishTripStarted(evt *OrderNotifyEvent) error {
	if evt.Event == "" {
		evt.Event = EventTripStarted
	}
	return PublishNotify(NotifyKeyTripStarted, evt)
}

// NotifyHandler 实时通知消费回调（apiGateway WebSocket 用）
type NotifyHandler func(body []byte) error

// StartConsumerWithRetry 启动 MQ 消费者；连接失败时按间隔重试，避免进程先起、RabbitMQ 未就绪导致永久无消费
func StartConsumerWithRetry(ctx context.Context, name string, start func(context.Context) error) {
	go func() {
		attempt := 0
		for {
			if ctx.Err() != nil {
				return
			}
			attempt++
			if err := start(ctx); err != nil {
				log.Printf("[rmq] %s 启动失败 attempt=%d: %v，5s 后重试", name, attempt, err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
				}
				continue
			}
			log.Printf("[rmq] %s 已启动", name)
			return
		}
	}()
}

// DelayHandler 延迟消息消费回调，返回 error 时按策略重试
type DelayHandler func(body []byte) error

// StartDelayConsumer 【第7步】rpcOrder 启动时注册，消费延迟到期任务（pool.HandleOrderDelay）。
func StartDelayConsumer(ctx context.Context, handler DelayHandler) error {
	if handler == nil {
		return errors.New("handler 不能为空")
	}
	conn, err := dial()
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("消费 Channel 失败: %w", err)
	}
	if err := ensureTopology(ch); err != nil {
		ch.Close()
		return err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		return err
	}

	deliveries, err := ch.Consume(ConsumeQueue, "", false, false, false, false, nil)
	if err != nil {
		ch.Close()
		return fmt.Errorf("注册消费者失败: %w", err)
	}

	go func() {
		defer ch.Close()
		log.Printf("[rmq] 延迟消费者已启动 queue=%s", ConsumeQueue)
		for {
			select {
			case <-ctx.Done():
				log.Println("[rmq] 延迟消费者已停止")
				return
			case d, ok := <-deliveries:
				if !ok {
					log.Println("[rmq] 消费通道已关闭")
					return
				}
				if err := handler(d.Body); err != nil {
					retry := getRetryCount(d)
					log.Printf("[rmq] 延迟消息处理失败 retry=%d: %v", retry, err)
					if retry >= consumerMaxRetry {
						log.Printf("[rmq] 超过最大重试，丢弃消息 body=%s", string(d.Body))
						_ = d.Ack(false)
					} else {
						_ = d.Nack(false, true)
					}
				} else {
					_ = d.Ack(false)
				}
			}
		}
	}()
	return nil
}

// StartNotifyConsumer 【第9步】apiGateway 启动时注册，把 MQ 通知交给 ws 包推 WebSocket。
func StartNotifyConsumer(ctx context.Context, handler NotifyHandler) error {
	if handler == nil {
		return errors.New("handler 不能为空")
	}
	conn, err := dial()
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("通知消费 Channel 失败: %w", err)
	}
	if err := ensureTopology(ch); err != nil {
		ch.Close()
		return err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		return err
	}

	deliveries, err := ch.Consume(NotifyQueue, "", false, false, false, false, nil)
	if err != nil {
		ch.Close()
		return fmt.Errorf("注册通知消费者失败: %w", err)
	}

	go func() {
		defer ch.Close()
		log.Printf("[rmq] 通知消费者已启动 queue=%s", NotifyQueue)
		for {
			select {
			case <-ctx.Done():
				log.Println("[rmq] 通知消费者已停止")
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				if err := handler(d.Body); err != nil {
					retry := getRetryCount(d)
					if retry >= consumerMaxRetry {
						_ = d.Ack(false)
					} else {
						_ = d.Nack(false, true)
					}
				} else {
					_ = d.Ack(false)
				}
			}
		}
	}()
	return nil
}

// getRetryCount 从 x-death 头估算重试次数（RabbitMQ 重新入队会累加）
func getRetryCount(d amqp.Delivery) int {
	if d.Headers == nil {
		if d.Redelivered {
			return 1
		}
		return 0
	}
	if deaths, ok := d.Headers["x-death"].([]interface{}); ok && len(deaths) > 0 {
		if table, ok := deaths[0].(amqp.Table); ok {
			if count, ok := table["count"].(int64); ok {
				return int(count)
			}
		}
	}
	if d.Redelivered {
		return 1
	}
	return 0
}
