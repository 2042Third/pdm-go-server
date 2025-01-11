package app

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"pdm-logic-server/pkg/handlers"
	"pdm-logic-server/pkg/middleware"
)

func (a *App) setupRoutes() {
	if a.config.Server.Environment == "development" {
		a.echo.Debug = true
	}
	// Serve static content
	a.echo.Static("/static", a.config.StaticContent.InternalPath+"/static")
	a.logger.Info("Serving static content from: " + a.config.StaticContent.InternalPath)

	// Health check endpoint
	a.echo.GET("/health", a.health.Handler)

	// Metrics endpoint if enabled
	if a.config.Metrics.Enabled {
		a.echo.GET(a.config.Metrics.Path, echo.WrapHandler(promhttp.Handler()))
	}

	// Create base handler
	baseHandler := handlers.NewBaseHandler(a.storage, a.authService, a.logger, a.config)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(baseHandler)
	notesHandler := handlers.NewNotesHandler(baseHandler)
	statusHandler := handlers.NewStatusHandler(baseHandler, a.config.StaticContent.StatusPassword)
	statusHandler.SetupRenderer(a.echo, a.config.StaticContent.InternalPath)

	// Initialize validator
	validator := middleware.NewCustomValidator()
	a.echo.Validator = validator

	// Public routes
	a.echo.POST("/login", userHandler.Login)
	a.echo.POST("/signup", userHandler.Register)
	a.echo.POST("/signup/verify", userHandler.ValidateVerificationCode)
	a.echo.GET("/status/*", statusHandler.StatusHandlerFunc)

	// Protected routes
	api := a.echo.Group("/api")
	api.Use(middleware.CreateJWTMiddleware(a.config.Auth.PublicKey))

	// User routes
	api.GET("/user/logout", userHandler.Logout)
	api.GET("/user", userHandler.GetUserInfo)

	// Notes routes
	api.GET("/notes", notesHandler.GetNotes)
	api.POST("/notes", notesHandler.CreateNote)
	api.PUT("/notes", notesHandler.UpdateNotes)
	api.DELETE("/notes", notesHandler.DeleteNotes)
}
