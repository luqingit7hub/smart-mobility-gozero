package model

import (
	"time"

	"gorm.io/gorm"
)

// 优惠券表
type Coupon struct {
	gorm.Model
	Uid       int       `gorm:"type:int(11);comment:'用户id';index:idx_uid;index:idx_uid_type,priority:1" json:"uid"`
	Type      int       `gorm:"type:int(11);comment:'优惠券类型(1:现金券,2:折扣券,3:免费乘车券)';index:idx_type;index:idx_uid_type,priority:2" json:"type"`
	QuanMoney float64   `gorm:"type:decimal(10,2);comment:'优惠券金额'" json:"quan_money"`
	Discount  float64   `gorm:"type:decimal(10,2);comment:'优惠券折扣'" json:"discount"`
	CityCode  string    `gorm:"type:varchar(200);comment:'适用城市编号'" json:"city_code"`
	OutTime   time.Time `gorm:"type:datetime;comment:'优惠券过期时间';index:idx_out_time" json:"out_time"`
}

// CouponFindId 按优惠券主键查询
func (c *Coupon) CouponFindId(db *gorm.DB, tid int64) error {
	return db.Debug().Where("id = ?", tid).First(c).Error
}

// CouponDel 删除已使用的优惠券
func (c *Coupon) CouponDel(db *gorm.DB, tid int64) error {
	return db.Debug().Where("id = ?", tid).Delete(c).Error
}

// CouponCreate 发放优惠券
func (c *Coupon) CouponCreate(db *gorm.DB) error {
	return db.Create(c).Error
}

// CouponListByUserID 按用户 id 查询其全部未软删优惠券（含已过期）
func (c *Coupon) CouponListByUserID(db *gorm.DB, uid int64) ([]Coupon, error) {
	var list []Coupon
	err := db.Where("uid = ?", uid).Order("id ASC").Find(&list).Error
	return list, err
}
