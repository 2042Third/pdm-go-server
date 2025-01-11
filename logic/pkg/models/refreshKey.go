package models

import (
	"time"
)

type RefreshKey struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	RefreshKey     string    `gorm:"column:refresh_key;not null" json:"refreshKey"`
	UserID         string    `gorm:"column:userid;type:uuid;not null" json:"userid"`
	ExpirationTime time.Time `gorm:"column:expiration_time;type:timestamp" json:"expirationTime"`
	CreationTime   time.Time `gorm:"column:creation_time;type:timestamp with time zone;default:current_timestamp" json:"creationTime"`
	UsageCount     int       `gorm:"column:usage_count;not null;default:0" json:"usageCount"`
}

// TableName overrides the default table name for GORM
func (RefreshKey) TableName() string {
	return "refresh_key"
}
