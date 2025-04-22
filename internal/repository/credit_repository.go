package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Abigotado/abi_banking/internal/database"
	"github.com/Abigotado/abi_banking/internal/models"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository() *CreditRepository {
	return &CreditRepository{
		db: database.DB,
	}
}

func (r *CreditRepository) Create(credit *models.Credit) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert credit
	query := `
		INSERT INTO credits (
			user_id, account_id, amount, interest_rate,
			term_months, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	err = tx.QueryRow(
		query,
		credit.UserID,
		credit.AccountID,
		credit.Amount,
		credit.InterestRate,
		credit.TermMonths,
		credit.Status,
	).Scan(&credit.ID)

	if err != nil {
		return err
	}

	// Generate and insert payment schedule
	schedule := models.GeneratePaymentSchedule(credit, time.Now())
	for _, payment := range schedule {
		query := `
			INSERT INTO payment_schedules (
				credit_id, amount, due_date, status
			)
			VALUES ($1, $2, $3, $4)
		`

		_, err := tx.Exec(
			query,
			credit.ID,
			payment.Amount,
			payment.DueDate,
			payment.Status,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *CreditRepository) GetByID(id int64) (*models.Credit, error) {
	credit := &models.Credit{}
	query := `
		SELECT id, user_id, account_id, amount, interest_rate,
			term_months, status, created_at, updated_at
		FROM credits
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&credit.ID,
		&credit.UserID,
		&credit.AccountID,
		&credit.Amount,
		&credit.InterestRate,
		&credit.TermMonths,
		&credit.Status,
		&credit.CreatedAt,
		&credit.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("credit not found")
		}
		return nil, err
	}

	return credit, nil
}

func (r *CreditRepository) GetByUserID(userID int64) ([]*models.Credit, error) {
	query := `
		SELECT id, user_id, account_id, amount, interest_rate,
			term_months, status, created_at, updated_at
		FROM credits
		WHERE user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []*models.Credit
	for rows.Next() {
		credit := &models.Credit{}
		err := rows.Scan(
			&credit.ID,
			&credit.UserID,
			&credit.AccountID,
			&credit.Amount,
			&credit.InterestRate,
			&credit.TermMonths,
			&credit.Status,
			&credit.CreatedAt,
			&credit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		credits = append(credits, credit)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return credits, nil
}

func (r *CreditRepository) GetPaymentSchedule(creditID int64) ([]*models.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, amount, due_date, status, created_at, updated_at
		FROM payment_schedules
		WHERE credit_id = $1
		ORDER BY due_date ASC
	`

	rows, err := r.db.Query(query, creditID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment schedule: %w", err)
	}
	defer rows.Close()

	var payments []*models.PaymentSchedule
	for rows.Next() {
		payment := &models.PaymentSchedule{}
		err := rows.Scan(
			&payment.ID,
			&payment.CreditID,
			&payment.Amount,
			&payment.DueDate,
			&payment.Status,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment schedule: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *CreditRepository) GetOverduePayments() ([]*models.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, amount, due_date, status
		FROM payment_schedules
		WHERE status = 'PENDING'
		AND due_date < CURRENT_TIMESTAMP
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.PaymentSchedule
	for rows.Next() {
		payment := &models.PaymentSchedule{}
		err := rows.Scan(
			&payment.ID,
			&payment.CreditID,
			&payment.Amount,
			&payment.DueDate,
			&payment.Status,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *CreditRepository) UpdateRemainingAmount(creditID int64, amount float64) error {
	query := `
		UPDATE credits
		SET remaining_amount = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(query, amount, creditID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("credit not found")
	}

	return nil
}

func (r *CreditRepository) BeginTransaction() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *CreditRepository) UpdatePaymentSchedule(payment *models.PaymentSchedule) error {
	query := `
		UPDATE payment_schedules
		SET status = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(query, payment.Status, payment.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("payment schedule not found")
	}

	return nil
}

func (r *CreditRepository) Update(credit *models.Credit) error {
	query := `
		UPDATE credits
		SET status = $1,
			remaining_amount = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := r.db.Exec(query, credit.Status, credit.RemainingAmount, credit.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("credit not found")
	}

	return nil
}

func (r *CreditRepository) CreatePaymentSchedule(payment *models.PaymentSchedule) error {
	query := `
		INSERT INTO payment_schedules (credit_id, amount, due_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		payment.CreditID,
		payment.Amount,
		payment.DueDate,
		payment.Status,
	).Scan(&payment.ID)
	if err != nil {
		return fmt.Errorf("failed to create payment schedule: %w", err)
	}

	return nil
}

// GetCreditsWithDuePayments retrieves all active credits with due payments
func (r *CreditRepository) GetCreditsWithDuePayments() ([]*models.Credit, error) {
	query := `
		SELECT c.id, c.user_id, c.account_id, c.amount, c.remaining_amount, c.interest_rate, 
			c.term_months, c.status, c.created_at, c.updated_at
		FROM credits c
		JOIN payment_schedules ps ON c.id = ps.credit_id
		WHERE c.status = $1 AND ps.status = $2 AND ps.due_date <= CURRENT_DATE
		GROUP BY c.id
	`

	rows, err := r.db.Query(query, models.CreditStatusActive, models.PaymentStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to query credits: %w", err)
	}
	defer rows.Close()

	var credits []*models.Credit
	for rows.Next() {
		credit := &models.Credit{}
		err := rows.Scan(
			&credit.ID, &credit.UserID, &credit.AccountID, &credit.Amount, &credit.RemainingAmount,
			&credit.InterestRate, &credit.TermMonths, &credit.Status, &credit.CreatedAt, &credit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credit: %w", err)
		}
		credits = append(credits, credit)
	}

	return credits, nil
}

// GetNextPayment retrieves the next due payment for a credit
func (r *CreditRepository) GetNextPayment(creditID int64) (*models.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, amount, due_date, status, created_at, updated_at
		FROM payment_schedules
		WHERE credit_id = $1 AND status = $2 AND due_date <= CURRENT_DATE
		ORDER BY due_date ASC
		LIMIT 1
	`

	payment := &models.PaymentSchedule{}
	err := r.db.QueryRow(query, creditID, models.PaymentStatusPending).Scan(
		&payment.ID, &payment.CreditID, &payment.Amount, &payment.DueDate,
		&payment.Status, &payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get next payment: %w", err)
	}

	return payment, nil
}

func (r *CreditRepository) UpdatePaymentStatus(paymentID int64, status string) error {
	query := `
		UPDATE payment_schedules
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(query, status, paymentID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}
