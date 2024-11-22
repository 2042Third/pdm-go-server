package services

import (
	"pdm-go-server/internal/models"

	"gorm.io/gorm"
)

func ValidateUser(DB *gorm.DB, email, password string) bool {
	var user models.User
	err := DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return false
	}

	return password == user.Spw
}

func GetUserByID(DB *gorm.DB, userID uint) (models.User, error) {
	var usr models.User
	err := DB.First(&usr, userID).Error
	return usr, err
}
