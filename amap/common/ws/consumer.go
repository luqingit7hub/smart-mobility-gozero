// Package ws 【第9步】RabbitMQ 通知 → WebSocket 推送的桥接消费者。
package ws

import (
	"common/rmq"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// StartOrderNotifyConsumer 消费 RabbitMQ 订单通知，推送到 WebSocket
func StartOrderNotifyConsumer(ctx context.Context, hub *Hub) error {
	if hub == nil {
		return fmt.Errorf("hub 不能为空")
	}
	return rmq.StartNotifyConsumer(ctx, func(body []byte) error {
		var evt rmq.OrderNotifyEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			log.Printf("[ws-consumer] JSON 解析失败: %v", err)
			return nil
		}

		var ok bool
		switch evt.Event {
		case rmq.EventNewOrderNearby:
			// 仅推司机：附近新单
			if evt.DriverId <= 0 {
				log.Printf("[ws-consumer] new_order_nearby 缺少 driver_id, drop order=%s", evt.OrderNo)
				return nil
			}
			ok = hub.PushDriver(evt.DriverId, body)
			if ok {
				log.Printf("[ws-consumer] 推司机 driver=%d order=%s event=%s", evt.DriverId, evt.OrderNo, evt.Event)
			} else {
				log.Printf("[ws-consumer] 司机离线 driver=%d order=%s", evt.DriverId, evt.OrderNo)
			}

		case rmq.EventOrderPushedDrivers,
			rmq.EventDriverAccepted,
			rmq.EventOrderCancelled,
			rmq.EventOrderCompleted,
			rmq.EventTripStarted:
			// 仅推乘客：推司机结果 / 接单 / 取消 / 完单
			if evt.UserId <= 0 {
				log.Printf("[ws-consumer] %s 缺少 user_id, drop order=%s", evt.Event, evt.OrderNo)
				return nil
			}
			ok = hub.PushUser(evt.UserId, body)
			if ok {
				log.Printf("[ws-consumer] 推乘客 user=%d order=%s event=%s", evt.UserId, evt.OrderNo, evt.Event)
			} else {
				log.Printf("[ws-consumer] 乘客离线 user=%d order=%s", evt.UserId, evt.OrderNo)
			}

		default:
			log.Printf("[ws-consumer] 未知 event=%s order=%s, drop", evt.Event, evt.OrderNo)
		}
		return nil
	})
}
