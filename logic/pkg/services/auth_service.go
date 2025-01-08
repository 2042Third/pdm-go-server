package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"pdm-logic-server/pkg/models"
	"strconv"
	"time"
)

func GenerateVerificationCode() string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code) // Ensures 6 digits with leading zeros
}

func MakeNewVerificationCode(S *Storage, ctx context.Context, email string) (string, error) {
	var user models.User
	err := S.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return "", fmt.Errorf("user with email %s not found", email)
	}

	// Generate new verification code
	code := GenerateVerificationCode()
	user.RegisterKey = code
	err = S.DB.Save(&user).Error
	if err != nil {
		return "", err
	}

	return code, nil
}

func RegisterUser(S *Storage, ctx context.Context, name, email, password string) (models.SignupInternalResponse, error) {
	// Check if user already exists
	var user models.User
	err := S.DB.Where("email = ?", email).First(&user).Error
	if err == nil {
		return models.SignupInternalResponse{
			UserId: 0,
		}, fmt.Errorf("user with email %s already exists", email)
	}

	// Create new user
	user = models.User{
		Name:        name,
		Email:       email,
		Spw:         password,
		Creation:    strconv.FormatInt(time.Now().UnixMilli(), 10), // Use unix timestamp
		Product:     "pdm web 2",
		RegisterKey: GenerateVerificationCode(),
		Registered:  "0",
	}

	err = S.DB.Create(&user).Error
	if err != nil {
		return models.SignupInternalResponse{UserId: 0}, err
	}

	return models.SignupInternalResponse{
		UserId:           user.ID,
		VerificationCode: user.RegisterKey,
	}, nil
}

func ValidateVerificationCode(S *Storage, userEmail, code string) bool {
	var user models.User
	err := S.DB.Where("email = ? AND register_key = ?", userEmail, code).First(&user).Error
	if err != nil {
		return false
	}

	// Update user to be registered
	user.Registered = "1"
	err = S.DB.Save(&user).Error
	if err != nil {
		log.Printf("Failed to update user registration status: %v", err)
		return false
	}

	return true
}

func ValidateUser(S *Storage, ctx context.Context, email, password string) (uint64, bool) {
	var user models.User
	err := S.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return 1, false
	}

	// If validation successful, cache the UserInfo
	if password == user.Spw {
		userInfo := models.UserInfo{
			ID:         user.ID,
			Name:       user.Name,
			Creation:   user.Creation,
			Product:    user.Product,
			Email:      user.Email,
			Registered: user.Registered,
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

func GetUserInfo(S *Storage, ctx context.Context, userID uint64) (*models.UserInfo, error) {
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
		Select("id", "name", "creation", "product", "email", "registered").
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

func GetUserByID(S *Storage, userID uint64) (models.User, error) {
	var usr models.User
	err := S.DB.First(&usr, userID).Error
	return usr, err
}
