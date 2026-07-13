// Package model 【第1步·地基】订单表结构与状态机。
//
// 在本项目中的作用：MySQL 是订单的「最终真相」。
// 状态流转：1待接单 → 2已接单 → 5用户已上车 → 3已完成 / 4已取消（超时或用户取消）。
// 下方带 WHERE status=? 的更新方法用于乐观锁，防止重复抢单、重复取消。
package model

import (
	"common/pkg"
	"errors"
	"time"

	"gorm.io/gorm"
)

// 订单状态常量（与表字段 status 注释一致）
const (
	OrderStatusWaiting   = 1 // 待接单
	OrderStatusAccepted  = 2 // 已接单
	OrderStatusCompleted = 3 // 已完成
	OrderStatusCancelled = 4 // 已取消
	OrderStatusOnBoard   = 5 // 用户已上车（行程中）
)

// OngoingOrderStatuses 司机进行中的订单状态（禁止下线、抢单占用等场景复用）
var OngoingOrderStatuses = []int{OrderStatusAccepted, OrderStatusOnBoard}

// UserOngoingOrderStatuses 乘客进行中的订单（待接单 + 已接单 + 行程中）
var UserOngoingOrderStatuses = []int{OrderStatusWaiting, OrderStatusAccepted, OrderStatusOnBoard}

type Order struct {
	gorm.Model
	OrderNo      string     `gorm:"type:varchar(32);comment:订单号;uniqueIndex:idx_order_no" json:"order_no"`
	UserId       int64      `gorm:"type:bigint;comment:用户ID;index:idx_user_id;index:idx_user_status,priority:1" json:"user_id"`
	DriverId     int64      `gorm:"type:bigint;default:0;comment:司机ID;index:idx_driver_id;index:idx_driver_status,priority:1" json:"driver_id"`
	StartLng     float64    `gorm:"type:decimal(15,12);comment:起点经度" json:"start_lng"`
	StartLat     float64    `gorm:"type:decimal(15,12);comment:起点纬度" json:"start_lat"`
	StartAddress string     `gorm:"type:varchar(255);comment:起点地址" json:"start_address"`
	EndLng       float64    `gorm:"type:decimal(15,12);comment:终点经度" json:"end_lng"`
	EndLat       float64    `gorm:"type:decimal(15,12);comment:终点纬度" json:"end_lat"`
	EndAddress   string     `gorm:"type:varchar(255);comment:终点地址" json:"end_address"`
	Distance     float64    `gorm:"type:decimal(10,2);default:0;comment:距离(公里)" json:"distance"`
	Duration     int        `gorm:"type:int;default:0;comment:预计时长(分钟)" json:"duration"`
	Price        float64    `gorm:"type:decimal(10,2);default:0;comment:价格" json:"price"`
	PayType      int        `gorm:"type:int;default:1;comment:1余额 2支付宝 3微信" json:"pay_type"`
	Status       int        `gorm:"type:int;default:1;comment:1待接单 2已接单 3已完成 4已取消 5用户已上车;index:idx_status;index:idx_user_status,priority:2;index:idx_driver_status,priority:2;index:idx_status_created,priority:1" json:"status"`
	FinanceId    int        `gorm:"type:int(11);comment:'优惠券id'" json:"finance_id"`
	CancelReason string     `gorm:"type:varchar(255);comment:取消原因" json:"cancel_reason"`
	AcceptTime   *time.Time `gorm:"type:datetime;comment:接单时间" json:"accept_time"`
	StartTime    *time.Time `gorm:"type:datetime;comment:开始行程时间(用户上车)" json:"start_time"`
}

// OrderFindOngoingByDriver 查询司机当前进行中订单（已接单或行程中）
func (o *Order) OrderFindOngoingByDriver(db *gorm.DB, driverId int64) error {
	return db.Debug().
		Where("driver_id = ? AND status IN ?", driverId, OngoingOrderStatuses).
		Order("id DESC").
		First(o).Error
}

// OrderFindOngoingByUser 查询乘客当前进行中订单（待接单或已接单，取最近一单）
func (o *Order) OrderFindOngoingByUser(db *gorm.DB, uid int64) error {
	return db.Debug().
		Where("user_id = ? AND status IN ?", uid, UserOngoingOrderStatuses).
		Order("id DESC").
		First(o).Error
}

func (o *Order) OrderHasOngoing(db *gorm.DB, driverId int64, statuses []int) (bool, error) {
	var count int64
	err := db.Debug().Model(o).
		Where("driver_id = ? AND status IN (?)", driverId, statuses).
		Count(&count).Error
	return count > 0, err
}

// OrderAdd 创建订单
func (o *Order) OrderAdd(db *gorm.DB) error {
	return db.Debug().Create(o).Error
}

// OrderModelFindNumber 按订单号查询
func (o *Order) OrderModelFindNumber(db *gorm.DB, orderNo string) error {
	return db.Debug().Where("order_no = ?", orderNo).First(o).Error
}

// OrderModelUpd 按主键更新（Updates 忽略零值字段）
func (o *Order) OrderModelUpd(db *gorm.DB) error {
	return db.Debug().Updates(o).Error
}

// OrderUpdateGrabbed 抢单落库：仅 status=待接单 时更新，乐观锁防重复消费
func (o *Order) OrderUpdateGrabbed(db *gorm.DB, orderNo string, driverId int64) error {
	now := time.Now()
	res := db.Debug().Model(&Order{}).
		Where("order_no = ? AND status = ?", orderNo, OrderStatusWaiting).
		Updates(map[string]interface{}{
			"driver_id":   driverId,
			"status":      OrderStatusAccepted,
			"accept_time": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("订单状态已变更，无法接单")
	}
	return nil
}

// OrderUpdateStarted 司机确认乘客上车：仅 status=已接单 时更新为 用户已上车
func (o *Order) OrderUpdateStarted(db *gorm.DB, orderNo string) error {
	now := time.Now()
	res := db.Debug().Model(&Order{}).
		Where("order_no = ? AND status = ?", orderNo, OrderStatusAccepted).
		Updates(map[string]interface{}{
			"status":     OrderStatusOnBoard,
			"start_time": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("订单不存在或状态不允许开始行程")
	}
	return nil
}

// OrderStatusName 订单状态中文名
func OrderStatusName(status int) string {
	switch status {
	case OrderStatusWaiting:
		return "待接单"
	case OrderStatusAccepted:
		return "已接单"
	case OrderStatusOnBoard:
		return "用户已上车"
	case OrderStatusCompleted:
		return "已完成"
	case OrderStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

// OrderListByUser 按乘客 user_id 分页查订单（id 倒序，最新在前）
func (o *Order) OrderListByUser(db *gorm.DB, userId int64, orderNo string, status, page, pageSize int) ([]Order, int64, error) {
	var list []Order
	var total int64

	scope := func(tx *gorm.DB) *gorm.DB {
		q := tx.Debug().Model(o).Where("user_id = ?", userId)
		if orderNo != "" {
			q = q.Where("order_no = ?", orderNo)
		}
		if status > 0 {
			q = q.Where("status = ?", status)
		}
		return q
	}

	if err := scope(db).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := scope(db).Order("id DESC").Scopes(pkg.Paginate(page, pageSize)).Find(&list).Error
	return list, total, err
}

// OrderListByDriver 按司机 driver_id 分页查订单（仅 driver_id>0 的订单，id 倒序）
func (o *Order) OrderListByDriver(db *gorm.DB, driverId int64, orderNo string, status, page, pageSize int) ([]Order, int64, error) {
	var list []Order
	var total int64

	scope := func(tx *gorm.DB) *gorm.DB {
		q := tx.Debug().Model(o).Where("driver_id = ?", driverId)
		if orderNo != "" {
			q = q.Where("order_no = ?", orderNo)
		}
		if status > 0 {
			q = q.Where("status = ?", status)
		}
		return q
	}

	if err := scope(db).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := scope(db).Order("id DESC").Scopes(pkg.Paginate(page, pageSize)).Find(&list).Error
	return list, total, err
}

// OrderUpdateStatus 按订单号更新状态（可附带取消原因等字段）
func (o *Order) OrderUpdateStatus(db *gorm.DB, orderNo string, fromStatus, toStatus int, extra map[string]interface{}) error {
	updates := map[string]interface{}{
		"status": toStatus,
	}
	for k, v := range extra {
		updates[k] = v
	}
	res := db.Debug().Model(&Order{}).
		Where("order_no = ? AND status = ?", orderNo, fromStatus).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("订单不存在或状态不允许变更")
	}
	return nil
}
