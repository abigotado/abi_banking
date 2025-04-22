package router

import (
	"net/http"

	"github.com/Abigotado/abi_banking/internal/config"
	"github.com/Abigotado/abi_banking/internal/handlers"
	"github.com/Abigotado/abi_banking/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func NewRouter(
	cfg *config.Config,
	handlers *handlers.Handlers,
	logger *logrus.Logger,
) http.Handler {
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
	public.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	public.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// Protected routes
	protected := apiRouter.PathPrefix("/").Subrouter()
	protected.Use(middleware.Auth(cfg.JWT.Secret))

	// Account routes
	protected.HandleFunc("/accounts", handlers.CreateAccountHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}", handlers.GetAccountHandler).Methods("GET")
	protected.HandleFunc("/accounts/user/{user_id}", handlers.GetUserAccountsHandler).Methods("GET")
	protected.HandleFunc("/accounts/transfer", handlers.TransferHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}/deposit", handlers.DepositHandler).Methods("POST")
	protected.HandleFunc("/accounts/{id}/withdraw", handlers.WithdrawHandler).Methods("POST")

	// Card routes
	protected.HandleFunc("/cards", handlers.CreateCardHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}", handlers.GetCardHandler).Methods("GET")
	protected.HandleFunc("/cards/user/{user_id}", handlers.GetUserCardsHandler).Methods("GET")
	protected.HandleFunc("/cards/{id}/block", handlers.BlockCardHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}/unblock", handlers.UnblockCardHandler).Methods("POST")
	protected.HandleFunc("/cards/{id}", handlers.DeleteCardHandler).Methods("DELETE")

	// Credit routes
	protected.HandleFunc("/credits", handlers.CreateCreditHandler).Methods("POST")
	protected.HandleFunc("/credits/{id}", handlers.GetCreditHandler).Methods("GET")
	protected.HandleFunc("/credits/user/{user_id}", handlers.GetUserCreditsHandler).Methods("GET")
	protected.HandleFunc("/credits/{id}/schedule", handlers.GetPaymentScheduleHandler).Methods("GET")
	protected.HandleFunc("/credits/{id}/pay", handlers.PayCreditHandler).Methods("POST")

	// Analytics routes
	protected.HandleFunc("/analytics/transactions", handlers.GetTransactionAnalyticsHandler).Methods("GET")
	protected.HandleFunc("/analytics/credits", handlers.GetCreditAnalyticsHandler).Methods("GET")

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
