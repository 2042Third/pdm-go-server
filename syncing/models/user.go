package models

import (
	"encoding/json"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/gorm"
)

// User represents the userinfo table in the database
type User struct {
	ID          string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
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

type UserInfo struct {
	ID         string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name       string `gorm:"column:name" json:"name"`
	Creation   string `gorm:"column:creation" json:"creation"`
	Product    string `gorm:"column:product" json:"product"`
	Email      string `gorm:"column:email" json:"email"`
	Registered string `gorm:"column:registered" json:"registered"`
}

func (u *UserInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *UserInfo) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}
