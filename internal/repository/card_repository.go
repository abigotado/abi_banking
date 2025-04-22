package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Abigotado/abi_banking/internal/models"
	"github.com/sirupsen/logrus"
)

// CardRepository handles database operations for cards
type CardRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewCardRepository creates a new CardRepository instance
func NewCardRepository(db *sql.DB, logger *logrus.Logger) *CardRepository {
	return &CardRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new card in the database
func (r *CardRepository) Create(card *models.Card) error {
	query := `
		INSERT INTO cards (
			user_id, account_id, card_number, expiry_date, cvv,
			card_type, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		card.UserID,
		card.AccountID,
		card.CardNumber,
		card.ExpiryDate,
		card.CVV,
		card.CardType,
		card.Status,
		time.Now(),
		time.Now(),
	).Scan(&card.ID)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create card")
		return err
	}

	return nil
}

// GetByID retrieves a card by its ID
func (r *CardRepository) GetByID(id int64) (*models.Card, error) {
	query := `
		SELECT id, user_id, account_id, card_number, expiry_date, cvv,
		       card_type, status, created_at, updated_at
		FROM cards
		WHERE id = $1
	`

	card := &models.Card{}
	err := r.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.UserID,
		&card.AccountID,
		&card.CardNumber,
		&card.ExpiryDate,
		&card.CVV,
		&card.CardType,
		&card.Status,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.WithError(err).Error("Failed to get card by ID")
		return nil, err
	}

	return card, nil
}

// GetByUserID retrieves all cards for a user
func (r *CardRepository) GetByUserID(userID int64) ([]*models.Card, error) {
	query := `
		SELECT id, user_id, account_id, card_number, expiry_date, cvv,
		       card_type, status, created_at, updated_at
		FROM cards
		WHERE user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get cards by user ID")
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		card := &models.Card{}
		err := rows.Scan(
			&card.ID,
			&card.UserID,
			&card.AccountID,
			&card.CardNumber,
			&card.ExpiryDate,
			&card.CVV,
			&card.CardType,
			&card.Status,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan card row")
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// UpdateStatus updates a card's status
func (r *CardRepository) UpdateStatus(id int64, status string) error {
	query := `
		UPDATE cards
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		r.logger.WithError(err).Error("Failed to update card status")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return errors.New("card not found")
	}

	return nil
}

// Delete deletes a card by its ID
func (r *CardRepository) Delete(id int64) error {
	query := `DELETE FROM cards WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete card")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return errors.New("card not found")
	}

	return nil
}
