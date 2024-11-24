package handlers

import (
	"context"
	"log"
	"net/http"
	"pdm-go-server/internal/auth"
	"pdm-go-server/internal/services"
	"strconv"

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
	userId, isValid := services.ValidateUser(h.S, creds.Email, creds.Password)
	if !isValid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid credentials"})
	}

	// Generate JWT token
	token, expiration, err := h.AuthService.GenerateToken(creds.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Token generation failed"})
	}

	// Use the context directly in the handler
	err = h.S.Ch.HSet(ctx, "userEmail:userId", creds.Email, strconv.Itoa(int(userId)))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed",
		})
	}
	err = h.S.Ch.HSet(ctx, "user:sessionKey", strconv.Itoa(int(userId)), token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed",
		})
	}

	h.S.R.DispatchAddSession(strconv.Itoa(int(userId)), token)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sessionKey": token,
		"expiration": expiration,
		"message":    "Login successful",
	})
}

func (h *UserHandler) Logout(c echo.Context) error {
	ctx := context.Background()
	log.Println("Logout request received")

	userEmail := c.Get("userEmail").(string)

	userId, err := h.S.Ch.HGet(ctx, "userEmail:userId", userEmail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HGet userEmail:userId",
		})
	}
	sessionKey, err := h.S.Ch.HGet(ctx, "user:sessionKey", userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HGet userId:sessionKey",
		})
	}

	err = h.S.Ch.HDel(ctx, "user:sessionKey", userId, sessionKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HDel userId:sessionKey",
		})
	}

	h.S.R.DispatchDeleteSession(userId, sessionKey)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}
