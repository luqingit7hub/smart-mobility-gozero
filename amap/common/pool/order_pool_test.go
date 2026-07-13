package pool

import (
	"common/config"
	"common/constants"
	"common/model"
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// setupTestRedis 用内存 Redis 替代真实实例，测试不依赖 docker。
func setupTestRedis(t *testing.T) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("启动 miniredis 失败: %v", err)
	}
	config.Rdb = redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = config.Rdb.Close()
		mr.Close()
	})
}

func seedPendingOrder(t *testing.T, orderNo string) {
	t.Helper()
	ctx := context.Background()
	o := &model.Order{
		OrderNo:      orderNo,
		UserId:       1001,
		StartLng:     116.397128,
		StartLat:     39.916527,
		StartAddress: "起点",
		EndLng:       116.407128,
		EndLat:       39.926527,
		EndAddress:   "终点",
		Distance:     5.2,
		Duration:     900,
		Price:        18.5,
	}
	if err := PublishOrderToPool(ctx, o); err != nil {
		t.Fatalf("写入抢单池失败: %v", err)
	}
}

// TestRunGrabOrder_ConcurrentOnlyOneWinner 多司机并发抢同一单，仅一人成功。
func TestRunGrabOrder_ConcurrentOnlyOneWinner(t *testing.T) {
	setupTestRedis(t)
	const orderNo = "ORD-CONCURRENT-001"
	seedPendingOrder(t, orderNo)

	const workers = 20
	var okCount atomic.Int32
	var lostCount atomic.Int32
	var wg sync.WaitGroup
	wg.Add(workers)

	ctx := context.Background()
	for i := 1; i <= workers; i++ {
		driverID := int64(10000 + i)
		go func() {
			defer wg.Done()
			code, _, err := RunGrabOrder(ctx, orderNo, driverID)
			if err != nil {
				t.Errorf("抢单返回 error: %v", err)
				return
			}
			switch code {
			case GrabCodeOK:
				okCount.Add(1)
			case GrabCodeTaken, GrabCodeExpired:
				// TAKEN：与赢家并发竞争；EXPIRED：赢家已 RemoveFromPool 删单
				lostCount.Add(1)
			default:
				t.Errorf("driver=%d 意外业务码 code=%d", driverID, code)
			}
		}()
	}
	wg.Wait()

	if got := okCount.Load(); got != 1 {
		t.Fatalf("成功抢单数=%d，期望恰好 1", got)
	}
	if got := lostCount.Load(); got != workers-1 {
		t.Fatalf("未抢中数=%d，期望 %d", got, workers-1)
	}

	streamLen, err := config.Rdb.XLen(ctx, constants.OrderGrabbedStream).Result()
	if err != nil {
		t.Fatalf("读取 Stream 长度失败: %v", err)
	}
	if streamLen != 1 {
		t.Fatalf("Stream 事件数=%d，期望 1（抢中与入流原子一致）", streamLen)
	}
}

// TestRunGrabOrder_AlreadyTaken 订单已被标记 grabbed 时，后续抢单返回 TAKEN。
func TestRunGrabOrder_AlreadyTaken(t *testing.T) {
	setupTestRedis(t)
	const orderNo = "ORD-TAKEN-001"
	ctx := context.Background()
	cacheKey := constants.OrderCacheKey(orderNo)
	now := time.Now().Unix()

	if err := config.Rdb.HSet(ctx, cacheKey, map[string]interface{}{
		"order_no":   orderNo,
		"status":     constants.PoolStatusGrabbed,
		"winner":     20001,
		"expires_at": now + constants.OrderPoolTTL,
	}).Err(); err != nil {
		t.Fatalf("预置订单失败: %v", err)
	}

	code, msg, err := RunGrabOrder(ctx, orderNo, 20002)
	if err != nil {
		t.Fatalf("抢单返回 error: %v", err)
	}
	if code != GrabCodeTaken {
		t.Fatalf("code=%d msg=%s，期望 GrabCodeTaken", code, msg)
	}
}

// TestRunGrabOrder_DriverBusy 司机已有进行中订单时，不能再抢新单。
func TestRunGrabOrder_DriverBusy(t *testing.T) {
	setupTestRedis(t)
	const orderNo = "ORD-BUSY-001"
	seedPendingOrder(t, orderNo)

	ctx := context.Background()
	const driverID int64 = 30001
	if err := config.Rdb.Set(ctx, constants.DriverOrderKey(driverID), "OTHER-ORDER", 0).Err(); err != nil {
		t.Fatalf("预置司机忙锁失败: %v", err)
	}

	code, msg, err := RunGrabOrder(ctx, orderNo, driverID)
	if err != nil {
		t.Fatalf("抢单返回 error: %v", err)
	}
	if code != GrabCodeBusy {
		t.Fatalf("code=%d msg=%s，期望 GrabCodeBusy", code, msg)
	}
}
