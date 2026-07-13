// Package pool 【乘客取消】待接单订单的统一取消逻辑（用户主动取消与超时取消共用）。
//
// 在本项目中的作用：MySQL status 1→4 + RemoveFromPool 清 Redis + 发 order_cancelled 通知。
// 已接单（status=2 或 Redis 已 grabbed）不允许乘客取消。
package pool

import (
	"common/config"
	"common/constants"
	"common/model"
	"common/rmq"
	"context"
	"errors"
	"fmt"
)

// CancelWaitingOrder 取消待接单订单：校验归属与状态 → MySQL 1→4 → 清 Redis 抢单池 → MQ 通知乘客。
// userId>0 时校验订单归属；userId=0 表示系统取消（如超时），跳过归属校验。
func CancelWaitingOrder(ctx context.Context, orderNo string, userId int64, reason string) error {
	if orderNo == "" {
		return errors.New("订单号不能为空")
	}
	if reason == "" {
		reason = "用户主动取消"
	}

	var order model.Order
	if err := order.OrderModelFindNumber(config.DB, orderNo); err != nil {
		return errors.New("订单不存在")
	}
	if userId > 0 && order.UserId != userId {
		return errors.New("无权取消该订单")
	}
	if order.Status == model.OrderStatusCancelled {
		return nil
	}
	if order.Status == model.OrderStatusAccepted || order.Status == model.OrderStatusOnBoard {
		return errors.New("司机已接单，无法取消")
	}
	if order.Status == model.OrderStatusCompleted {
		return errors.New("订单已完成，无法取消")
	}
	if order.Status != model.OrderStatusWaiting {
		return errors.New("订单状态异常，无法取消")
	}

	// 抢单 Lua 已标记 grabbed 但 MySQL 尚未落库时，拒绝用户取消
	if st, err := config.Rdb.HGet(ctx, constants.OrderCacheKey(orderNo), "status").Result(); err == nil && st == constants.PoolStatusGrabbed {
		return errors.New("司机已接单，无法取消")
	}

	if err := order.OrderUpdateStatus(config.DB, orderNo, model.OrderStatusWaiting, model.OrderStatusCancelled, map[string]interface{}{
		"cancel_reason": reason,
		"driver_id":     0,
	}); err != nil {
		return fmt.Errorf("取消订单失败: %w", err)
	}

	RemoveFromPool(ctx, orderNo)

	evt := &rmq.OrderNotifyEvent{
		Event:   rmq.EventOrderCancelled,
		OrderNo: orderNo,
		UserId:  order.UserId,
		Msg:     reason,
	}
	if err := rmq.PublishOrderCancelled(evt); err != nil {
		fmt.Printf("[cancel-order] 取消通知发布失败 order=%s err=%v\n", orderNo, err)
	}
	fmt.Printf("[cancel-order] 完成 order=%s user=%d reason=%s\n", orderNo, order.UserId, reason)
	return nil
}
