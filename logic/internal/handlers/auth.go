package handlers

import (
	"log"
	"net/http"
	"pdm-go-server/internal/auth"
	"pdm-go-server/internal/services"

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
	log.Println("Login request received")
	creds := new(Credentials)
	if err := c.Bind(creds); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	log.Println("Login attempt for:", creds.Email)

	// Validate user credentials
	if !services.ValidateUser(h.S, creds.Email, creds.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid credentials"})
	}

	// Generate JWT token
	token, expiration, err := h.AuthService.GenerateToken(creds.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Token generation failed"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sessionKey": token,
		"expiration": expiration,
		"message":    "Login successful",
	})
}
