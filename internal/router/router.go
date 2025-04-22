package router

import (
	"net/http"

	"github.com/Abigotado/abi_banking/internal/config"
	"github.com/Abigotado/abi_banking/internal/handlers"
	"github.com/Abigotado/abi_banking/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func NewRouter(cfg *config.Config, logger *logrus.Logger, h *handlers.Handlers) http.Handler {
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
	public.HandleFunc("/register", h.RegisterHandler).Methods("POST")
	public.HandleFunc("/login", h.LoginHandler).Methods("POST")

	// Protected routes
	protected := apiRouter.PathPrefix("/").Subrouter()
	protected.Use(middleware.Auth(cfg.JWT.Secret))

	// Account routes
	protected.HandleFunc("/accounts", h.CreateAccountHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}", h.GetAccountHandler).Methods("GET")
	protected.HandleFunc("/accounts/transfer", h.TransferHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}/deposit", h.DepositHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}/withdraw", h.WithdrawHandler).Methods("POST")

	// Credit routes
	protected.HandleFunc("/credits", h.CreateCreditHandler).Methods("POST")
	protected.HandleFunc("/credits/{id}", h.GetCreditHandler).Methods("GET")
	protected.HandleFunc("/credits/{id}/schedule", h.GetPaymentScheduleHandler).Methods("GET")
	protected.HandleFunc("/credits/{id}/pay", h.PayCreditHandler).Methods("POST")

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
