package services

import (
	"pdm-go-server/internal/models"
)

func ValidateUser(S *Storage, email, password string) (uint, bool) {
	var user models.User
	err := S.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return 1, false
	}

	return user.ID, password == user.Spw
}

func GetUserByID(S *Storage, userID uint) (models.User, error) {
	var usr models.User
	err := S.DB.First(&usr, userID).Error
	return usr, err
}
