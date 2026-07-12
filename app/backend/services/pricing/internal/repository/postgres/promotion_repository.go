package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/embedded-market/backend/services/pricing/internal/domain"
)

type promotionRepository struct{ db pgxConn }

func (r *promotionRepository) Create(ctx context.Context, p *domain.Promotion) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO promotions (id, name, code, discount_type, value_cents, percent_off, active, starts_at, ends_at, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, p.ID, p.Name, p.Code, p.DiscountType, p.ValueCents, p.PercentOff, p.Active, p.StartsAt, p.EndsAt, p.CreatedAt, p.UpdatedAt); err != nil {
		return fmt.Errorf("insert promotion: %w", err)
	}
	return nil
}

func (r *promotionRepository) List(ctx context.Context, limit, offset int32) ([]*domain.Promotion, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM promotions`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count promotions: %w", err)
	}
	rows, err := r.db.Query(ctx, `SELECT id, name, code, discount_type, value_cents, percent_off, active, starts_at, ends_at, created_at, updated_at FROM promotions ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list promotions: %w", err)
	}
	defer rows.Close()
	items := make([]*domain.Promotion, 0)
	for rows.Next() {
		var item domain.Promotion
		var endsAt *time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.DiscountType, &item.ValueCents, &item.PercentOff, &item.Active, &item.StartsAt, &endsAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan promotion: %w", err)
		}
		item.EndsAt = endsAt
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate promotions: %w", err)
	}
	return items, total, nil
}
