package router

import (
	"net/http"

	"github.com/Abigotado/abi_banking/internal/config"
	"github.com/Abigotado/abi_banking/internal/handlers"
	"github.com/Abigotado/abi_banking/internal/middleware"
	"github.com/Abigotado/abi_banking/internal/models"
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
		middleware.RequestID(),
		middleware.RateLimiter(cfg.RateLimit.Requests),
		middleware.ContentType("application/json"),
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
	accountRouter := protected.PathPrefix("/accounts").Subrouter()
	accountRouter.HandleFunc("", middleware.ValidateRequest(&models.CreateAccountRequest{})(handlers.CreateAccountHandler)).Methods("POST")
	accountRouter.HandleFunc("/{id}", handlers.GetAccountHandler).Methods("GET")
	accountRouter.HandleFunc("/user/{user_id}", handlers.GetUserAccountsHandler).Methods("GET")
	accountRouter.HandleFunc("/transfer", middleware.ValidateRequest(&models.TransferRequest{})(handlers.TransferHandler)).Methods("POST")
	accountRouter.HandleFunc("/{id}/deposit", middleware.ValidateRequest(&models.DepositRequest{})(handlers.DepositHandler)).Methods("POST")
	accountRouter.HandleFunc("/{id}/withdraw", middleware.ValidateRequest(&models.WithdrawRequest{})(handlers.WithdrawHandler)).Methods("POST")

	// Card routes
	cardRouter := protected.PathPrefix("/cards").Subrouter()
	cardRouter.HandleFunc("", middleware.ValidateRequest(&models.CreateCardRequest{})(handlers.CreateCardHandler)).Methods("POST")
	cardRouter.HandleFunc("/{id}", handlers.GetCardHandler).Methods("GET")
	cardRouter.HandleFunc("/user/{user_id}", handlers.GetUserCardsHandler).Methods("GET")
	cardRouter.HandleFunc("/{id}/block", handlers.BlockCardHandler).Methods("POST")
	cardRouter.HandleFunc("/{id}/unblock", handlers.UnblockCardHandler).Methods("POST")
	cardRouter.HandleFunc("/{id}", handlers.DeleteCardHandler).Methods("DELETE")

	// Credit routes
	creditRouter := protected.PathPrefix("/credits").Subrouter()
	creditRouter.HandleFunc("", middleware.ValidateRequest(&models.CreateCreditRequest{})(handlers.CreateCreditHandler)).Methods("POST")
	creditRouter.HandleFunc("/{id}", handlers.GetCreditHandler).Methods("GET")
	creditRouter.HandleFunc("/user/{user_id}", handlers.GetUserCreditsHandler).Methods("GET")
	creditRouter.HandleFunc("/{id}/schedule", handlers.GetPaymentScheduleHandler).Methods("GET")
	creditRouter.HandleFunc("/{id}/pay", middleware.ValidateRequest(&models.PayCreditRequest{})(handlers.PayCreditHandler)).Methods("POST")

	// Analytics routes
	analyticsRouter := protected.PathPrefix("/analytics").Subrouter()
	analyticsRouter.HandleFunc("/transactions", handlers.GetTransactionAnalyticsHandler).Methods("GET")
	analyticsRouter.HandleFunc("/credits", handlers.GetCreditAnalyticsHandler).Methods("GET")

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
