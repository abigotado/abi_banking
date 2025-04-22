package models

import (
	"time"
)

// Credit represents a credit account
type Credit struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id" validate:"required"`
	AccountID       int64     `json:"account_id" validate:"required"`
	Amount          float64   `json:"amount" validate:"required,gt=0"`
	InterestRate    float64   `json:"interest_rate" validate:"required,gt=0"`
	TermMonths      int       `json:"term_months" validate:"required,gt=0"`
	RemainingAmount float64   `json:"remaining_amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PaymentSchedule represents a scheduled payment for a credit
type PaymentSchedule struct {
	ID            int64     `json:"id"`
	CreditID      int64     `json:"credit_id"`
	PaymentNumber int       `json:"payment_number"`
	PaymentDate   time.Time `json:"payment_date"`
	Amount        float64   `json:"amount"`
	Principal     float64   `json:"principal"`
	Interest      float64   `json:"interest"`
	Status        string    `json:"status"`
}

// CreateCreditRequest represents a request to create a new credit
type CreateCreditRequest struct {
	UserID       int64   `json:"user_id" validate:"required"`
	AccountID    int64   `json:"account_id" validate:"required"`
	Amount       float64 `json:"amount" validate:"required,gt=0"`
	InterestRate float64 `json:"interest_rate" validate:"required,gt=0"`
	TermMonths   int     `json:"term_months" validate:"required,gt=0"`
}

// PayCreditRequest represents a request to make a payment towards a credit
type PayCreditRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

// GeneratePaymentSchedule generates a payment schedule for a credit
func GeneratePaymentSchedule(credit *Credit, startDate time.Time) []*PaymentSchedule {
	monthlyInterestRate := credit.InterestRate / 12 / 100
	monthlyPayment := calculateMonthlyPayment(credit.Amount, monthlyInterestRate, credit.TermMonths)

	schedule := make([]*PaymentSchedule, credit.TermMonths)
	remainingPrincipal := credit.Amount

	for i := 0; i < credit.TermMonths; i++ {
		monthlyInterest := remainingPrincipal * monthlyInterestRate
		monthlyPrincipal := monthlyPayment - monthlyInterest

		if i == credit.TermMonths-1 {
			// Last payment - adjust for rounding errors
			monthlyPayment = remainingPrincipal * (1 + monthlyInterestRate)
			monthlyPrincipal = remainingPrincipal
			monthlyInterest = monthlyPayment - monthlyPrincipal
		}

		schedule[i] = &PaymentSchedule{
			PaymentNumber: i + 1,
			PaymentDate:   startDate.AddDate(0, i+1, 0),
			Amount:        monthlyPayment,
			Principal:     monthlyPrincipal,
			Interest:      monthlyInterest,
			Status:        "PENDING",
		}

		remainingPrincipal -= monthlyPrincipal
	}

	return schedule
}

// calculateMonthlyPayment calculates the monthly payment amount using the PMT formula
func calculateMonthlyPayment(principal, monthlyInterestRate float64, termMonths int) float64 {
	if monthlyInterestRate == 0 {
		return principal / float64(termMonths)
	}

	pow := 1.0
	for i := 0; i < termMonths; i++ {
		pow *= (1 + monthlyInterestRate)
	}

	return principal * monthlyInterestRate * pow / (pow - 1)
}
