package repository

import (
	"database/sql"
	"errors"

	"github.com/Abigotado/abi_banking/internal/database"
	"github.com/Abigotado/abi_banking/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository() *CardRepository {
	return &CardRepository{
		db: database.DB,
	}
}

func (r *CardRepository) Create(card *models.Card, pgpKey string) error {
	// Encrypt card number and expiry date using PGP
	encryptedNumber, err := encryptWithPGP(card.CardNumber, pgpKey)
	if err != nil {
		return err
	}

	encryptedExpiry, err := encryptWithPGP(card.ExpiryDate, pgpKey)
	if err != nil {
		return err
	}

	// Hash CVV
	cvvHash, err := bcrypt.GenerateFromPassword([]byte(card.CVV), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO cards (
			user_id, account_id, card_number, expiry_date,
			cvv_hash, card_type, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	err = r.db.QueryRow(
		query,
		card.UserID,
		card.AccountID,
		encryptedNumber,
		encryptedExpiry,
		string(cvvHash),
		card.CardType,
		card.Status,
	).Scan(&card.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *CardRepository) GetByID(id int64, pgpKey string) (*models.Card, error) {
	card := &models.Card{}
	var encryptedNumber, encryptedExpiry []byte
	var cvvHash string

	query := `
		SELECT id, user_id, account_id, card_number, expiry_date,
			cvv_hash, card_type, status, created_at, updated_at
		FROM cards
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.UserID,
		&card.AccountID,
		&encryptedNumber,
		&encryptedExpiry,
		&cvvHash,
		&card.CardType,
		&card.Status,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("card not found")
		}
		return nil, err
	}

	// Decrypt card data
	cardNumber, err := decryptWithPGP(encryptedNumber, pgpKey)
	if err != nil {
		return nil, err
	}
	card.CardNumber = cardNumber

	expiryDate, err := decryptWithPGP(encryptedExpiry, pgpKey)
	if err != nil {
		return nil, err
	}
	card.ExpiryDate = expiryDate

	return card, nil
}

func (r *CardRepository) GetByUserID(userID int64, pgpKey string) ([]*models.Card, error) {
	query := `
		SELECT id, user_id, account_id, card_number, expiry_date,
			cvv_hash, card_type, status, created_at, updated_at
		FROM cards
		WHERE user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		card := &models.Card{}
		var encryptedNumber, encryptedExpiry []byte
		var cvvHash string

		err := rows.Scan(
			&card.ID,
			&card.UserID,
			&card.AccountID,
			&encryptedNumber,
			&encryptedExpiry,
			&cvvHash,
			&card.CardType,
			&card.Status,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Decrypt card data
		cardNumber, err := decryptWithPGP(encryptedNumber, pgpKey)
		if err != nil {
			return nil, err
		}
		card.CardNumber = cardNumber

		expiryDate, err := decryptWithPGP(encryptedExpiry, pgpKey)
		if err != nil {
			return nil, err
		}
		card.ExpiryDate = expiryDate

		cards = append(cards, card)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

func (r *CardRepository) UpdateStatus(id int64, status string) error {
	query := `
		UPDATE cards
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("card not found")
	}

	return nil
}

// Helper functions for PGP encryption/decryption
func encryptWithPGP(data, key string) ([]byte, error) {
	// TODO: Implement PGP encryption
	return []byte(data), nil
}

func decryptWithPGP(data []byte, key string) (string, error) {
	// TODO: Implement PGP decryption
	return string(data), nil
}
