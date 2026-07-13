// Package pool 【第6/8/9步·MQ 通知发布】把业务结果转成 OrderNotifyEvent 发到 RabbitMQ。
//
// 在本项目中的作用：rpcOrder 落库/完单成功后，发 MQ → apiGateway 消费 → WebSocket 推给乘客。
// 发布失败只打日志，不影响主流程（抢单落库仍会 ACK Stream）。
package pool

import (
	"common/config"
	"common/model"
	"common/rmq"
	"fmt"
	"time"
)

// publishOrderGrabbedNotify 落库成功后通知乘客（发布失败只打日志，不影响 Stream ACK）
func publishOrderGrabbedNotify(orderNo string, driverId int64) {
	var order model.Order
	if err := order.OrderModelFindNumber(config.DB, orderNo); err != nil {
		fmt.Printf("[order-notify] 查询订单失败 order=%s err=%v\n", orderNo, err)
		return
	}

	var driver model.Driver
	if err := driver.DriverModelFindId(config.DB, driverId); err != nil {
		fmt.Printf("[order-notify] 查询司机失败 driver=%d err=%v\n", driverId, err)
	}

	acceptAt := time.Now().Unix()
	if order.AcceptTime != nil {
		acceptAt = order.AcceptTime.Unix()
	}

	evt := &rmq.OrderNotifyEvent{
		Event:      rmq.EventDriverAccepted,
		OrderNo:    orderNo,
		UserId:     order.UserId,
		DriverId:   driverId,
		DriverName: driver.Name,
		CarNumber:  driver.CarNumber,
		CarType:    driver.CarType,
		Rating:     driver.Rating,
		AcceptAt:   acceptAt,
	}
	if err := rmq.PublishOrderGrabbed(evt); err != nil {
		fmt.Printf("[order-notify] 发布失败 order=%s user=%d err=%v\n", orderNo, order.UserId, err)
		return
	}
	fmt.Printf("[order-notify] 已发布 order=%s user=%d driver=%d\n", orderNo, order.UserId, driverId)
}

// PublishOrderCompletedNotify 完单成功后通知乘客（发布失败只打日志，不影响完单结果）
func PublishOrderCompletedNotify(orderNo string, userId, driverId int64, payPrice float64) {
	var driver model.Driver
	if err := driver.DriverModelFindId(config.DB, driverId); err != nil {
		fmt.Printf("[order-notify] 完单查询司机失败 driver=%d err=%v\n", driverId, err)
	}

	evt := &rmq.OrderNotifyEvent{
		Event:      rmq.EventOrderCompleted,
		OrderNo:    orderNo,
		UserId:     userId,
		DriverId:   driverId,
		DriverName: driver.Name,
		CarNumber:  driver.CarNumber,
		CarType:    driver.CarType,
		Rating:     driver.Rating,
		Msg:        fmt.Sprintf("订单已完成，实付%.2f元", payPrice),
	}
	if err := rmq.PublishOrderCompleted(evt); err != nil {
		fmt.Printf("[order-notify] 完单发布失败 order=%s user=%d err=%v\n", orderNo, userId, err)
		return
	}
	fmt.Printf("[order-notify] 完单已发布 order=%s user=%d pay=%.2f\n", orderNo, userId, payPrice)
}

// PublishTripStartedNotify 司机确认乘客上车后通知乘客（发布失败只打日志）
func PublishTripStartedNotify(orderNo string, userId, driverId int64) {
	var driver model.Driver
	if err := driver.DriverModelFindId(config.DB, driverId); err != nil {
		fmt.Printf("[order-notify] 开始行程查询司机失败 driver=%d err=%v\n", driverId, err)
	}

	evt := &rmq.OrderNotifyEvent{
		Event:      rmq.EventTripStarted,
		OrderNo:    orderNo,
		UserId:     userId,
		DriverId:   driverId,
		DriverName: driver.Name,
		CarNumber:  driver.CarNumber,
		CarType:    driver.CarType,
		Rating:     driver.Rating,
		Msg:        "行程已开始，请系好安全带，祝您旅途愉快",
	}
	if err := rmq.PublishTripStarted(evt); err != nil {
		fmt.Printf("[order-notify] 开始行程发布失败 order=%s user=%d err=%v\n", orderNo, userId, err)
		return
	}
	fmt.Printf("[order-notify] 开始行程已发布 order=%s user=%d\n", orderNo, userId)
}
