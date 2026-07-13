package model

import "gorm.io/gorm"

type Driver struct {
	gorm.Model
	Phone        string  `gorm:"type:varchar(11);comment:手机号" json:"phone"`
	Password     string  `gorm:"type:varchar(255);comment:密码" json:"password"`
	Name         string  `gorm:"type:varchar(50);comment:真实姓名" json:"name"`
	IdCard       string  `gorm:"type:varchar(18);comment:身份证号" json:"id_card"`
	Avatar       string  `gorm:"type:varchar(255);comment:头像URL" json:"avatar"`
	CarNumber    string  `gorm:"type:varchar(20);comment:车牌号" json:"car_number"`
	CarType      string  `gorm:"type:varchar(50);comment:车型" json:"car_type"`
	CarColor     string  `gorm:"type:varchar(20);comment:车辆颜色" json:"car_color"`
	LicensePhoto string  `gorm:"type:varchar(255);comment:驾驶证照片URL" json:"license_photo"`
	VehiclePhoto string  `gorm:"type:varchar(255);comment:行驶证照片URL" json:"vehicle_photo"`
	Balance      float64 `gorm:"type:decimal(10,2);default:0.00;comment:余额" json:"balance"`
	Rating       float64 `gorm:"type:decimal(3,2);default:5.00;comment:评分" json:"rating"`
	OrderCount   int     `gorm:"type:int;default:0;comment:接单数" json:"order_count"`
	Status       int     `gorm:"type:int;default:0;comment:1正常 2未实名 3禁用" json:"status"`
	OnlineStatus int     `gorm:"type:int;default:0;comment:1在线 2离线 " json:"online_status"`
	Email        string  `gorm:"type:varchar(100);comment:'qq邮箱地址'" json:"email"`
	CurrentLng   float64 `gorm:"type:decimal(15,12);default:0;comment:当前经度" json:"current_lng"`
	CurrentLat   float64 `gorm:"type:decimal(15,12);default:0;comment:当前纬度" json:"current_lat"`
}

func (d *Driver) DriverModelFindPhone(db *gorm.DB, phone string) error {
	return db.Debug().Where("phone=?", phone).First(d).Error
}

func (d *Driver) DriverDataRegister(db *gorm.DB) error {
	return db.Debug().Create(d).Error
}

func (d *Driver) DriverModelFindId(db *gorm.DB, id int64) error {
	return db.Debug().Where("id=?", id).First(d).Error
}

func (d *Driver) DriverModelUpd(db *gorm.DB) error {
	return db.Debug().Updates(d).Error
}

// DriverListOnline 查询已上线且已上报经纬度的司机
func (d *Driver) DriverListOnline(db *gorm.DB) ([]Driver, error) {
	var list []Driver
	err := db.Debug().
		Where("online_status = ? AND current_lng != 0 AND current_lat != 0", 1).
		Find(&list).Error
	return list, err
}
