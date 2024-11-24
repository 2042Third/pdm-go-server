package authMiddleware

import (
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"pdm-go-server/internal/auth"
)

func JWTMiddleware(authService *auth.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Missing or invalid token"})
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := authService.ValidateToken(tokenStr)
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
			}

			claims := token.Claims.(jwt.MapClaims)
			c.Set("user", claims["email"])
			return next(c)
		}
	}
}
