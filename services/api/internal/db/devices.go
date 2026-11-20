package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Device struct {
	ID                   string
	UserID               *string
	WidevineHash         string
	PlayIntegrityVerdict *string
	DeviceRecallBits     int64
	Status               string
	Blocked              bool
	LastSeenAt           *time.Time
	CreatedAt            time.Time
}

type DeviceRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceRepository(pool *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{pool: pool}
}

func (r *DeviceRepository) Insert(ctx context.Context, d Device) (Device, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO devices (
			user_id, widevine_hash, play_integrity_verdict,
			device_recall_bits, status, blocked, last_seen_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, user_id::text, widevine_hash, play_integrity_verdict,
		          device_recall_bits, status, blocked, last_seen_at, created_at`,
		d.UserID, d.WidevineHash, d.PlayIntegrityVerdict,
		d.DeviceRecallBits, defaultStatus(d.Status), d.Blocked, d.LastSeenAt,
	)
	var out Device
	err := row.Scan(
		&out.ID, &out.UserID, &out.WidevineHash, &out.PlayIntegrityVerdict,
		&out.DeviceRecallBits, &out.Status, &out.Blocked, &out.LastSeenAt, &out.CreatedAt,
	)
	return out, err
}

func (r *DeviceRepository) GetByWidevineHash(ctx context.Context, hash string) (Device, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, widevine_hash, play_integrity_verdict,
		       device_recall_bits, status, blocked, last_seen_at, created_at
		FROM devices WHERE widevine_hash = $1`, hash)
	var out Device
	err := row.Scan(
		&out.ID, &out.UserID, &out.WidevineHash, &out.PlayIntegrityVerdict,
		&out.DeviceRecallBits, &out.Status, &out.Blocked, &out.LastSeenAt, &out.CreatedAt,
	)
	return out, err
}

func (r *DeviceRepository) GetByID(ctx context.Context, id string) (Device, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, widevine_hash, play_integrity_verdict,
		       device_recall_bits, status, blocked, last_seen_at, created_at
		FROM devices WHERE id = $1::uuid`, id)
	var out Device
	err := row.Scan(
		&out.ID, &out.UserID, &out.WidevineHash, &out.PlayIntegrityVerdict,
		&out.DeviceRecallBits, &out.Status, &out.Blocked, &out.LastSeenAt, &out.CreatedAt,
	)
	return out, err
}

// UpdateIntegrityState persists Play Integrity / Device Recall outcome on the device row.
func (r *DeviceRepository) UpdateIntegrityState(
	ctx context.Context,
	deviceID string,
	widevineHash string,
	verdict string,
	deviceRecallBits int64,
	lastSeenAt time.Time,
) (Device, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE devices
		SET widevine_hash = $2,
		    play_integrity_verdict = $3,
		    device_recall_bits = $4,
		    last_seen_at = $5
		WHERE id = $1::uuid
		RETURNING id::text, user_id::text, widevine_hash, play_integrity_verdict,
		          device_recall_bits, status, blocked, last_seen_at, created_at`,
		deviceID, widevineHash, verdict, deviceRecallBits, lastSeenAt.UTC(),
	)
	var out Device
	err := row.Scan(
		&out.ID, &out.UserID, &out.WidevineHash, &out.PlayIntegrityVerdict,
		&out.DeviceRecallBits, &out.Status, &out.Blocked, &out.LastSeenAt, &out.CreatedAt,
	)
	return out, err
}

// ResetDeviceRecallBits clears all three Device Recall abuse bits (Super Admin support).
func (r *DeviceRepository) ResetDeviceRecallBits(ctx context.Context, deviceID string) (Device, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE devices
		SET device_recall_bits = 0
		WHERE id = $1::uuid
		RETURNING id::text, user_id::text, widevine_hash, play_integrity_verdict,
		          device_recall_bits, status, blocked, last_seen_at, created_at`,
		deviceID,
	)
	var out Device
	err := row.Scan(
		&out.ID, &out.UserID, &out.WidevineHash, &out.PlayIntegrityVerdict,
		&out.DeviceRecallBits, &out.Status, &out.Blocked, &out.LastSeenAt, &out.CreatedAt,
	)
	return out, err
}

func defaultStatus(s string) string {
	if s == "" {
		return "active"
	}
	return s
}
