package service

import (
	"time"

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
