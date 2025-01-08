package models

import (
	"time"
)

type SessionKey struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionKey     string    `gorm:"column:session_key;not null" json:"sessionKey"`
	UserID         uint64    `gorm:"column:user_id;not null" json:"userid"`
	ExpirationTime time.Time `gorm:"column:expiration_time;type:timestamp" json:"expirationTime"`
	CreationTime   time.Time `gorm:"column:creation_time;type:timestamp with time zone;default:current_timestamp" json:"creationTime"`
	Valid          string    `gorm:"column:valid;type:varchar(1);default:'0'" json:"valid"`
}

// TableName overrides the default table name for GORM
func (SessionKey) TableName() string {
	return "session_key"
}
