package models

import (
	"time"
)

type RefreshKey struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RefreshKey     string    `gorm:"column:refresh_key;not null" json:"refreshKey"`
	UserID         uint      `gorm:"column:userid;not null" json:"userid"`
	ExpirationTime time.Time `gorm:"column:expiration_time;type:timestamp" json:"expirationTime"`
	UsageCount     int       `gorm:"column:usage_count;not null;default:0" json:"usageCount"`
}

// TableName overrides the default table name for GORM
func (RefreshKey) TableName() string {
	return "refresh_key"
}
