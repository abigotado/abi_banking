package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	CardStatusActive  = "active"
	CardStatusBlocked = "blocked"
)

// Card represents a bank card
type Card struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id" validate:"required"`
	AccountID  int64     `json:"account_id" validate:"required"`
	CardNumber string    `json:"card_number" validate:"required,len=16"`
	ExpiryDate string    `json:"expiry_date" validate:"required,len=5"`
	CVV        string    `json:"-"` // Never exposed in JSON
	CardType   string    `json:"card_type" validate:"required,oneof=debit credit"`
	Status     string    `json:"status" validate:"required,oneof=active blocked"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateCardRequest represents a request to create a new card
type CreateCardRequest struct {
	UserID    int64  `json:"user_id" validate:"required"`
	AccountID int64  `json:"account_id" validate:"required"`
	CardType  string `json:"card_type" validate:"required,oneof=debit credit"`
}

// BlockCardRequest represents a request to block a card
type BlockCardRequest struct {
	CardID int64  `json:"card_id" validate:"required"`
	Reason string `json:"reason" validate:"required"`
}

// CardResponse represents a card response with masked number
type CardResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	AccountID  int64     `json:"account_id"`
	CardNumber string    `json:"card_number"` // Masked number
	ExpiryDate string    `json:"expiry_date"`
	CardType   string    `json:"card_type"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (c *Card) GenerateHMAC(secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(c.CardNumber + c.ExpiryDate))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Card) VerifyHMAC(secret, hmac string) bool {
	expectedHMAC := c.GenerateHMAC(secret)
	return hmac == expectedHMAC
}

func LuhnCheck(number string) bool {
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// MaskNumber masks the card number, showing only first 4 and last 4 digits
func (c *Card) MaskNumber() string {
	if len(c.CardNumber) < 8 {
		return c.CardNumber
	}
	return c.CardNumber[:4] + "****" + c.CardNumber[len(c.CardNumber)-4:]
}

// ToResponse converts a Card to a CardResponse with masked number
func (c *Card) ToResponse() *CardResponse {
	return &CardResponse{
		ID:         c.ID,
		UserID:     c.UserID,
		AccountID:  c.AccountID,
		CardNumber: c.MaskNumber(),
		ExpiryDate: c.ExpiryDate,
		CardType:   c.CardType,
		Status:     c.Status,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}
