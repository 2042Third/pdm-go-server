package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"pdm-go-server/internal/auth"
	"pdm-go-server/internal/authMiddleware"
	"pdm-go-server/internal/cache"
	"pdm-go-server/internal/db"
	"pdm-go-server/internal/handlers"
	"pdm-go-server/internal/services"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Load environment variables
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY is not set")
	}

	// Initialize database
	database := db.NewDatabase()

	// Initialize cache
	cfg := &cache.CacheConfig{
		Address:  "localhost:6379",
		Password: "",
		DB:       0,
	}

	// Initialize Redis client
	redisClient := cache.NewRedisClient(cfg)
	cacheLayer := cache.NewCache(redisClient)

	// Initialize RabbitMQ connection
	rabbitMQCtx, _ := services.NewRabbitMQHandler()

	// Initialize storage
	storage := services.NewStorage(database.DB, rabbitMQCtx, cacheLayer)

	// Initialize auth service
	authService := auth.NewAuthService(secretKey)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(storage, authService)

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
	e.GET("/api/user/logout", userHandler.Logout, authMiddleware.JWTMiddleware(authService))
	e.GET("/protected", func(c echo.Context) error {
		user := c.Get("user").(string)
		return c.JSON(http.StatusOK, map[string]string{"message": "Access granted", "user": user})
	}, authMiddleware.JWTMiddleware(authService))

	// Start server
	log.Fatal(e.Start(":8080"))
}

func cacheTest() {
	cfg := &cache.CacheConfig{
		Address:  "localhost:6379",
		Password: "",
		DB:       0,
	}

	redisClient := cache.NewRedisClient(cfg)
	cacheLayer := cache.NewCache(redisClient)

	ctx := context.Background()

	// Map-like operations
	err := cacheLayer.HSet(ctx, "user:123", "name", "John Doe")
	if err != nil {
		log.Fatalf("Failed to HSet: %v", err)
	}

	name, err := cacheLayer.HGet(ctx, "user:123", "name")
	if err != nil {
		log.Fatalf("Failed to HGet: %v", err)
	}
	fmt.Printf("Name: %s\n", name)

	// Set-like operations
	err = cacheLayer.SAdd(ctx, "colors", "red", "blue", "green")
	if err != nil {
		log.Fatalf("Failed to SAdd: %v", err)
	}

	members, err := cacheLayer.SMembers(ctx, "colors")
	if err != nil {
		log.Fatalf("Failed to SMembers: %v", err)
	}
	fmt.Printf("Colors: %v\n", members)

	isMember, err := cacheLayer.SIsMember(ctx, "colors", "red")
	if err != nil {
		log.Fatalf("Failed to SIsMember: %v", err)
	}
	fmt.Printf("Is red a member? %v\n", isMember)

	// Set a cache value
	err = cacheLayer.Set(ctx, "key", "value", 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to set cache: %v", err)
	}

	// Get the cache value
	val, err := cacheLayer.Get(ctx, "key")
	if err != nil {
		log.Fatalf("Failed to get cache: %v", err)
	}
	fmt.Printf("Cache value: %s\n", val)

}
