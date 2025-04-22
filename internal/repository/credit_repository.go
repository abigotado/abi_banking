package repository

import (
	"database/sql"
	"errors"
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

func (r *CreditRepository) SetDB(db *sql.DB) {
	r.db = db
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

	// Set the initial remaining amount equal to the total amount
	credit.RemainingAmount = credit.Amount

	// Generate and insert payment schedule
	schedule := models.GeneratePaymentSchedule(credit, time.Now())
	for _, payment := range schedule {
		query := `
			INSERT INTO payment_schedules (
				credit_id, payment_number, payment_date,
				amount, principal, interest, status
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err := tx.Exec(
			query,
			credit.ID,
			payment.PaymentNumber,
			payment.PaymentDate,
			payment.Amount,
			payment.Principal,
			payment.Interest,
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

	// Calculate remaining amount from payment schedule
	schedule, err := r.GetPaymentSchedule(id)
	if err != nil {
		return nil, err
	}

	credit.RemainingAmount = credit.Amount
	for _, payment := range schedule {
		if payment.Status == "PAID" {
			credit.RemainingAmount -= payment.Amount
		}
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

		// Calculate remaining amount from payment schedule
		schedule, err := r.GetPaymentSchedule(credit.ID)
		if err != nil {
			return nil, err
		}

		credit.RemainingAmount = credit.Amount
		for _, payment := range schedule {
			if payment.Status == "PAID" {
				credit.RemainingAmount -= payment.Amount
			}
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
		SELECT id, credit_id, payment_number, payment_date,
			amount, principal, interest, status
		FROM payment_schedules
		WHERE credit_id = $1
		ORDER BY payment_number
	`

	rows, err := r.db.Query(query, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedule []*models.PaymentSchedule
	for rows.Next() {
		payment := &models.PaymentSchedule{}
		err := rows.Scan(
			&payment.ID,
			&payment.CreditID,
			&payment.PaymentNumber,
			&payment.PaymentDate,
			&payment.Amount,
			&payment.Principal,
			&payment.Interest,
			&payment.Status,
		)
		if err != nil {
			return nil, err
		}
		schedule = append(schedule, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return schedule, nil
}

func (r *CreditRepository) UpdatePaymentStatus(paymentID int64, status string) error {
	query := `
		UPDATE payment_schedules
		SET status = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(query, status, paymentID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("payment not found")
	}

	return nil
}

func (r *CreditRepository) GetOverduePayments() ([]*models.PaymentSchedule, error) {
	query := `
		SELECT id, credit_id, payment_number, payment_date,
			amount, principal, interest, status
		FROM payment_schedules
		WHERE status = 'PENDING'
		AND payment_date < CURRENT_TIMESTAMP
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
			&payment.PaymentNumber,
			&payment.PaymentDate,
			&payment.Amount,
			&payment.Principal,
			&payment.Interest,
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

func (r *CreditRepository) BeginTransaction() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *CreditRepository) Update(credit *models.Credit) error {
	query := `
		UPDATE credits
		SET status = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(query, credit.Status, credit.ID)
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
		INSERT INTO payment_schedules (
			credit_id, payment_number, payment_date,
			amount, principal, interest, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		payment.CreditID,
		payment.PaymentNumber,
		payment.PaymentDate,
		payment.Amount,
		payment.Principal,
		payment.Interest,
		payment.Status,
	).Scan(&payment.ID)

	return err
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
		return errors.New("payment not found")
	}

	return nil
}
