package handlers

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

func refresh(c echo.Context) error {
	type RefreshRequest struct {
		RefreshKey string `json:"refreshKey"`
	}

	req := new(RefreshRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	// Validate refresh key (mock logic here)
	if req.RefreshKey != "valid-refresh-key" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid refresh key"})
	}

	// Generate a new session key
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = "user@example.com" // Replace with actual user email
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	t, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Token generation failed"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sessionKey": t,
		"expiration": claims["exp"],
		"message":    "Refresh successful",
	})
}
