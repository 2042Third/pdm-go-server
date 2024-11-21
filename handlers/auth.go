package handlers

import (
	"log"
	"net/http"
	"pdm-go-server/services"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c echo.Context) error {

	log.Println("Login request received")
	creds := new(Credentials)
	if err := c.Bind(creds); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}
	log.Println("Login request received for:", creds.Email)

	// Validate user credentials
	if !services.ValidateUser(creds.Email, creds.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid credentials"})
	}

	// Generate JWT token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = creds.Email
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	t, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Token generation failed"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sessionKey": t,
		"expiration": claims["exp"],
		"message":    "Login successful",
	})
}
