package service

import (
	"errors"
	"time"

	"github.com/Abigotado/abi_banking/internal/models"
	"github.com/Abigotado/abi_banking/internal/repository"
)

type CreditService struct {
	creditRepo     *repository.CreditRepository
	accountService *AccountService
}

func NewCreditService(creditRepo *repository.CreditRepository, accountService *AccountService) *CreditService {
	return &CreditService{
		creditRepo:     creditRepo,
		accountService: accountService,
	}
}

// CreateCredit creates a new credit and generates its payment schedule
func (s *CreditService) CreateCredit(req *models.CreateCreditRequest) (*models.Credit, error) {
	// Validate user exists and has sufficient credit score (to be implemented)

	credit := &models.Credit{
		UserID:          req.UserID,
		Amount:          req.Amount,
		InterestRate:    req.InterestRate,
		TermMonths:      req.TermMonths,
		RemainingAmount: req.Amount,
		Status:          "ACTIVE",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := s.creditRepo.Create(credit)
	if err != nil {
		return nil, err
	}

	return credit, nil
}

// GetCreditByID retrieves a credit by its ID
func (s *CreditService) GetCreditByID(creditID int64) (*models.Credit, error) {
	credit, err := s.creditRepo.GetByID(creditID)
	if err != nil {
		return nil, err
	}

	return credit, nil
}

// GetCreditsByUserID retrieves all credits for a user
func (s *CreditService) GetCreditsByUserID(userID int64) ([]*models.Credit, error) {
	credits, err := s.creditRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return credits, nil
}

// PayCredit processes a payment for a credit
func (s *CreditService) PayCredit(creditID int64, req *models.PayCreditRequest) error {
	credit, err := s.creditRepo.GetByID(creditID)
	if err != nil {
		return err
	}

	if credit.Status != "ACTIVE" {
		return errors.New("credit is not active")
	}

	if req.Amount <= 0 {
		return errors.New("payment amount must be greater than zero")
	}

	if req.Amount > credit.RemainingAmount {
		return errors.New("payment amount exceeds remaining credit amount")
	}

	// Get next pending payment from schedule
	schedule, err := s.creditRepo.GetPaymentSchedule(creditID)
	if err != nil {
		return err
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
	err = s.creditRepo.UpdatePaymentStatus(nextPayment.ID, "PAID")
	if err != nil {
		return err
	}

	// Update credit remaining amount
	credit.RemainingAmount -= req.Amount
	if credit.RemainingAmount == 0 {
		credit.Status = "COMPLETED"
	}

	return nil
}

// GetPaymentSchedule retrieves the payment schedule for a credit
func (s *CreditService) GetPaymentSchedule(creditID int64) ([]*models.PaymentSchedule, error) {
	schedule, err := s.creditRepo.GetPaymentSchedule(creditID)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

// GetOverduePayments retrieves all overdue payments
func (s *CreditService) GetOverduePayments() ([]*models.PaymentSchedule, error) {
	payments, err := s.creditRepo.GetOverduePayments()
	if err != nil {
		return nil, err
	}

	return payments, nil
}
