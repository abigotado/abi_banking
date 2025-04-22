package scheduler

import (
	"time"

	"github.com/Abigotado/abi_banking/internal/models"
	"github.com/Abigotado/abi_banking/internal/repository"
	"github.com/Abigotado/abi_banking/internal/service"
	"github.com/sirupsen/logrus"
)

// PaymentScheduler handles automatic payment processing
type PaymentScheduler struct {
	creditRepo *repository.CreditRepository
	accountSvc *service.AccountService
	logger     *logrus.Logger
	ticker     *time.Ticker
	done       chan bool
}

// NewPaymentScheduler creates a new payment scheduler
func NewPaymentScheduler(
	creditRepo *repository.CreditRepository,
	accountSvc *service.AccountService,
	logger *logrus.Logger,
) *PaymentScheduler {
	return &PaymentScheduler{
		creditRepo: creditRepo,
		accountSvc: accountSvc,
		logger:     logger,
		ticker:     time.NewTicker(12 * time.Hour),
		done:       make(chan bool),
	}
}

// Start begins the scheduler
func (s *PaymentScheduler) Start() {
	s.logger.Info("Starting payment scheduler")
	go s.run()
}

// Stop stops the scheduler
func (s *PaymentScheduler) Stop() {
	s.logger.Info("Stopping payment scheduler")
	s.ticker.Stop()
	s.done <- true
}

// run executes the scheduler loop
func (s *PaymentScheduler) run() {
	for {
		select {
		case <-s.ticker.C:
			s.processPayments()
		case <-s.done:
			return
		}
	}
}

// processPayments handles automatic payment processing
func (s *PaymentScheduler) processPayments() {
	s.logger.Info("Processing scheduled payments")

	// Get all active credits with due payments
	credits, err := s.creditRepo.GetCreditsWithDuePayments()
	if err != nil {
		s.logger.Errorf("Failed to get credits with due payments: %v", err)
		return
	}

	for _, credit := range credits {
		// Get the next payment
		payment, err := s.creditRepo.GetNextPayment(credit.ID)
		if err != nil {
			s.logger.Errorf("Failed to get next payment for credit %d: %v", credit.ID, err)
			continue
		}

		// Check if payment is due
		if time.Now().Before(payment.DueDate) {
			continue
		}

		// Process payment
		if err := s.processPayment(credit, payment); err != nil {
			s.logger.Errorf("Failed to process payment for credit %d: %v", credit.ID, err)
			continue
		}
	}
}

// processPayment handles a single payment
func (s *PaymentScheduler) processPayment(credit *models.Credit, payment *models.PaymentSchedule) error {
	// Start transaction
	tx, err := s.creditRepo.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if account has sufficient funds
	account, err := s.accountSvc.GetAccountByID(credit.AccountID)
	if err != nil {
		return err
	}

	if account.Balance < payment.Amount {
		// Apply penalty for insufficient funds
		penalty := payment.Amount * 0.1 // 10% penalty
		payment.Amount += penalty
		s.logger.Warnf("Insufficient funds for credit %d, applying penalty of %.2f", credit.ID, penalty)
	}

	// Withdraw funds from account
	if err := s.accountSvc.Withdraw(credit.AccountID, payment.Amount); err != nil {
		return err
	}

	// Update payment status
	if err := s.creditRepo.UpdatePaymentStatus(payment.ID, string(models.PaymentStatusPaid)); err != nil {
		return err
	}

	// Update credit remaining amount
	if err := s.creditRepo.UpdateRemainingAmount(credit.ID, credit.RemainingAmount-payment.Amount); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	s.logger.Infof("Successfully processed payment for credit %d", credit.ID)
	return nil
}
