package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Abigotado/abi_banking/internal/config"
	"github.com/Abigotado/abi_banking/internal/database"
	"github.com/Abigotado/abi_banking/internal/handlers"
	"github.com/Abigotado/abi_banking/internal/router"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

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

	// Create HTTP server
	// Initialize handlers
	handlers := handlers.New(cfg, logger)

	server := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: router.NewRouter(cfg, handlers, logger),
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Server starting on port %s", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}

func setupRouter(cfg *config.Config, logger *logrus.Logger) http.Handler {
	// TODO: Implement router setup with middleware
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
}

// Placeholder handlers - to be implemented
func registerHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement registration
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement login
}

func createAccountHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement account creation
}

func getAccountHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement account retrieval
}

func createCardHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement card creation
}

func getCardHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement card retrieval
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement money transfer
}

func createCreditHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement credit creation
}

func getCreditScheduleHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement credit schedule retrieval
}

func getAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement analytics
}
