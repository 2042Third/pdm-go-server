package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
	"pdm-go-server/internal/auth"
	"pdm-go-server/internal/authMiddleware"
	"pdm-go-server/internal/cache"
	"pdm-go-server/internal/db"
	"pdm-go-server/internal/handlers"
	"pdm-go-server/internal/services"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Initialize database
	database := db.NewDatabase()

	// Initialize
	redisUrl := os.Getenv("REDIS_URL")
	redisPass := os.Getenv("REDIS_PASSWORD")
	cfg := &cache.CacheConfig{
		Address:  redisUrl,
		Password: redisPass,
		DB:       0,
	}

	// Initialize Redis client
	redisClient := cache.NewRedisClient(cfg)
	cacheLayer := cache.NewCache(redisClient)

	// Initialize RabbitMQ connection
	rabbitMQCtx, _ := services.NewRabbitMQHandler()

	if rabbitMQCtx == nil {
		log.Fatalf("Failed to initialize RabbitMQ")
	}

	// Initialize storage
	storage := services.NewStorage(database.DB, rabbitMQCtx, cacheLayer)

	// Initialize auth service
	// Load keys from environment
	privateKey, publicKey, err := auth.LoadKeys()
	if err != nil {
		log.Fatalf("Failed to load keys: %v", err)
	}

	// JWT config
	// Create middleware config
	jwtConfig := authMiddleware.JWTMiddlewareConfig{
		PublicKey: publicKey,
	}

	// Create auth service
	authService := auth.NewAuthService(privateKey, publicKey)
	err = authService.HealthCheck()
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	// Initialize handlers
	userHandler := handlers.NewUserHandler(storage, authService)
	notesHandler := handlers.NewNotesHandler(storage, authService)

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

	// Protected routes group
	api := e.Group("")
	api.Use(authMiddleware.CreateJWTMiddleware(jwtConfig))

	// User routes
	api.GET("/api/user/logout", userHandler.Logout)
	api.GET("/api/user", userHandler.GetUserInfo)
	api.POST("/api/user", userHandler.GetUserInfo)

	// Notes routes
	api.POST("/api/notes/new", notesHandler.CreateNote)
	api.GET("/api/notes", notesHandler.GetNotes)
	api.POST("/api/notes", notesHandler.UpdateNotes)

	// Other routes
	api.GET("/protected", func(c echo.Context) error {
		email := c.Get("email").(string)
		return c.JSON(http.StatusOK, map[string]string{"message": "Access granted", "email": email})
	})

	// Start server
	log.Fatal(e.Start(":8080"))
}
