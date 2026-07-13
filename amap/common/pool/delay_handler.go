// Package pool 【第7步·延迟兜底】消费 RabbitMQ 延迟到期任务。
//
// 在本项目中的作用：乘客下单后若一直无人接单，系统自动：
//   - 6 分钟：给附近在线司机发 WebSocket，并通知乘客「已推送给 N 公里内司机」
//   - 10 分钟：将订单置为已取消，清 Redis 池，通知乘客
//
// 若订单已被抢/已取消/已完成，消费时查 MySQL 状态后直接跳过（幂等）。
package pool

import (
	"common/config"
	"common/model"
	"common/rmq"
	"encoding/json"
	"fmt"
	"math"
)

const nearbyPushRadiusKm = 20.0 // 与 GrabListNearby 默认半径一致

// HandleOrderDelay 延迟任务入口（注册到 rmq.StartDelayConsumer）
func HandleOrderDelay(body []byte) error {
	var msg rmq.OrderDelayMsg
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("解析延迟消息失败: %w", err)
	}
	if msg.OrderNo == "" || msg.Action == "" {
		return fmt.Errorf("延迟消息字段无效")
	}
	switch msg.Action {
	case rmq.ActionPushDrivers:
		return handlePushDrivers(msg.OrderNo)
	case rmq.ActionCancelOrder:
		return handleCancelOrder(msg.OrderNo)
	default:
		fmt.Printf("[delay] 未知 action=%s order=%s，已忽略\n", msg.Action, msg.OrderNo)
		return nil
	}
}

// handlePushDrivers 6 分钟：仍待接单则通知附近在线司机，并告知乘客推送结果
func handlePushDrivers(orderNo string) error {
	var order model.Order
	if err := order.OrderModelFindNumber(config.DB, orderNo); err != nil {
		fmt.Printf("[delay] push_drivers 订单不存在 order=%s\n", orderNo)
		return nil
	}
	if order.Status != model.OrderStatusWaiting {
		fmt.Printf("[delay] push_drivers 订单已结束 order=%s status=%d\n", orderNo, order.Status)
		return nil
	}

	drivers, err := (&model.Driver{}).DriverListOnline(config.DB)
	if err != nil {
		return fmt.Errorf("查询在线司机失败: %w", err)
	}

	pushed := 0
	for _, d := range drivers {
		if distanceKm(d.CurrentLng, d.CurrentLat, order.StartLng, order.StartLat) > nearbyPushRadiusKm {
			continue
		}
		// 司机 WS：event=new_order_nearby，文案与字段与乘客通知完全区分
		driverEvt := &rmq.OrderNotifyEvent{
			Event:         rmq.EventNewOrderNearby,
			OrderNo:       orderNo,
			DriverId:      int64(d.ID),
			StartAddress:  order.StartAddress,
			Price:         order.Price,
			Distance:      order.Distance,
			Msg:           fmt.Sprintf("附近有新订单！起点：%s，预估 %.2f 元，约 %.1f 公里，请打开抢单列表", order.StartAddress, order.Price, order.Distance),
		}
		if err := rmq.PublishOrderNearby(driverEvt); err != nil {
			fmt.Printf("[delay] 推送司机失败 driver=%d order=%s err=%v\n", d.ID, orderNo, err)
			continue
		}
		pushed++
	}

	// 乘客 WS：event=order_pushed_drivers，仅乘客端展示推送结果（不含司机业务字段）
	userMsg := fmt.Sprintf("您的订单已主动推送给%.0f公里范围内的%d位司机，请耐心等待", nearbyPushRadiusKm, pushed)
	if pushed == 0 {
		userMsg = fmt.Sprintf("您的订单已主动推送给%.0f公里范围内的司机，当前暂无在线司机，请耐心等待", nearbyPushRadiusKm)
	}
	userEvt := &rmq.OrderNotifyEvent{
		Event:             rmq.EventOrderPushedDrivers,
		OrderNo:           orderNo,
		UserId:            order.UserId,
		Msg:               userMsg,
		PushRadiusKm:      nearbyPushRadiusKm,
		PushedDriverCount: pushed,
	}
	if err := rmq.PublishOrderPushedDrivers(userEvt); err != nil {
		fmt.Printf("[delay] 推送乘客失败 order=%s user=%d err=%v\n", orderNo, order.UserId, err)
	}

	fmt.Printf("[delay] push_drivers 完成 order=%s 通知司机数=%d\n", orderNo, pushed)
	return nil
}

// handleCancelOrder 10 分钟：仍待接单则取消并通知乘客
func handleCancelOrder(orderNo string) error {
	var order model.Order
	if err := order.OrderModelFindNumber(config.DB, orderNo); err != nil {
		fmt.Printf("[delay] cancel_order 订单不存在 order=%s\n", orderNo)
		return nil
	}
	if order.Status != model.OrderStatusWaiting {
		fmt.Printf("[delay] cancel_order 订单已结束 order=%s status=%d\n", orderNo, order.Status)
		return nil
	}
	if err := CancelWaitingOrder(config.Ctx, orderNo, 0, "超时无人接单"); err != nil {
		return fmt.Errorf("取消订单失败: %w", err)
	}
	fmt.Printf("[delay] cancel_order 完成 order=%s\n", orderNo)
	return nil
}

func distanceKm(lng1, lat1, lng2, lat2 float64) float64 {
	const earthRadiusKm = 6371.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}
