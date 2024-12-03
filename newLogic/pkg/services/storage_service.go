package services

import (
	"gorm.io/gorm"
	"pdm-logic-server/pkg/cache"
)

type Storage struct {
	DB *gorm.DB
	R  *RabbitMQCtx
	Ch *cache.RedisCache
}

func NewStorage(db *gorm.DB, r *RabbitMQCtx, ch *cache.RedisCache) *Storage {
	return &Storage{DB: db, R: r, Ch: ch}
}
