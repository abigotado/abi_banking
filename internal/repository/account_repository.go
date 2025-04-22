package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Abigotado/abi_banking/internal/models"
	"github.com/sirupsen/logrus"
)

type AccountRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewAccountRepository() *AccountRepository {
	return &AccountRepository{
		logger: logrus.New(),
	}
}

func (r *AccountRepository) SetDB(db *sql.DB) {
	r.db = db
}

func (r *AccountRepository) BeginTransaction() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *AccountRepository) Create(account *models.Account) error {
	query := `
		INSERT INTO accounts (user_id, balance, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		account.UserID,
		account.Balance,
		account.Currency,
		account.CreatedAt,
		account.UpdatedAt,
	).Scan(&account.ID)
}

func (r *AccountRepository) GetByID(id int64) (*models.Account, error) {
	account := &models.Account{}
	query := `
		SELECT id, user_id, balance, currency, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	return account, nil
}

func (r *AccountRepository) GetByUserID(userID int64) ([]*models.Account, error) {
	query := `
		SELECT id, user_id, balance, currency, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		account := &models.Account{}
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Balance,
			&account.Currency,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (r *AccountRepository) UpdateBalance(id int64, newBalance float64) error {
	query := `
		UPDATE accounts
		SET balance = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(query, newBalance, time.Now(), id)
	return err
}

func (r *AccountRepository) CreateTransaction(transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (from_account_id, to_account_id, amount, type, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		transaction.FromAccountID,
		transaction.ToAccountID,
		transaction.Amount,
		transaction.Type,
		transaction.CreatedAt,
	).Scan(&transaction.ID)
}

// GetTransactions retrieves transactions for an account within a date range
func (r *AccountRepository) GetTransactions(accountID int64, startDate, endDate time.Time) ([]*models.Transaction, error) {
	query := `
		SELECT id, from_account_id, to_account_id, amount, type, created_at
		FROM transactions
		WHERE (from_account_id = $1 OR to_account_id = $1)
		AND created_at >= $2
		AND created_at <= $3
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, accountID, startDate, endDate)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get transactions")
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		tx := &models.Transaction{}
		err := rows.Scan(
			&tx.ID,
			&tx.FromAccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Type,
			&tx.CreatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan transaction")
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}
