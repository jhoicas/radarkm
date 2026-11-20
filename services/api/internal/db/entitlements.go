package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Entitlement struct {
	ID        string
	DeviceID  string
	UserID    *string
	PlanID    *string
	Status    string
	StartsAt  time.Time
	EndsAt    *time.Time
	CreatedAt time.Time
}

type EntitlementRepository struct {
	pool *pgxpool.Pool
}

func NewEntitlementRepository(pool *pgxpool.Pool) *EntitlementRepository {
	return &EntitlementRepository{pool: pool}
}

func (r *EntitlementRepository) Insert(ctx context.Context, e Entitlement) (Entitlement, error) {
	var startsAt any
	if e.StartsAt.IsZero() {
		startsAt = nil
	} else {
		startsAt = e.StartsAt.UTC()
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO entitlements (device_id, user_id, plan_id, status, starts_at, ends_at)
		VALUES ($1::uuid, $2, $3, $4, COALESCE($5, NOW()), $6)
		RETURNING id::text, device_id::text, user_id::text, plan_id::text,
		          status, starts_at, ends_at, created_at`,
		e.DeviceID, e.UserID, e.PlanID, defaultEntitlementStatus(e.Status), startsAt, e.EndsAt,
	)
	var out Entitlement
	err := row.Scan(
		&out.ID, &out.DeviceID, &out.UserID, &out.PlanID,
		&out.Status, &out.StartsAt, &out.EndsAt, &out.CreatedAt,
	)
	return out, err
}

// HasActive returns true when device (or its user) has a non-expired active entitlement at now.
func (r *EntitlementRepository) HasActive(ctx context.Context, deviceID string, userID *string, now time.Time) (bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM entitlements e
			WHERE e.status = 'active'
			  AND e.starts_at <= $3
			  AND (e.ends_at IS NULL OR e.ends_at > $3)
			  AND (
			        e.device_id = $1::uuid
			        OR ($2::uuid IS NOT NULL AND e.user_id = $2::uuid)
			      )
		)`, deviceID, userID, now.UTC())
	var ok bool
	err := row.Scan(&ok)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return ok, err
}

func defaultEntitlementStatus(s string) string {
	if s == "" {
		return "active"
	}
	return s
}
