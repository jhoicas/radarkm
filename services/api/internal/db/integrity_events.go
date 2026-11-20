package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IntegrityEvent struct {
	ID        string
	DeviceID  *string
	Verdict   string
	RawClaims json.RawMessage
	CreatedAt time.Time
}

type IntegrityEventRepository struct {
	pool *pgxpool.Pool
}

func NewIntegrityEventRepository(pool *pgxpool.Pool) *IntegrityEventRepository {
	return &IntegrityEventRepository{pool: pool}
}

func (r *IntegrityEventRepository) Insert(ctx context.Context, e IntegrityEvent) (IntegrityEvent, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO integrity_events (device_id, verdict, raw_claims)
		VALUES ($1, $2, COALESCE($3, '{}'::jsonb))
		RETURNING id::text, device_id::text, verdict, raw_claims, created_at`,
		e.DeviceID, e.Verdict, nullJSON(e.RawClaims),
	)
	var out IntegrityEvent
	err := row.Scan(&out.ID, &out.DeviceID, &out.Verdict, &out.RawClaims, &out.CreatedAt)
	return out, err
}

func (r *IntegrityEventRepository) ListByDevice(ctx context.Context, deviceID string, limit int) ([]IntegrityEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, device_id::text, verdict, raw_claims, created_at
		FROM integrity_events
		WHERE device_id = $1::uuid
		ORDER BY created_at DESC
		LIMIT $2`, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []IntegrityEvent
	for rows.Next() {
		var e IntegrityEvent
		if err := rows.Scan(&e.ID, &e.DeviceID, &e.Verdict, &e.RawClaims, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func nullJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
