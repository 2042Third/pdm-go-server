package auth

import (
	"fmt"
	"pdm-go-server/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AuthService struct {
	SecretKey string
}

func NewAuthService(secretKey string) *AuthService {
	if secretKey == "" {
		panic("JWT secret key is not set")
	}
	return &AuthService{SecretKey: secretKey}
}

// GenerateToken generates a new JWT token with provided claims.
func (a *AuthService) GenerateToken(email string) (string, int64, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	expiration := time.Now().Add(config.TokenValidityDuration).Unix()

	claims["email"] = email
	claims["exp"] = expiration

	t, err := token.SignedString([]byte(a.SecretKey))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return t, expiration, nil
}

// ValidateToken validates the provided JWT token.
func (a *AuthService) ValidateToken(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.SecretKey), nil
	})
}
