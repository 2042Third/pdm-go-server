package handlers

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"pdm-go-server/internal/auth"
	"pdm-go-server/internal/services"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserHandler struct {
	S           *services.Storage
	AuthService *auth.AuthService
}

func NewUserHandler(storage *services.Storage, authService *auth.AuthService) *UserHandler {
	return &UserHandler{S: storage, AuthService: authService}
}

func (h *UserHandler) Login(c echo.Context) error {
	ctx := context.Background()
	log.Println("Login request received")
	creds := new(Credentials)
	if err := c.Bind(creds); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	log.Println("Login attempt for:", creds.Email)

	// Validate user credentials
	userId, isValid := services.ValidateUser(h.S, ctx, creds.Email, creds.Password)
	if !isValid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid credentials"})
	}

	// Generate JWT token
	tokenStr, expiration, err := h.AuthService.GenerateToken(creds.Email, userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Token generation failed"})
	}

	fmt.Printf("Expiration: %v\n", time.Unix(expiration, 0))
	// Use the context directly in the handler
	err = h.S.Ch.HSet(ctx, "userEmail:userId", creds.Email, strconv.Itoa(int(userId)))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed",
		})
	}
	token, err := h.AuthService.ValidateToken(tokenStr)
	if err != nil {
		log.Println("Token invalid at login (status 500): ", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Server Error",
		})
	}

	claims := token.Claims.(jwt.MapClaims)
	parsedExp := time.Unix(int64(claims["exp"].(float64)), 0)
	ttl := time.Until(parsedExp) // Calculate duration until expiration (for Redis TTL)

	key := fmt.Sprintf("user:%s:sessionKey", strconv.Itoa(int(userId)))
	err = h.S.Ch.Set(ctx, key, tokenStr, ttl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed",
		})
	}

	h.S.R.DispatchAddSession(strconv.Itoa(int(userId)), tokenStr, expiration)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sessionKey": tokenStr,
		"expiration": expiration,
		"message":    "Login successful",
	})
}

func (h *UserHandler) Logout(c echo.Context) error {
	ctx := context.Background()
	log.Println("Logout request received")

	userEmail := c.Get("email").(string)

	userId, err := h.S.Ch.HGet(ctx, "userEmail:userId", userEmail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HGet userEmail:userId",
		})
	}
	key := fmt.Sprintf("user:%s:sessionKey", userId)

	sessionKey, err := h.S.Ch.Get(ctx, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, Get user:%s:sessionKeyy",
		})
	}

	token, err := h.AuthService.ValidateToken(sessionKey)
	if err != nil {
		log.Println("Token invalid at logout: ", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	parsedExp := time.Unix(int64(claims["exp"].(float64)), 0)
	ttl := time.Until(parsedExp) // Calculate duration until expiration (for Redis TTL)

	// Get blocked list count
	maxCount, err := h.S.Ch.CountKeys(ctx, fmt.Sprintf("user:%s:blocked:*", userId))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, CountKeys user:%s:blocked:*",
		})
	}

	// Add session key to blocked list
	blockListKey := fmt.Sprintf("user:%s:blocked:%d", userId, maxCount+1)
	err = h.S.Ch.Set(ctx, blockListKey, sessionKey, ttl)

	err = h.S.Ch.Delete(ctx, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, Delete user:%s:sessionKey",
		})
	}

	h.S.R.DispatchDeleteSession(userId, sessionKey)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

func (h *UserHandler) GetUserInfo(c echo.Context) error {
	ctx := context.Background()
	log.Println("Get user info request received")

	userEmail := c.Get("email").(string)

	userId, err := h.S.Ch.HGet(ctx, "userEmail:userId", userEmail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HGet userEmail:userId",
		})
	}

	intUserId, err := strconv.Atoi(userId)
	userInfo, err := services.GetUserInfo(h.S, ctx, uint(intUserId))

	return c.JSON(http.StatusOK, userInfo)
}
