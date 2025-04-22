package service

import (
	"math"
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
			if payment.Status == "PENDING" && (nextPaymentDate == nil || payment.DueDate.Before(*nextPaymentDate)) {
				nextPaymentDate = &payment.DueDate
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

// CreateCredit creates a new credit
func (s *CreditService) CreateCredit(userID int64, amount float64, termMonths int, interestRate float64) (*models.Credit, error) {
	// Create credit record
	credit := &models.Credit{
		UserID:          userID,
		Amount:          amount,
		RemainingAmount: amount,
		TermMonths:      termMonths,
		InterestRate:    interestRate,
		Status:          string(models.CreditStatusActive),
	}

	// Start transaction
	tx, err := s.creditRepo.BeginTransaction()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create credit
	if err := s.creditRepo.Create(credit); err != nil {
		return nil, err
	}

	// Generate payment schedule
	schedule, err := s.GeneratePaymentSchedule(credit)
	if err != nil {
		return nil, err
	}

	// Save payment schedule
	for _, payment := range schedule {
		if err := s.creditRepo.CreatePaymentSchedule(payment); err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
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

// GeneratePaymentSchedule generates a payment schedule for a credit
func (s *CreditService) GeneratePaymentSchedule(credit *models.Credit) ([]*models.PaymentSchedule, error) {
	// Calculate monthly payment using annuity formula
	monthlyRate := credit.InterestRate / 12 / 100
	monthlyPayment := credit.Amount * (monthlyRate * math.Pow(1+monthlyRate, float64(credit.TermMonths))) / (math.Pow(1+monthlyRate, float64(credit.TermMonths)) - 1)

	// Generate schedule
	var schedule []*models.PaymentSchedule
	remainingAmount := credit.Amount
	dueDate := time.Now().AddDate(0, 1, 0) // First payment due in 1 month

	for i := 0; i < credit.TermMonths; i++ {
		// Calculate interest for this period
		interest := remainingAmount * monthlyRate
		principal := monthlyPayment - interest

		// Create payment entry
		payment := &models.PaymentSchedule{
			CreditID: credit.ID,
			Amount:   monthlyPayment,
			DueDate:  dueDate,
			Status:   models.PaymentStatusPending,
		}

		// Add to schedule
		schedule = append(schedule, payment)

		// Update for next period
		remainingAmount -= principal
		dueDate = dueDate.AddDate(0, 1, 0)
	}

	return schedule, nil
}
