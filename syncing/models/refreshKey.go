package models

import (
	"time"
)

type RefreshKey struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RefreshKey     string    `gorm:"column:refresh_key;not null" json:"refreshKey"`
	UserID         uint64    `gorm:"column:user_id;not null" json:"userId"`
	ExpirationTime time.Time `gorm:"column:expiration_time;type:timestamp" json:"expirationTime"`
	UsageCount     int       `gorm:"column:usage_count;not null;default:0" json:"usageCount"`
}

// TableName overrides the default table name for GORM
func (RefreshKey) TableName() string {
	return "refresh_key"
}
