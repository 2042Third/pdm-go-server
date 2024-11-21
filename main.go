package main

import (
	"log"
	"pdm-go-server/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentSecurityPolicy: "default-src 'self'; frame-ancestors 'none';",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	// Routes
	e.POST("/login", handlers.Login)

	// Start server
	log.Fatal(e.Start(":8080"))
}
