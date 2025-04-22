package models

import (
	"time"
)

// Account represents a bank account
type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id" validate:"required"`
	Balance   float64   `json:"balance" validate:"gte=0"`
	Currency  string    `json:"currency" validate:"required,len=3"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID            int64     `json:"id"`
	FromAccountID int64     `json:"from_account_id" validate:"required"`
	ToAccountID   int64     `json:"to_account_id" validate:"required"`
	Amount        float64   `json:"amount" validate:"required,gt=0"`
	Type          string    `json:"type" validate:"required,oneof=transfer deposit withdrawal"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateAccountRequest represents a request to create a new account
type CreateAccountRequest struct {
	UserID   int64   `json:"user_id" validate:"required"`
	Currency string  `json:"currency" validate:"required,len=3"`
	Balance  float64 `json:"balance" validate:"gte=0"`
}

// TransferRequest represents a money transfer request
type TransferRequest struct {
	FromAccountID int64   `json:"from_account_id" validate:"required"`
	ToAccountID   int64   `json:"to_account_id" validate:"required,nefield=FromAccountID"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
}
