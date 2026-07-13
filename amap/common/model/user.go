package model

import "gorm.io/gorm"

// CompanyUserID 公司财务账户（users 表 id=999）
const CompanyUserID int64 = 999

// 钱包流水用户类型
const (
	WalletUserTypePassenger = 1
	WalletUserTypeDriver    = 2
	WalletUserTypeCompany   = 3
)

type User struct {
	gorm.Model
	Phone       string  `gorm:"type:varchar(11);comment:手机号;uniqueIndex:idx_phone" json:"phone"`
	Password    string  `gorm:"type:varchar(255);comment:密码(MD5)" json:"password"`
	Nickname    string  `gorm:"type:varchar(50);default:乘客;comment:昵称" json:"nickname"`
	Name        string  `gorm:"type:varchar(50);comment:真实姓名" json:"name"`
	IdCard      string  `gorm:"type:varchar(18);comment:身份证号" json:"id_card"`
	Avatar      string  `gorm:"type:varchar(255);comment:头像URL" json:"avatar"`
	Gender      int     `gorm:"type:tinyint;default:0;comment:性别 0未知 1男 2女" json:"gender"`
	Balance     float64 `gorm:"type:decimal(10,2);default:0.00;comment:余额" json:"balance"`
	CreditScore int     `gorm:"type:int;default:100;comment:信用分" json:"credit_score"`
	Email       string  `gorm:"type:varchar(100);comment:'qq邮箱地址'" json:"email"`
	Status      int     `gorm:"type:tinyint;default:1;comment:状态 1正常 2未实名 3禁用 ;index:idx_status" json:"status"`
	CurrentLng  float64 `gorm:"type:decimal(15,12);default:0;comment:当前经度" json:"current_lng"`
	CurrentLat  float64 `gorm:"type:decimal(15,12);default:0;comment:当前纬度" json:"current_lat"`
}

func (u *User) UserModelFindPhone(db *gorm.DB, phone string) error {
	return db.Debug().Where("phone=?", phone).First(u).Error
}

func (u *User) UserDataRegister(db *gorm.DB) error {
	return db.Debug().Omit("Nickname").Create(u).Error
}

func (u *User) UserModelFindId(db *gorm.DB, id int64) error {
	return db.Debug().Where("id=?", id).First(u).Error
}

func (u *User) UserModelUpd(db *gorm.DB) error {
	return db.Debug().Updates(u).Error
}

// UserModelListCouponTargets 可接收优惠券的乘客（排除公司账户、禁用用户）
func (u *User) UserModelListCouponTargets(db *gorm.DB) ([]User, error) {
	var list []User
	err := db.Where("id <> ? AND status = ?", CompanyUserID, 1).Find(&list).Error
	return list, err
}
