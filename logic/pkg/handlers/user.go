package handlers

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"pdm-logic-server/pkg/errors"
	"pdm-logic-server/pkg/services"
	"strconv"
	"time"
)

type UserHandler struct {
	*BaseHandler
}

func NewUserHandler(base *BaseHandler) *UserHandler {
	return &UserHandler{BaseHandler: base}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *UserHandler) Login(c echo.Context) error {
	ctx := context.Background()

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request format", err)
	}

	if err := c.Validate(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request data", err)
	}

	userId, isValid := services.ValidateUser(h.storage, ctx, req.Email, req.Password)
	if !isValid {
		return errors.NewAppError(http.StatusUnauthorized, "Invalid credentials", nil)
	}

	token, expiration, err := h.authService.GenerateToken(req.Email, userId)
	if err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to generate token", err)
	}

	h.storage.R.DispatchAddSession(strconv.Itoa(int(userId)), token, expiration)
	if err := h.cacheUserSession(ctx, req.Email, userId, token, time.Unix(expiration, 0)); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sessionKey": token,
		"expiration": expiration,
		"message":    "Login successful",
	})
}

func (h *UserHandler) cacheUserSession(ctx context.Context, email string, userId uint, token string, expiration time.Time) error {
	// Cache user ID mapping
	if err := h.storage.Ch.HSet(ctx, "userEmail:userId", email, strconv.FormatUint(uint64(userId), 10)); err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to cache user mapping", err)
	}

	// Cache session token
	key := fmt.Sprintf("user:%d:sessionKey", userId)
	ttl := time.Until(expiration)
	if err := h.storage.Ch.Set(ctx, key, token, ttl); err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to cache session", err)
	}

	return nil
}

func (h *UserHandler) Logout(c echo.Context) error {
	ctx := context.Background()
	log.Println("Logout request received")

	userEmail := c.Get("email").(string)

	userId, err := h.storage.Ch.HGet(ctx, "userEmail:userId", userEmail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HGet userEmail:userId",
		})
	}
	key := fmt.Sprintf("user:%s:sessionKey", userId)

	sessionKey, err := h.storage.Ch.Get(ctx, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, Get user:%s:sessionKeyy",
		})
	}

	token, err := h.authService.ValidateToken(sessionKey)
	if err != nil {
		log.Println("Token invalid at logout: ", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	parsedExp := time.Unix(int64(claims["exp"].(float64)), 0)
	ttl := time.Until(parsedExp) // Calculate duration until expiration (for Redis TTL)

	// Get blocked list count
	maxCount, err := h.storage.Ch.CountKeys(ctx, fmt.Sprintf("user:%s:blocked:*", userId))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, CountKeys user:%s:blocked:*",
		})
	}

	// Add session key to blocked list
	blockListKey := fmt.Sprintf("user:%s:blocked:%d", userId, maxCount+1)
	err = h.storage.Ch.Set(ctx, blockListKey, sessionKey, ttl)

	err = h.storage.Ch.Delete(ctx, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, Delete user:%s:sessionKey",
		})
	}

	// Keep the session key in the database
	//h.storage.R.DispatchDeleteSession(userId, sessionKey)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

func (h *UserHandler) GetUserInfo(c echo.Context) error {
	ctx := context.Background()
	log.Println("Get user info request received")

	userEmail := c.Get("email").(string)

	userId, err := h.storage.Ch.HGet(ctx, "userEmail:userId", userEmail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HGet userEmail:userId",
		})
	}

	intUserId, err := strconv.Atoi(userId)
	userInfo, err := services.GetUserInfo(h.storage, ctx, uint(intUserId))

	return c.JSON(http.StatusOK, userInfo)
}
