package model

import (
	"common/pkg"

	"gorm.io/gorm"
)

// WalletLog 钱包流水表
type WalletLog struct {
	gorm.Model
	UserId        int64   `gorm:"type:bigint;comment:用户ID;index:idx_user_id;index:idx_user_type,priority:1" json:"user_id"`
	UserType      int     `gorm:"type:int;comment:用户类型 1乘客 2司机 3公司;index:idx_user_type,priority:2" json:"user_type"`
	OrderNo       string  `gorm:"type:varchar(50);comment:订单号;index:idx_order_no" json:"order_no"`
	Amount        float64 `gorm:"type:decimal(10,2);comment:金额(正数收入,负数支出)" json:"amount"`
	BalanceBefore float64 `gorm:"type:decimal(10,2);comment:变动前余额" json:"balance_before"`
	BalanceAfter  float64 `gorm:"type:decimal(10,2);comment:变动后余额" json:"balance_after"`
	Type          int     `gorm:"type:int;comment:类型 1充值 2消费 3提现 4退款;index:idx_type" json:"type"`
	Status        int     `gorm:"type:int;comment:订单状态 1已支付 2待支付;index:idx_status" json:"status"`
	Remark        string  `gorm:"type:varchar(255);comment:备注" json:"remark"`
}

func (l *WalletLog) WalletLogModel(db *gorm.DB) error {
	return db.Debug().Create(l).Error
}

func (l *WalletLog) WalletLogModelFindOrderNo(db *gorm.DB, no string) error {
	return db.Debug().Where("order_no=?", no).First(l).Error
}

func (l *WalletLog) WalletLogModelUpd(db *gorm.DB) error {
	return db.Debug().Updates(l).Error
}

// WalletLogListByUser 按 user_id + user_type 分页查流水（默认 id 倒序，最新在前）
func (l *WalletLog) WalletLogListByUser(db *gorm.DB, userId int64, userType int, orderNo string, page, pageSize int) ([]WalletLog, int64, error) {
	var list []WalletLog
	var total int64

	scope := func(tx *gorm.DB) *gorm.DB {
		q := tx.Debug().Model(l).Where("user_id = ? AND user_type = ?", userId, userType)
		if orderNo != "" {
			q = q.Where("order_no = ?", orderNo)
		}
		return q
	}

	if err := scope(db).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := scope(db).Order("id DESC").Scopes(pkg.Paginate(page, pageSize)).Find(&list).Error
	return list, total, err
}
