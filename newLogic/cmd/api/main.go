package main

import (
	"context"
	"os"
	"os/signal"
	"pdm-logic-server/pkg/app"
	"pdm-logic-server/pkg/config"
	"pdm-logic-server/pkg/logging"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		panic(err)
	}

	// Create new application instance with logger
	application, err := app.NewApp(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create application")
	}

	// Start application
	go func() {
		if err := application.Start(); err != nil {
			logger.WithError(err).Fatal("Failed to start application")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown application
	if err := application.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Error during shutdown")
	}
}
