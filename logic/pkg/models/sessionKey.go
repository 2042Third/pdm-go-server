package models

import (
	"time"
)

type SessionKey struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	SessionKey     string    `gorm:"column:session_key;not null" json:"sessionKey"`
	UserID         string    `gorm:"column:userid;type:uuid;not null" json:"userid"`
	ExpirationTime time.Time `gorm:"column:expiration_time;type:timestamp" json:"expirationTime"`
	CreationTime   time.Time `gorm:"column:creation_time;type:timestamp with time zone;default:current_timestamp" json:"creationTime"`
	Valid          string    `gorm:"column:valid;type:varchar(1);default:'0'" json:"valid"`
}

// TableName overrides the default table name for GORM
func (SessionKey) TableName() string {
	return "session_key"
}
