package models

import (
	"time"
)

type Notes struct {
	NoteID     string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4();column:noteid" json:"noteid"`
	UserID     string    `gorm:"column:userid;type:uuid;not null" json:"userid"`
	Content    string    `gorm:"column:content" json:"content"`
	H          string    `gorm:"column:h" json:"h"`
	Intgrh     string    `gorm:"column:intgrh" json:"intgrh"`
	Time       time.Time `gorm:"column:time;type:timestamptz;default:CURRENT_TIMESTAMP" json:"time"`
	UpdateTime time.Time `gorm:"column:update_time;type:timestamptz;default:CURRENT_TIMESTAMP" json:"update_time"`
	Heading    string    `gorm:"column:heading" json:"heading"`
	Deleted    int       `gorm:"column:deleted;not null;default:0" json:"deleted"`
}

// TableName overrides the default table name for GORM
func (Notes) TableName() string {
	return "notes"
}
