package router

import (
	"net/http"

	"github.com/Abigotado/abi_banking/internal/config"
	"github.com/Abigotado/abi_banking/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func NewRouter(cfg *config.Config, logger *logrus.Logger) http.Handler {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(
		middleware.Logging(logger),
		middleware.Recovery(logger),
		middleware.CORS(cfg.API.CORSAllowedOrigins),
	)

	// API version prefix
	apiRouter := router.PathPrefix(cfg.API.Prefix).Subrouter()

	// Public routes
	public := apiRouter.PathPrefix("/public").Subrouter()
	public.HandleFunc("/register", registerHandler).Methods("POST")
	public.HandleFunc("/login", loginHandler).Methods("POST")

	// Protected routes
	protected := apiRouter.PathPrefix("/").Subrouter()
	protected.Use(middleware.Auth(cfg.JWT.Secret))

	// Account routes
	protected.HandleFunc("/accounts", createAccountHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}", getAccountHandler).Methods("GET")
	protected.HandleFunc("/accounts/transfer", transferHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}/deposit", depositHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}/withdraw", withdrawHandler).Methods("POST")

	// Card routes
	protected.HandleFunc("/cards", createCardHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}", getCardHandler).Methods("GET")
	protected.HandleFunc("/cards/{id}/block", blockCardHandler).Methods("POST")

	// Credit routes
	protected.HandleFunc("/credits", createCreditHandler).Methods("POST")
	protected.HandleFunc("/credits/{id}", getCreditHandler).Methods("GET")
	protected.HandleFunc("/credits/{id}/schedule", getCreditScheduleHandler).Methods("GET")
	protected.HandleFunc("/credits/{id}/pay", payCreditHandler).Methods("POST")

	// Analytics routes
	protected.HandleFunc("/analytics/transactions", getTransactionAnalyticsHandler).Methods("GET")
	protected.HandleFunc("/analytics/credits", getCreditAnalyticsHandler).Methods("GET")

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
