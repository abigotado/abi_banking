package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Abigotado/abi_banking/internal/config"
	"github.com/Abigotado/abi_banking/internal/database"
	"github.com/Abigotado/abi_banking/internal/middleware"
	"github.com/Abigotado/abi_banking/internal/models"
	"github.com/Abigotado/abi_banking/internal/repository"
	"github.com/Abigotado/abi_banking/internal/service"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handlers struct {
	userService    *service.UserService
	accountService *service.AccountService
	creditService  *service.CreditService
	cardService    *service.CardService
	logger         *logrus.Logger
}

func New(cfg *config.Config, logger *logrus.Logger) *Handlers {
	creditRepo := repository.NewCreditRepository()
	cardRepo := repository.NewCardRepository(database.DB, logger)
	accountRepo := repository.NewAccountRepository()

	return &Handlers{
		userService:    service.NewUserService(logger),
		accountService: service.NewAccountService(logger),
		creditService:  service.NewCreditService(creditRepo, logger),
		cardService:    service.NewCardService(cardRepo, accountRepo, logger),
		logger:         logger,
	}
}

// RegisterHandler handles user registration
func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.userService.Register(&req); err != nil {
		h.logger.WithError(err).Error("Failed to register user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// LoginHandler handles user login
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.Login(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to login user")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateAccountHandler handles account creation
func (h *Handlers) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := middleware.GetRequestBodyFromContext(r.Context()).(*models.CreateAccountRequest)
	if !ok {
		h.logger.Error("Failed to get request body from context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	account, err := h.accountService.CreateAccount(req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create account")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

// GetAccountHandler handles account retrieval
func (h *Handlers) GetAccountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid account ID")
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.GetAccountByID(accountID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get account")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

// GetUserAccountsHandler handles user accounts retrieval
func (h *Handlers) GetUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseInt(vars["user_id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	accounts, err := h.accountService.GetUserAccounts(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user accounts")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// TransferHandler handles money transfer between accounts
func (h *Handlers) TransferHandler(w http.ResponseWriter, r *http.Request) {
	var req models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.accountService.Transfer(&req); err != nil {
		h.logger.WithError(err).Error("Failed to transfer money")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateCreditHandler handles credit creation
func (h *Handlers) CreateCreditHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	credit, err := h.creditService.CreateCredit(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create credit")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(credit)
}

// GetCreditHandler handles credit retrieval
func (h *Handlers) GetCreditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid credit ID")
		http.Error(w, "Invalid credit ID", http.StatusBadRequest)
		return
	}

	credit, err := h.creditService.GetCreditByID(creditID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get credit")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(credit)
}

// GetUserCreditsHandler handles user credits retrieval
func (h *Handlers) GetUserCreditsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseInt(vars["user_id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	credits, err := h.creditService.GetCreditsByUserID(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user credits")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(credits)
}

// PayCreditHandler handles credit payment
func (h *Handlers) PayCreditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid credit ID")
		http.Error(w, "Invalid credit ID", http.StatusBadRequest)
		return
	}

	var req models.PayCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.creditService.PayCredit(creditID, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to pay credit")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetPaymentScheduleHandler handles payment schedule retrieval
func (h *Handlers) GetPaymentScheduleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid credit ID")
		http.Error(w, "Invalid credit ID", http.StatusBadRequest)
		return
	}

	credit, err := h.creditService.GetCreditByID(creditID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get credit")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	schedule := models.GeneratePaymentSchedule(credit, time.Now())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// DepositHandler handles account deposits
func (h *Handlers) DepositHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID int64   `json:"account_id" validate:"required"`
		Amount    float64 `json:"amount" validate:"required,gt=0"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.accountService.Deposit(req.AccountID, req.Amount); err != nil {
		h.logger.WithError(err).Error("Failed to deposit money")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// WithdrawHandler handles account withdrawals
func (h *Handlers) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID int64   `json:"account_id" validate:"required"`
		Amount    float64 `json:"amount" validate:"required,gt=0"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.accountService.Withdraw(req.AccountID, req.Amount); err != nil {
		h.logger.WithError(err).Error("Failed to withdraw money")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateCardHandler handles card creation
func (h *Handlers) CreateCardHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	card, err := h.cardService.CreateCard(userID, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create card")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card.ToResponse())
}

// GetCardHandler handles card retrieval
func (h *Handlers) GetCardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cardID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid card ID")
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	card, err := h.cardService.GetCard(userID, cardID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get card")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card.ToResponse())
}

// GetUserCardsHandler handles user cards retrieval
func (h *Handlers) GetUserCardsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cards, err := h.cardService.GetUserCards(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user cards")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert cards to responses
	responses := make([]*models.CardResponse, len(cards))
	for i, card := range cards {
		responses[i] = card.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// BlockCardHandler handles card blocking
func (h *Handlers) BlockCardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cardID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid card ID")
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.cardService.BlockCard(userID, cardID); err != nil {
		h.logger.WithError(err).Error("Failed to block card")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UnblockCardHandler handles card unblocking
func (h *Handlers) UnblockCardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cardID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid card ID")
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.cardService.UnblockCard(userID, cardID); err != nil {
		h.logger.WithError(err).Error("Failed to unblock card")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteCardHandler handles card deletion
func (h *Handlers) DeleteCardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cardID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid card ID")
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.cardService.DeleteCard(userID, cardID); err != nil {
		h.logger.WithError(err).Error("Failed to delete card")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetTransactionAnalyticsHandler handles transaction analytics retrieval
func (h *Handlers) GetTransactionAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Convert dates to time.Time
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		h.logger.WithError(err).Error("Invalid start date")
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		h.logger.WithError(err).Error("Invalid end date")
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	analytics, err := h.accountService.GetTransactionAnalytics(userID, start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get transaction analytics")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// GetCreditAnalyticsHandler handles credit analytics retrieval
func (h *Handlers) GetCreditAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.logger.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	analytics, err := h.creditService.GetCreditAnalytics(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get credit analytics")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}
