package middleware

import (
	"crypto/ed25519"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"net/http"
)

// JWTMiddlewareConfig holds the configuration for the JWT middleware
type JWTMiddlewareConfig struct {
	PublicKey ed25519.PublicKey
}

// CreateJWTMiddleware creates a new JWT middleware with the provided public key
func CreateJWTMiddleware(publicKey ed25519.PublicKey) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from header
			tokenString := c.Request().Header.Get("Authorization")
			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Remove 'Bearer ' prefix if present
			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
			}

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate the signing method
				if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return publicKey, nil
			})

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("invalid token: %v", err))
			}

			if !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Safely type assert and access claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			// Safely get email from claims
			email, ok := claims["email"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid email claim")
			}

			userId, ok := claims["userId"]
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid userId claim")
			}

			// Set claims in context
			c.Set("email", email)
			c.Set("userId", userId)
			c.Set("token", token)

			return next(c)
		}
	}
}
