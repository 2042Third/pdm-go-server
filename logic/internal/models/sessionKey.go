package models

import (
	"time"
)

type SessionKey struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionKey     string    `gorm:"column:session_key;not null" json:"sessionKey"`
	UserID         uint      `gorm:"column:userid;not null" json:"userid"`
	ExpirationTime time.Time `gorm:"column:expiration_time;type:timestamp" json:"expirationTime"`
}

// TableName overrides the default table name for GORM
func (SessionKey) TableName() string {
	return "session_key"
}
