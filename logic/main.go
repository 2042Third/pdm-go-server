package main

import (
	"github.com/joho/godotenv"
	"log"
	"pdm-go-server/internal/db"
	"pdm-go-server/internal/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Initialize database
	database := db.NewDatabase()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(database.DB)

	// Initialize Echo server
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
	e.POST("/login", userHandler.Login)

	// Start server
	log.Fatal(e.Start(":8080"))
}
