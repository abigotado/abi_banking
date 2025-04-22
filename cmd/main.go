package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Abigotado/abi_banking/internal/config"
	"github.com/Abigotado/abi_banking/internal/database"
	"github.com/Abigotado/abi_banking/internal/handlers"
	"github.com/Abigotado/abi_banking/internal/router"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Load .env file
	if err := godotenv.Load(); err != nil {
		logger.Warnf("Error loading .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		logger.Warnf("Invalid log level %s, using info level", cfg.Log.Level)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Initialize database
	if err := database.InitDB(cfg, logger); err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Initialize handlers
	h := handlers.New(cfg, logger)

	// Initialize router
	r := router.NewRouter(cfg, h, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Starting server on port %s", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Server is shutting down...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited properly")
}
