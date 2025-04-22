package service

import (
	"time"

	"errors"

	"github.com/Abigotado/abi_banking/internal/models"
	"github.com/Abigotado/abi_banking/internal/repository"
	"github.com/sirupsen/logrus"
)

// CreditService handles business logic for credit operations
type CreditService struct {
	creditRepo *repository.CreditRepository
	logger     *logrus.Logger
}

// NewCreditService creates a new CreditService instance
func NewCreditService(creditRepo *repository.CreditRepository, logger *logrus.Logger) *CreditService {
	return &CreditService{
		creditRepo: creditRepo,
		logger:     logger,
	}
}

// CreditAnalytics represents credit analytics data
type CreditAnalytics struct {
	TotalCredits      int            `json:"total_credits"`
	TotalAmount       float64        `json:"total_amount"`
	TotalPaid         float64        `json:"total_paid"`
	TotalRemaining    float64        `json:"total_remaining"`
	AverageInterest   float64        `json:"average_interest"`
	CreditsByStatus   map[string]int `json:"credits_by_status"`
	NextPaymentDate   *time.Time     `json:"next_payment_date"`
	NextPaymentAmount float64        `json:"next_payment_amount"`
}

// GetCreditAnalytics retrieves credit analytics for a user
func (s *CreditService) GetCreditAnalytics(userID int64) (*CreditAnalytics, error) {
	// Get user credits
	credits, err := s.creditRepo.GetByUserID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user credits")
		return nil, err
	}

	// Calculate analytics
	var totalCredits int
	var totalAmount float64
	var totalPaid float64
	var totalRemaining float64
	var totalInterest float64
	creditsByStatus := make(map[string]int)
	var nextPaymentDate *time.Time
	var nextPaymentAmount float64

	for _, credit := range credits {
		totalCredits++
		totalAmount += credit.Amount
		totalInterest += credit.InterestRate
		creditsByStatus[credit.Status]++

		// Get payment schedule for the credit
		schedule, err := s.creditRepo.GetPaymentSchedule(credit.ID)
		if err != nil {
			s.logger.WithError(err).Error("Failed to get payment schedule")
			return nil, err
		}

		// Calculate paid and remaining amounts
		for _, payment := range schedule {
			if payment.Status == "PAID" {
				totalPaid += payment.Amount
			} else {
				totalRemaining += payment.Amount
			}

			// Find the next payment
			if payment.Status == "PENDING" && (nextPaymentDate == nil || payment.PaymentDate.Before(*nextPaymentDate)) {
				nextPaymentDate = &payment.PaymentDate
				nextPaymentAmount = payment.Amount
			}
		}
	}

	// Calculate average interest
	var averageInterest float64
	if totalCredits > 0 {
		averageInterest = totalInterest / float64(totalCredits)
	}

	return &CreditAnalytics{
		TotalCredits:      totalCredits,
		TotalAmount:       totalAmount,
		TotalPaid:         totalPaid,
		TotalRemaining:    totalRemaining,
		AverageInterest:   averageInterest,
		CreditsByStatus:   creditsByStatus,
		NextPaymentDate:   nextPaymentDate,
		NextPaymentAmount: nextPaymentAmount,
	}, nil
}

// CreateCredit creates a new credit for a user
func (s *CreditService) CreateCredit(req *models.CreateCreditRequest) (*models.Credit, error) {
	// Validate input
	if req.Amount <= 0 {
		return nil, errors.New("invalid credit amount")
	}
	if req.TermMonths <= 0 {
		return nil, errors.New("invalid credit term")
	}
	if req.InterestRate <= 0 {
		return nil, errors.New("invalid interest rate")
	}

	// Create credit
	credit := &models.Credit{
		UserID:          req.UserID,
		AccountID:       req.AccountID,
		Amount:          req.Amount,
		TermMonths:      req.TermMonths,
		InterestRate:    req.InterestRate,
		Status:          "ACTIVE",
		RemainingAmount: req.Amount,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save credit to database
	err := s.creditRepo.Create(credit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create credit")
		return nil, err
	}

	return credit, nil
}

// GetCreditByID retrieves a credit by its ID
func (s *CreditService) GetCreditByID(creditID int64) (*models.Credit, error) {
	credit, err := s.creditRepo.GetByID(creditID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get credit by ID")
		return nil, err
	}
	return credit, nil
}

// GetCreditsByUserID retrieves all credits for a user
func (s *CreditService) GetCreditsByUserID(userID int64) ([]*models.Credit, error) {
	credits, err := s.creditRepo.GetByUserID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user credits")
		return nil, err
	}
	return credits, nil
}

// PayCredit processes a credit payment
func (s *CreditService) PayCredit(creditID int64, req *models.PayCreditRequest) error {
	// Get credit
	credit, err := s.creditRepo.GetByID(creditID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get credit")
		return err
	}

	// Validate payment amount
	if req.Amount <= 0 {
		return errors.New("invalid payment amount")
	}
	if req.Amount > credit.RemainingAmount {
		return errors.New("payment amount exceeds remaining credit amount")
	}

	// Update remaining amount
	newRemainingAmount := credit.RemainingAmount - req.Amount
	err = s.creditRepo.UpdateRemainingAmount(creditID, newRemainingAmount)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update credit remaining amount")
		return err
	}

	// Update payment schedule
	schedule, err := s.creditRepo.GetPaymentSchedule(creditID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get payment schedule")
		return err
	}

	// Find and update the next pending payment
	for _, payment := range schedule {
		if payment.Status == "PENDING" {
			if req.Amount >= payment.Amount {
				// Full payment
				err = s.creditRepo.UpdatePaymentStatus(payment.ID, "PAID")
				if err != nil {
					s.logger.WithError(err).Error("Failed to update payment status")
					return err
				}
				req.Amount -= payment.Amount
			} else {
				// Partial payment - update the payment amount
				err = s.creditRepo.UpdatePaymentStatus(payment.ID, "PARTIAL")
				if err != nil {
					s.logger.WithError(err).Error("Failed to update payment status")
					return err
				}
				break
			}
		}
	}

	return nil
}
