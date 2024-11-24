package services

import (
	"context"
	"gorm.io/gorm"

	"pdm-go-server/internal/cache"
)

type Storage struct {
	DB  *gorm.DB
	ch  *cache.Cache
	chc *context.Context
}

func NewStorage(db *gorm.DB, ch *cache.Cache, chc *context.Context) *Storage {
	return &Storage{DB: db, ch: ch, chc: chc}
}
