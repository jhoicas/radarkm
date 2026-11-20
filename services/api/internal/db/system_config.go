package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SystemConfig struct {
	ID                string
	FreeDailyQuota    int64
	QuotaCountingMode string
	MinVersionCode    int64
	CityOverrides     json.RawMessage
	UpdatedAt         time.Time
}

type SystemConfigRepository struct {
	pool *pgxpool.Pool
}

func NewSystemConfigRepository(pool *pgxpool.Pool) *SystemConfigRepository {
	return &SystemConfigRepository{pool: pool}
}

func (r *SystemConfigRepository) GetLatest(ctx context.Context) (SystemConfig, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, free_daily_quota, quota_counting_mode, min_version_code,
		       city_overrides, updated_at
		FROM system_config
		ORDER BY updated_at DESC
		LIMIT 1`)
	var cfg SystemConfig
	err := row.Scan(
		&cfg.ID, &cfg.FreeDailyQuota, &cfg.QuotaCountingMode, &cfg.MinVersionCode,
		&cfg.CityOverrides, &cfg.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return SystemConfig{
			FreeDailyQuota:    2,
			QuotaCountingMode: "per_device",
			MinVersionCode:    1,
			CityOverrides:     json.RawMessage(`{}`),
		}, nil
	}
	return cfg, err
}

func (r *SystemConfigRepository) UpsertDefault(ctx context.Context, freeDailyQuota int64, mode string) (SystemConfig, error) {
	if mode == "" {
		mode = "per_device"
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO system_config (free_daily_quota, quota_counting_mode, min_version_code, city_overrides)
		VALUES ($1, $2, 1, '{}'::jsonb)
		RETURNING id::text, free_daily_quota, quota_counting_mode, min_version_code, city_overrides, updated_at`,
		freeDailyQuota, mode,
	)
	var cfg SystemConfig
	err := row.Scan(
		&cfg.ID, &cfg.FreeDailyQuota, &cfg.QuotaCountingMode, &cfg.MinVersionCode,
		&cfg.CityOverrides, &cfg.UpdatedAt,
	)
	return cfg, err
}

func (r *SystemConfigRepository) UpdateFreeDailyQuota(ctx context.Context, id string, freeDailyQuota int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE system_config
		SET free_daily_quota = $2, updated_at = NOW()
		WHERE id = $1::uuid`, id, freeDailyQuota)
	return err
}
