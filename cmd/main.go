package main

import (
	"net/http"
	"os"

	"github.com/Abigotado/abi_banking/internal/database"
	"github.com/Abigotado/abi_banking/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// Initialize database
	if err := database.InitDB(); err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Create router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/register", registerHandler).Methods("POST")
	router.HandleFunc("/login", loginHandler).Methods("POST")

	// Protected routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	protected.HandleFunc("/accounts", createAccountHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}", getAccountHandler).Methods("GET")
	protected.HandleFunc("/cards", createCardHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}", getCardHandler).Methods("GET")
	protected.HandleFunc("/transfer", transferHandler).Methods("POST")
	protected.HandleFunc("/credits", createCreditHandler).Methods("POST")
	protected.HandleFunc("/credits/{id}/schedule", getCreditScheduleHandler).Methods("GET")
	protected.HandleFunc("/analytics", getAnalyticsHandler).Methods("GET")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Infof("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}
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
