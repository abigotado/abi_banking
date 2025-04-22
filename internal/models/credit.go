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

type CreateCreditRequest struct {
	UserID       int64   `json:"user_id" validate:"required"`
	AccountID    int64   `json:"account_id" validate:"required"`
	Amount       float64 `json:"amount" validate:"required,gt=0"`
	InterestRate float64 `json:"interest_rate" validate:"required,gt=0"`
	TermMonths   int     `json:"term_months" validate:"required,gt=0"`
}

type PayCreditRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

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
			CreditID:      credit.ID,
			PaymentNumber: i + 1,
			PaymentDate:   startDate.AddDate(0, i, 0),
			Amount:        monthlyPayment,
			Principal:     principal,
			Interest:      interest,
			Status:        "PENDING",
		}
	}

	return schedule
}
