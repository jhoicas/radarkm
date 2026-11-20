package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Transaction struct {
	ID                 string
	DeviceID           *string
	UserID             *string
	PlanID             *string
	Channel            string
	Status             string
	ExternalReference  string
	AmountCop          int64
	DianInvoiceStatus  string
	CreatedAt          time.Time
}

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (r *TransactionRepository) Insert(ctx context.Context, tx Transaction) (Transaction, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO transactions (
			device_id, user_id, plan_id, channel, status,
			external_reference, amount_cop, dian_invoice_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE(NULLIF($8, ''), 'not_required'))
		RETURNING id::text, device_id::text, user_id::text, plan_id::text,
		          channel, status, external_reference, amount_cop,
		          dian_invoice_status, created_at`,
		tx.DeviceID, tx.UserID, tx.PlanID, tx.Channel, tx.Status,
		tx.ExternalReference, tx.AmountCop, tx.DianInvoiceStatus,
	)
	var out Transaction
	err := row.Scan(
		&out.ID, &out.DeviceID, &out.UserID, &out.PlanID,
		&out.Channel, &out.Status, &out.ExternalReference, &out.AmountCop,
		&out.DianInvoiceStatus, &out.CreatedAt,
	)
	return out, err
}
