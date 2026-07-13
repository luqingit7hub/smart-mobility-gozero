// Package rmq 【第3步·消息体】RabbitMQ 里流转的 JSON 结构定义。
//
// 在本项目中有两类消息：
//  1. OrderDelayMsg — 延迟任务（第7步）：下单后 6 分钟推司机、10 分钟无人接单自动取消
//  2. OrderNotifyEvent — 实时通知（第9步）：接单/取消/完单/附近新单，最终推到 WebSocket
package rmq

// 延迟任务 action（下单时发两条不同 TTL 的延迟消息）
const (
	ActionPushDrivers = "push_drivers" // 6 分钟：仍待接单则推附近司机
	ActionCancelOrder = "cancel_order" // 10分钟：仍待接单则自动取消
)

// 延迟时间（毫秒，用于 Publish 的 per-message Expiration）
const (
	DelayPushDriversMs int64 = 6 * 60 * 1000  // 6 分钟
	DelayCancelOrderMs int64 = 10 * 60 * 1000 // 10 分钟
)

// OrderDelayMsg 写入延迟队列，TTL 到期后经死信交换机进入消费队列
type OrderDelayMsg struct {
	OrderNo string `json:"order_no"`
	Action  string `json:"action"` // push_drivers | cancel_order
}

// 实时通知 event 类型（WebSocket 推送用，第 9～10 步）
const (
	EventDriverAccepted     = "driver_accepted"
	EventOrderCancelled     = "order_cancelled"
	EventOrderCompleted     = "order_completed"
	EventNewOrderNearby     = "new_order_nearby"
	EventOrderPushedDrivers = "order_pushed_drivers" // 6 分钟延迟推司机后，通知乘客
	EventTripStarted        = "trip_started"         // 司机确认乘客上车，行程开始
)

// OrderNotifyEvent 订单通知事件（JSON 经 MQ 传到 apiGateway → WebSocket）
type OrderNotifyEvent struct {
	Event      string  `json:"event"`
	OrderNo    string  `json:"order_no"`
	UserId     int64   `json:"user_id,omitempty"`   // 乘客 ID
	DriverId   int64   `json:"driver_id,omitempty"` // 司机 ID
	DriverName string  `json:"driver_name,omitempty"`
	CarNumber  string  `json:"car_number,omitempty"`
	CarType    string  `json:"car_type,omitempty"`
	Rating     float64 `json:"rating,omitempty"`
	AcceptAt          int64   `json:"accept_at,omitempty"`
	Msg               string  `json:"msg,omitempty"`                  // 展示文案
	PushRadiusKm      float64 `json:"push_radius_km,omitempty"`      // 推司机搜索半径（公里，仅乘客）
	PushedDriverCount int     `json:"pushed_driver_count,omitempty"` // 通知司机人数（仅乘客）
	StartAddress      string  `json:"start_address,omitempty"`     // 起点（司机 nearby 用）
	Price             float64 `json:"price,omitempty"`             // 预估价（司机 nearby 用）
	Distance          float64 `json:"distance,omitempty"`          // 距离公里（司机 nearby 用）
}
