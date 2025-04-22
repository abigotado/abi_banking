package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Abigotado/abi_banking/internal/models"
	"github.com/Abigotado/abi_banking/internal/repository"
	"github.com/sirupsen/logrus"
)

type AccountService struct {
	accountRepo *repository.AccountRepository
	logger      *logrus.Logger
}

func NewAccountService(logger *logrus.Logger) *AccountService {
	return &AccountService{
		accountRepo: repository.NewAccountRepository(),
		logger:      logger,
	}
}

type CreateAccountRequest struct {
	UserID   int64  `json:"user_id" validate:"required"`
	Currency string `json:"currency" validate:"required,len=3"`
}

type TransferRequest struct {
	FromAccountID int64   `json:"from_account_id" validate:"required"`
	ToAccountID   int64   `json:"to_account_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	Description   string  `json:"description"`
}

func (s *AccountService) CreateAccount(req *CreateAccountRequest) (*models.Account, error) {
	account := &models.Account{
		UserID:    req.UserID,
		Balance:   0,
		Currency:  req.Currency,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.accountRepo.Create(account); err != nil {
		s.logger.WithError(err).Error("Failed to create account")
		return nil, errors.New("internal server error")
	}

	return account, nil
}

func (s *AccountService) GetAccountByID(accountID int64) (*models.Account, error) {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get account by ID")
		return nil, errors.New("account not found")
	}

	return account, nil
}

func (s *AccountService) GetUserAccounts(userID int64) ([]*models.Account, error) {
	accounts, err := s.accountRepo.GetByUserID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user accounts")
		return nil, errors.New("internal server error")
	}

	return accounts, nil
}

func (s *AccountService) Transfer(req *TransferRequest) error {
	// Start a database transaction
	tx, err := s.accountRepo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get source account
	srcAccount, err := s.accountRepo.GetByID(req.FromAccountID)
	if err != nil {
		return fmt.Errorf("failed to get source account: %w", err)
	}

	// Get destination account
	dstAccount, err := s.accountRepo.GetByID(req.ToAccountID)
	if err != nil {
		return fmt.Errorf("failed to get destination account: %w", err)
	}

	// Validate currencies match
	if srcAccount.Currency != dstAccount.Currency {
		return errors.New("currency mismatch between accounts")
	}

	// Check if source account has sufficient funds
	if srcAccount.Balance < req.Amount {
		return errors.New("insufficient funds")
	}

	// Update balances
	srcAccount.Balance -= req.Amount
	dstAccount.Balance += req.Amount

	// Update source account
	if err := s.accountRepo.UpdateBalance(srcAccount.ID, srcAccount.Balance); err != nil {
		return fmt.Errorf("failed to update source account balance: %w", err)
	}

	// Update destination account
	if err := s.accountRepo.UpdateBalance(dstAccount.ID, dstAccount.Balance); err != nil {
		return fmt.Errorf("failed to update destination account balance: %w", err)
	}

	// Create transaction record
	transaction := &models.Transaction{
		FromAccountID:   req.FromAccountID,
		ToAccountID:     req.ToAccountID,
		Amount:          req.Amount,
		Currency:        srcAccount.Currency,
		Description:     req.Description,
		TransactionType: "transfer",
		CreatedAt:       time.Now(),
	}

	if err := s.accountRepo.CreateTransaction(transaction); err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *AccountService) Deposit(accountID int64, amount float64) error {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get account")
		return errors.New("account not found")
	}

	newBalance := account.Balance + amount
	if err := s.accountRepo.UpdateBalance(accountID, newBalance); err != nil {
		s.logger.WithError(err).Error("Failed to update account balance")
		return errors.New("internal server error")
	}

	// Create transaction record
	transaction := &models.Transaction{
		ToAccountID:     accountID,
		Amount:          amount,
		Currency:        account.Currency,
		Description:     "Deposit",
		TransactionType: "DEPOSIT",
		CreatedAt:       time.Now(),
	}

	if err := s.accountRepo.CreateTransaction(transaction); err != nil {
		s.logger.WithError(err).Error("Failed to create transaction record")
		return errors.New("internal server error")
	}

	return nil
}

func (s *AccountService) Withdraw(accountID int64, amount float64) error {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get account")
		return errors.New("account not found")
	}

	if account.Balance < amount {
		return errors.New("insufficient funds")
	}

	newBalance := account.Balance - amount
	if err := s.accountRepo.UpdateBalance(accountID, newBalance); err != nil {
		s.logger.WithError(err).Error("Failed to update account balance")
		return errors.New("internal server error")
	}

	// Create transaction record
	transaction := &models.Transaction{
		FromAccountID:   accountID,
		Amount:          amount,
		Currency:        account.Currency,
		Description:     "Withdrawal",
		TransactionType: "WITHDRAWAL",
		CreatedAt:       time.Now(),
	}

	if err := s.accountRepo.CreateTransaction(transaction); err != nil {
		s.logger.WithError(err).Error("Failed to create transaction record")
		return errors.New("internal server error")
	}

	return nil
}
