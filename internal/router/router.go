package router

import (
	"log"
	"net/http"

	"github.com/Abigotado/abi_banking/internal/handlers"
	"github.com/Abigotado/abi_banking/internal/middleware"
	"github.com/Abigotado/abi_banking/internal/service"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// NewRouter creates a new router with all routes configured
func NewRouter(userService *service.UserService, accountService *service.AccountService, creditService *service.CreditService, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// Create handlers instance
	h := handlers.NewHandlers(userService, accountService, creditService, logger)

	// Public routes
	router.HandleFunc("/register", h.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", h.LoginHandler).Methods("POST")

	// Protected routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// Account routes
	protected.HandleFunc("/accounts", h.CreateAccountHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}", h.GetAccountHandler).Methods("GET")
	protected.HandleFunc("/accounts/user/{userID}", h.GetUserAccountsHandler).Methods("GET")
	protected.HandleFunc("/accounts/transfer", h.TransferHandler).Methods("POST")
	protected.HandleFunc("/accounts/deposit", h.DepositHandler).Methods("POST")
	protected.HandleFunc("/accounts/withdraw", h.WithdrawHandler).Methods("POST")

	// Credit routes
	protected.HandleFunc("/credits", h.CreateCreditHandler).Methods("POST")
	protected.HandleFunc("/credits/{id}", h.GetCreditHandler).Methods("GET")
	protected.HandleFunc("/credits/user/{userID}", h.GetUserCreditsHandler).Methods("GET")
	protected.HandleFunc("/credits/{id}/pay", h.PayCreditHandler).Methods("POST")
	protected.HandleFunc("/credits/{id}/schedule", h.GetPaymentScheduleHandler).Methods("GET")

	// Log all requests
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
			next.ServeHTTP(w, r)
		})
	})

	return router
}

// Helper function to handle CORS
func handleCORS(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					break
				}
			}
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
