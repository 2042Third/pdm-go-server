package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"log"
	"net/http"
	"pdm-logic-server/pkg/cache"
	"pdm-logic-server/pkg/config"
	"pdm-logic-server/pkg/db"
	"pdm-logic-server/pkg/health"
	"pdm-logic-server/pkg/metrics"
	"pdm-logic-server/pkg/models"
	"pdm-logic-server/pkg/services"
	"sync"
	"time"
)

type App struct {
	config      *config.Config
	echo        *echo.Echo
	logger      *logrus.Logger
	metrics     *metrics.Metrics
	health      *health.HealthChecker
	storage     *services.Storage
	authService *services.AuthService
}

func NewApp(cfg *config.Config, logger *logrus.Logger) (*App, error) {
	e := echo.New()

	// Initialize metrics
	metricsCollector := metrics.NewMetrics()

	// Initialize dependencies
	db, err := initDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	// Migrate database (only use it when needed)
	//if err := migrateDB(db.DB); err != nil {
	//	log.Fatalf("Failed to migrate database schema: %v", err)
	//	return nil, err
	//}

	cache, err := initCache(cfg, logger)
	if err != nil {
		return nil, err
	}

	// Initialize RabbitMQ connection
	rabbitMQCtx, err := services.NewRabbitMQHandler()
	if err != nil {
		return nil, err
	}

	storage := services.NewStorage(db.DB, rabbitMQCtx, cache)

	// Create health checker
	healthChecker := health.NewHealthChecker(db, cache)

	authService := services.NewAuthService(cfg.Auth.PrivateKey, cfg.Auth.PublicKey)

	app := &App{
		config:      cfg,
		echo:        e,
		logger:      logger,
		metrics:     metricsCollector,
		health:      healthChecker,
		storage:     storage,
		authService: authService,
	}

	// Setup everything
	app.setupMiddleware()
	app.setupRoutes()

	return app, nil
}

func (a *App) Start() error {
	// Create custom server with timeouts
	server := &http.Server{
		Addr:         ":" + a.config.Server.Port,
		Handler:      a.echo, // Echo instance
		ReadTimeout:  a.config.Server.ReadTimeout,
		WriteTimeout: a.config.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Log server startup
	a.logger.WithFields(logrus.Fields{
		"port":         a.config.Server.Port,
		"env":          a.config.Server.Environment,
		"readTimeout":  a.config.Server.ReadTimeout,
		"writeTimeout": a.config.Server.WriteTimeout,
	}).Info("Starting server")

	// Pre-startup health check
	if err := a.performPreflightChecks(); err != nil {
		a.logger.WithError(err).Errorf("preflight checks failed: %s", err.Error())
		return fmt.Errorf("preflight checks failed: %w", err)
	}

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		a.logger.WithError(err).Errorf("failed to start server: %s", err.Error())
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Initiating graceful shutdown...")

	// Create a WaitGroup to track all shutdown operations
	var wg sync.WaitGroup

	// Create error channel to collect errors from goroutines
	errChan := make(chan error, 3)

	// Shutdown HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.logger.Info("Shutting down HTTP server...")

		if err := a.echo.Shutdown(ctx); err != nil {
			errChan <- fmt.Errorf("error shutting down HTTP server: %w", err)
			return
		}
		a.logger.Info("HTTP server shutdown complete")
	}()

	// Close database connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.logger.Info("Closing database connections...")

		// Get SQL DB instance from GORM
		sqlDB, err := a.storage.DB.DB()
		if err != nil {
			errChan <- fmt.Errorf("error getting SQL DB instance: %w", err)
			return
		}

		// Close the connection pool
		if err := sqlDB.Close(); err != nil {
			errChan <- fmt.Errorf("error closing database connections: %w", err)
			return
		}
		a.logger.Info("Database connections closed")
	}()

	// Close cache connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.logger.Info("Closing cache connections...")

		//if err := a.storage.Ch.Close(); err != nil {
		//	errChan <- fmt.Errorf("error closing cache connections: %w", err)
		//	return
		//}
		//a.logger.Info("Cache connections closed")
	}()

	// Wait for all cleanup tasks or context deadline
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for either completion or timeout
	select {
	case <-done:
		a.logger.Info("All connections closed successfully")
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}

	// Check if any errors occurred during shutdown
	close(errChan)
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// Log shutdown metrics if enabled
	if a.config.Metrics.Enabled {
		a.logShutdownMetrics()
	}

	// If there were any errors, combine them
	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	a.logger.Info("Graceful shutdown completed")
	return nil
}

func (a *App) logShutdownMetrics() {
	metrics := map[string]interface{}{
		"shutdown_duration": time.Since(time.Now()),
		//"active_connections": a.echo.Stats.CurrentRequests,
		//"total_requests":     a.echo.Stats.RequestCount,
	}

	a.logger.WithFields(logrus.Fields{
		"metrics": metrics,
	}).Info("Shutdown metrics")
}

// Helper function to perform checks before server starts
func (a *App) performPreflightChecks() error {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check database connection
	//if err := a.db.PingContext(ctx); err != nil {
	//	return fmt.Errorf("database connection check failed: %w", err)
	//}

	// Check cache connection
	//if err := a.cache.Ping(ctx); err != nil {
	//	return fmt.Errorf("cache connection check failed: %w", err)
	//}

	// Check if required environment variables are set
	if err := a.validateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}

	return nil
}

func (a *App) validateEnvironment() error {
	required := []string{
		"JWT_PRIVATE_KEY",
		"JWT_PUBLIC_KEY",
		"DB_HOST",
		"DB_NAME",
	}

	for _, env := range required {
		if value := a.config.GetEnv(env); value == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}

	return nil
}

func initDB(cfg *config.Config, logger *logrus.Logger) (*db.Database, error) {
	// Initialize database
	return db.NewDatabase(), nil
}

func migrateDB(db *gorm.DB) error {
	// Run migrations
	if err := db.AutoMigrate(&models.SessionKey{}); err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}
	log.Println("Migration for SessionKey completed!")

	if err := db.AutoMigrate(&models.RefreshKey{}); err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}
	log.Println("Migration for RefreshKey completed!")

	log.Println("Database migration completed!")

	return nil
}

func initCache(cfg *config.Config, logger *logrus.Logger) (*cache.RedisCache, error) {

	// Initialize Redis client
	redisClient := cache.NewRedisClient(&cfg.Redis)
	cacheLayer := cache.NewCache(redisClient)

	return cacheLayer, nil
}

func (a *App) setupMiddleware() {
	// Add logging middleware
	//a.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
	//	return func(c echo.Context) error {
	//		start := time.Now()
	//
	//		err := next(c)
	//
	//		a.logger.WithFields(logrus.Fields{
	//			"method":     c.Request().Method,
	//			"path":       c.Request().URL.Path,
	//			"status":     c.Response().Status,
	//			"latency":    time.Since(start).String(),
	//			"ip":         c.RealIP(),
	//			"user_agent": c.Request().UserAgent(),
	//		}).Info("Request completed")
	//
	//		return err
	//	}
	//})

	a.echo.Use(middleware.Logger())
	a.echo.Use(middleware.Recover())
	a.echo.Use(cspMiddleware) // Add CSP middleware for static content
	a.echo.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentSecurityPolicy: "", // We handle CSP in our middleware
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	// Add metrics middleware if enabled
	if a.config.Metrics.Enabled {
		a.echo.Use(a.metrics.Middleware())
	}
}

// generateNonce creates a random nonce
func generateNonce() string {
	nonce := make([]byte, 16)
	rand.Read(nonce)
	return base64.StdEncoding.EncodeToString(nonce)
}

// CSP middleware that adds nonce and headers
func cspMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		nonce := generateNonce()
		c.Set("nonce", nonce)

		// Set CSP header with nonce
		c.Response().Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'nonce-"+nonce+"'; "+
				"style-src 'self'")

		return next(c)
	}
}
