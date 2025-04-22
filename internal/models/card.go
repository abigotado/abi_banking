package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Card struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	AccountID  int64     `json:"account_id"`
	CardNumber string    `json:"card_number"`
	ExpiryDate string    `json:"expiry_date"`
	CVV        string    `json:"-"`
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
