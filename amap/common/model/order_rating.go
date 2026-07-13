package model

import (
	"errors"

	"gorm.io/gorm"
)

// OrderRating 订单评价表
type OrderRating struct {
	gorm.Model
	OrderNo  string `gorm:"type:varchar(50);comment:订单单号;uniqueIndex:idx_rating_order_no" json:"order_no"`
	UserId   int64  `gorm:"type:bigint;comment:用户ID;index:idx_rating_user_id" json:"user_id"`
	DriverId int64  `gorm:"type:bigint;comment:司机ID;index:idx_rating_driver_id" json:"driver_id"`
	Rating   int64  `gorm:"type:int;comment:评分 1-5星" json:"rating"`
	Comment  string `gorm:"type:varchar(500);comment:评价内容" json:"comment"`
	Tags     string `gorm:"type:varchar(255);comment:标签(JSON数组)" json:"tags"`
}

func (r *OrderRating) OrderRatingFindByOrderNo(db *gorm.DB, orderNo string) error {
	return db.Debug().Where("order_no = ?", orderNo).First(r).Error
}

func (r *OrderRating) OrderRatingAdd(db *gorm.DB) error {
	return db.Debug().Create(r).Error
}

func (r *OrderRating) OrderRatingDelete(db *gorm.DB) error {
	return db.Debug().Delete(r).Error
}

// OrderRatingDriverAvg 统计司机历史评价均分（不含软删）
func OrderRatingDriverAvg(db *gorm.DB, driverId int64) (avg float64, count int64, err error) {
	err = db.Debug().Model(&OrderRating{}).
		Where("driver_id = ?", driverId).
		Select("COALESCE(AVG(rating), 0) as avg, COUNT(*) as count").
		Row().Scan(&avg, &count)
	return avg, count, err
}

// OrderRatingExistsByOrderNo 是否已评价
func OrderRatingExistsByOrderNo(db *gorm.DB, orderNo string) (bool, error) {
	var n int64
	err := db.Debug().Model(&OrderRating{}).Where("order_no = ?", orderNo).Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// ErrOrderAlreadyRated 重复评价
var ErrOrderAlreadyRated = errors.New("该订单已评价")
