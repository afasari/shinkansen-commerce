package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID            uuid.UUID
	OrderID       uuid.UUID
	Method        string
	AmountMinor   int
	Currency      string
	Status        string
	TransactionID *string
	PaymentData   []byte
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreatePaymentParams struct {
	OrderID     uuid.UUID
	Method      string
	AmountMinor int
	Currency    string
}

type UpdatePaymentStatusParams struct {
	ID            uuid.UUID
	Status        string
	TransactionID *string
}

type UpdatePaymentDataParams struct {
	ID          uuid.UUID
	PaymentData []byte
}

type Querier interface {
	CreatePayment(ctx context.Context, arg CreatePaymentParams) (uuid.UUID, error)
	GetPayment(ctx context.Context, id uuid.UUID) (Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (Payment, error)
	UpdatePaymentStatus(ctx context.Context, arg UpdatePaymentStatusParams) error
	UpdatePaymentData(ctx context.Context, arg UpdatePaymentDataParams) error
	ListPaymentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]Payment, error)
}

type Queries struct {
	db *DB
}

func NewQueries(db *DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) CreatePayment(ctx context.Context, arg CreatePaymentParams) (uuid.UUID, error) {
	const sql = `
		INSERT INTO payments.payments (order_id, method, amount_minor, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'PAYMENT_STATUS_PENDING', NOW(), NOW())
		RETURNING id
	`
	row := q.db.pool.QueryRow(ctx, sql, arg.OrderID, arg.Method, arg.AmountMinor, arg.Currency)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

func (q *Queries) GetPayment(ctx context.Context, id uuid.UUID) (Payment, error) {
	const sql = `
		SELECT id, order_id, method, amount_minor, currency, status, transaction_id, payment_data, created_at, updated_at
		FROM payments.payments
		WHERE id = $1
	`
	row := q.db.pool.QueryRow(ctx, sql, id)
	var p Payment
	err := row.Scan(
		&p.ID, &p.OrderID, &p.Method, &p.AmountMinor, &p.Currency, &p.Status,
		&p.TransactionID, &p.PaymentData, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

func (q *Queries) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (Payment, error) {
	const sql = `
		SELECT id, order_id, method, amount_minor, currency, status, transaction_id, payment_data, created_at, updated_at
		FROM payments.payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := q.db.pool.QueryRow(ctx, sql, orderID)
	var p Payment
	err := row.Scan(
		&p.ID, &p.OrderID, &p.Method, &p.AmountMinor, &p.Currency, &p.Status,
		&p.TransactionID, &p.PaymentData, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

func (q *Queries) UpdatePaymentStatus(ctx context.Context, arg UpdatePaymentStatusParams) error {
	const sql = `
		UPDATE payments.payments
		SET
			status = $2,
			transaction_id = COALESCE($3, transaction_id),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.db.pool.Exec(ctx, sql, arg.ID, arg.Status, arg.TransactionID)
	return err
}

func (q *Queries) UpdatePaymentData(ctx context.Context, arg UpdatePaymentDataParams) error {
	const sql = `
		UPDATE payments.payments
		SET
			payment_data = $2,
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.db.pool.Exec(ctx, sql, arg.ID, arg.PaymentData)
	return err
}

func (q *Queries) ListPaymentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]Payment, error) {
	const sql = `
		SELECT id, order_id, method, amount_minor, currency, status, transaction_id, payment_data, created_at, updated_at
		FROM payments.payments
		WHERE order_id = $1
		ORDER BY created_at DESC
	`
	rows, err := q.db.pool.Query(ctx, sql, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		err := rows.Scan(
			&p.ID, &p.OrderID, &p.Method, &p.AmountMinor, &p.Currency, &p.Status,
			&p.TransactionID, &p.PaymentData, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}
