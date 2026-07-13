package init

import (
	"common/config"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func RedisInit() {
	data := config.DataConfig.Redis
	Addr := fmt.Sprintf("%s:%d", data.Host, data.Port)
	config.Rdb = redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: data.Password, // no password set
		DB:       data.Database, // use default DB
	})
	if err := config.Rdb.Ping(config.Ctx).Err(); err != nil {
		panic(err)
	}
	fmt.Println("redis成功")
}
