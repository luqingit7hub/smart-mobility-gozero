package config

import (
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	DataConfig AppConfig
	DB         *gorm.DB
	Rdb        *redis.Client
	Ctx        = context.Background()
	Esc        *elastic.Client
)
