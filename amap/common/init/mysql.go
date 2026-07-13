package init

import (
	"common/config"
	"common/model"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var once sync.Once

func MysqlInit() {
	data := config.DataConfig.Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		data.User,
		data.Password,
		data.Host,
		data.Port,
		data.Database,
	)
	var err error
	once.Do(func() {
		if config.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
			panic(err)
		}
		fmt.Println("数据库链接成功")
	})
	if err = config.DB.AutoMigrate(
		&model.User{},
		&model.Driver{},
		&model.Order{},
		&model.Coupon{},
		&model.OrderRating{},
		&model.WalletLog{},
	); err != nil {
		panic(err)
	}
	fmt.Println("数据表创建成功")
	sqlDB, err := config.DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
}
