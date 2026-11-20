package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Plan struct {
	ID        string
	Name      string
	PlanType  string
	PriceCop  int64
	PlaySKU   *string
	Active    bool
	CityCode  *string
	CreatedAt time.Time
}

type PlanRepository struct {
	pool *pgxpool.Pool
}

func NewPlanRepository(pool *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{pool: pool}
}

func (r *PlanRepository) Insert(ctx context.Context, p Plan) (Plan, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO plans (name, plan_type, price_cop, play_sku, active, city_code)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, name, plan_type, price_cop, play_sku, active, city_code, created_at`,
		p.Name, p.PlanType, p.PriceCop, p.PlaySKU, p.Active, p.CityCode,
	)
	var out Plan
	err := row.Scan(
		&out.ID, &out.Name, &out.PlanType, &out.PriceCop, &out.PlaySKU,
		&out.Active, &out.CityCode, &out.CreatedAt,
	)
	return out, err
}

func (r *PlanRepository) ListActive(ctx context.Context) ([]Plan, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, name, plan_type, price_cop, play_sku, active, city_code, created_at
		FROM plans
		WHERE active = TRUE
		ORDER BY price_cop ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Plan
	for rows.Next() {
		var p Plan
		if err := rows.Scan(
			&p.ID, &p.Name, &p.PlanType, &p.PriceCop, &p.PlaySKU,
			&p.Active, &p.CityCode, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
