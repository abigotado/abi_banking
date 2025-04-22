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
	creditRepo  *repository.CreditRepository
	logger      *logrus.Logger
}

func NewAccountService(logger *logrus.Logger) *AccountService {
	return &AccountService{
		accountRepo: repository.NewAccountRepository(),
		creditRepo:  repository.NewCreditRepository(),
		logger:      logger,
	}
}

func (s *AccountService) CreateAccount(req *models.CreateAccountRequest) (*models.Account, error) {
	account := &models.Account{
		UserID:    req.UserID,
		Balance:   req.Balance,
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

func (s *AccountService) Transfer(req *models.TransferRequest) error {
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
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Type:          "transfer",
		CreatedAt:     time.Now(),
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
		ToAccountID: accountID,
		Amount:      amount,
		Type:        "deposit",
		CreatedAt:   time.Now(),
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
		FromAccountID: accountID,
		Amount:        amount,
		Type:          "withdrawal",
		CreatedAt:     time.Now(),
	}

	if err := s.accountRepo.CreateTransaction(transaction); err != nil {
		s.logger.WithError(err).Error("Failed to create transaction record")
		return errors.New("internal server error")
	}

	return nil
}

// Credit-related methods

func (s *AccountService) CreateCredit(req *models.CreateCreditRequest) (*models.Credit, error) {
	credit := &models.Credit{
		UserID:          req.UserID,
		AccountID:       req.AccountID,
		Amount:          req.Amount,
		InterestRate:    req.InterestRate,
		TermMonths:      req.TermMonths,
		RemainingAmount: req.Amount,
		Status:          "ACTIVE",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.creditRepo.Create(credit); err != nil {
		s.logger.WithError(err).Error("Failed to create credit")
		return nil, errors.New("internal server error")
	}

	// Generate payment schedule
	schedule := models.GeneratePaymentSchedule(credit, time.Now())
	for _, payment := range schedule {
		payment.CreditID = credit.ID
		if err := s.creditRepo.CreatePaymentSchedule(payment); err != nil {
			s.logger.WithError(err).Error("Failed to create payment schedule")
			return nil, errors.New("internal server error")
		}
	}

	return credit, nil
}

func (s *AccountService) GetCreditByID(creditID int64) (*models.Credit, error) {
	credit, err := s.creditRepo.GetByID(creditID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get credit by ID")
		return nil, errors.New("credit not found")
	}
	return credit, nil
}

func (s *AccountService) GetCreditsByUserID(userID int64) ([]*models.Credit, error) {
	credits, err := s.creditRepo.GetByUserID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get credits by user ID")
		return nil, errors.New("internal server error")
	}
	return credits, nil
}

func (s *AccountService) PayCredit(creditID int64, amount float64) error {
	credit, err := s.creditRepo.GetByID(creditID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get credit")
		return errors.New("credit not found")
	}

	if credit.Status != "ACTIVE" {
		return errors.New("credit is not active")
	}

	if amount <= 0 {
		return errors.New("payment amount must be greater than zero")
	}

	if amount > credit.RemainingAmount {
		return errors.New("payment amount exceeds remaining credit amount")
	}

	// Start transaction
	tx, err := s.creditRepo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get next pending payment
	schedule, err := s.creditRepo.GetPaymentSchedule(creditID)
	if err != nil {
		return fmt.Errorf("failed to get payment schedule: %w", err)
	}

	var nextPayment *models.PaymentSchedule
	for _, payment := range schedule {
		if payment.Status == "PENDING" {
			nextPayment = payment
			break
		}
	}

	if nextPayment == nil {
		return errors.New("no pending payments found")
	}

	// Update payment status
	nextPayment.Status = "PAID"
	if err := s.creditRepo.UpdatePaymentSchedule(nextPayment); err != nil {
		return fmt.Errorf("failed to update payment schedule: %w", err)
	}

	// Calculate new remaining amount and update credit status if needed
	credit.RemainingAmount -= amount
	if credit.RemainingAmount == 0 {
		credit.Status = "COMPLETED"
		if err := s.creditRepo.Update(credit); err != nil {
			return fmt.Errorf("failed to update credit: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
