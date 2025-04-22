package models

import (
	"math"
	"time"
)

type Credit struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	AccountID       int64     `json:"account_id"`
	Amount          float64   `json:"amount"`
	RemainingAmount float64   `json:"remaining_amount"`
	InterestRate    float64   `json:"interest_rate"`
	TermMonths      int       `json:"term_months"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateCreditRequest represents a request to create a credit
type CreateCreditRequest struct {
	UserID       int64   `json:"user_id" validate:"required"`
	AccountID    int64   `json:"account_id" validate:"required"`
	Amount       float64 `json:"amount" validate:"required,gt=0"`
	TermMonths   int     `json:"term_months" validate:"required,gt=0"`
	InterestRate float64 `json:"interest_rate" validate:"required,gt=0"`
}

type PayCreditRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

// CreditStatus represents the status of a credit
type CreditStatus string

const (
	CreditStatusActive  CreditStatus = "active"
	CreditStatusPaid    CreditStatus = "paid"
	CreditStatusDefault CreditStatus = "default"
	CreditStatusClosed  CreditStatus = "closed"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusLate    PaymentStatus = "late"
)

// PaymentSchedule represents a scheduled payment for a credit
type PaymentSchedule struct {
	ID        int64         `json:"id"`
	CreditID  int64         `json:"credit_id"`
	Amount    float64       `json:"amount"`
	DueDate   time.Time     `json:"due_date"`
	Status    PaymentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func CalculateAnnuityPayment(amount float64, annualRate float64, termMonths int) float64 {
	monthlyRate := annualRate / 12 / 100
	denominator := 1 - math.Pow(1+monthlyRate, float64(-termMonths))
	return amount * (monthlyRate / denominator)
}

func GeneratePaymentSchedule(credit *Credit, startDate time.Time) []PaymentSchedule {
	monthlyPayment := CalculateAnnuityPayment(credit.Amount, credit.InterestRate, credit.TermMonths)
	remainingPrincipal := credit.Amount
	schedule := make([]PaymentSchedule, credit.TermMonths)

	for i := 0; i < credit.TermMonths; i++ {
		interest := remainingPrincipal * (credit.InterestRate / 12 / 100)
		principal := monthlyPayment - interest
		remainingPrincipal -= principal

		schedule[i] = PaymentSchedule{
			CreditID: credit.ID,
			Amount:   monthlyPayment,
			DueDate:  startDate.AddDate(0, i, 0),
			Status:   PaymentStatusPending,
		}
	}

	return schedule
}
