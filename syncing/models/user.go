package models

import (
	_ "gorm.io/driver/postgres"
	_ "gorm.io/gorm"
)

// User represents the userinfo table in the database
type User struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	Spw         string `gorm:"column:spw" json:"spw"`
	Creation    string `gorm:"column:creation" json:"creation"`
	Product     string `gorm:"column:product" json:"product"`
	Email       string `gorm:"column:email" json:"email"`
	RegisterKey string `gorm:"column:register_key" json:"register_key"`
	Logs        string `gorm:"column:logs" json:"logs"`
	Registered  string `gorm:"column:registered" json:"registered"`
}

func (User) TableName() string {
	return "userinfo"
}
