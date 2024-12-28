package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"pdm-logic-server/pkg/models"
	"time"
)

func GenerateVerificationCode() string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code) // Ensures 6 digits with leading zeros
}

func RegisterUser(S *Storage, ctx context.Context, name, email, password string) (uint, error) {
	// Check if user already exists
	var user models.User
	err := S.DB.Where("email = ?", email).First(&user).Error
	if err == nil {
		return 0, fmt.Errorf("user with email %s already exists", email)
	}

	// Create new user
	user = models.User{
		Name:        name,
		Email:       email,
		Spw:         password,
		Creation:    time.Now().Format("2006-01-02 15:04:05"),
		Product:     "Web",
		RegisterKey: GenerateVerificationCode(),
	}

	err = S.DB.Create(&user).Error
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

func ValidateUser(S *Storage, ctx context.Context, email, password string) (uint, bool) {
	var user models.User
	err := S.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return 1, false
	}

	// If validation successful, cache the UserInfo
	if password == user.Spw {
		userInfo := models.UserInfo{
			ID:       user.ID,
			Name:     user.Name,
			Creation: user.Creation,
			Product:  user.Product,
			Email:    user.Email,
		}

		// Serialize to JSON before caching
		jsonData, err := json.Marshal(userInfo)
		if err != nil {
			log.Printf("Failed to marshal userInfo: %v", err)
		} else {
			key := fmt.Sprintf("user:%d:userinfo", user.ID)
			err = S.Ch.Set(ctx, key, string(jsonData), 24*time.Hour)
			if err != nil {
				log.Printf("Failed to cache userInfo: %v", err)
			}
		}
	}

	return user.ID, password == user.Spw
}

func GetUserInfo(S *Storage, ctx context.Context, userID uint) (*models.UserInfo, error) {
	var userInfo models.UserInfo

	// Try to get from cache first
	key := fmt.Sprintf("user:%d:userinfo", userID)
	jsonData, err := S.Ch.Get(ctx, key)
	if err == nil {
		// Cache hit - need to deserialize
		err = json.Unmarshal([]byte(jsonData), &userInfo)
		if err == nil {
			return &userInfo, nil
		}
		// If unmarshal fails, log and continue to DB
		log.Printf("Failed to unmarshal cached userInfo: %v", err)
	}

	// Cache miss or unmarshal error, get from DB
	err = S.DB.Model(&models.User{}).
		Select("id", "name", "creation", "product", "email").
		Where("id = ?", userID).
		First(&userInfo).Error
	if err != nil {
		return nil, err
	}

	// Cache the result for next time
	bytes, err := json.Marshal(userInfo)
	jsonData = string(bytes)
	if err == nil {
		err = S.Ch.Set(ctx, key, jsonData, 24*time.Hour)
		if err != nil {
			log.Printf("Failed to cache userInfo: %v", err)
		}
	}

	return &userInfo, nil
}

func GetUserByID(S *Storage, userID uint) (models.User, error) {
	var usr models.User
	err := S.DB.First(&usr, userID).Error
	return usr, err
}
