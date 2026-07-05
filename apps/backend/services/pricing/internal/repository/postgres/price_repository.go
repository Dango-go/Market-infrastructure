package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/pricing/internal/domain"
	"github.com/google/uuid"
)

type priceRepository struct{ db pgxConn }

func (r *priceRepository) Create(ctx context.Context, p *domain.Price) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO prices (id, product_id, currency, amount_cents, compare_at_cents, active, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, p.ID, p.ProductID, p.Currency, p.AmountCents, p.CompareAtCents, p.Active, p.CreatedAt, p.UpdatedAt); err != nil {
		return fmt.Errorf("insert price: %w", err)
	}
	return nil
}

func (r *priceRepository) Update(ctx context.Context, p *domain.Price) error {
	tag, err := r.db.Exec(ctx, `UPDATE prices SET currency = $2, amount_cents = $3, compare_at_cents = $4, active = $5, updated_at = $6 WHERE id = $1`, p.ID, p.Currency, p.AmountCents, p.CompareAtCents, p.Active, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update price: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPriceNotFound
	}
	return nil
}

func (r *priceRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*domain.Price, error) {
	var item domain.Price
	if err := r.db.QueryRow(ctx, `SELECT id, product_id, currency, amount_cents, compare_at_cents, active, created_at, updated_at FROM prices WHERE product_id = $1`, productID).Scan(&item.ID, &item.ProductID, &item.Currency, &item.AmountCents, &item.CompareAtCents, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("get price: %w", err)
	}
	return &item, nil
}

func (r *priceRepository) List(ctx context.Context, limit, offset int32) ([]*domain.Price, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM prices`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count prices: %w", err)
	}
	rows, err := r.db.Query(ctx, `SELECT id, product_id, currency, amount_cents, compare_at_cents, active, created_at, updated_at FROM prices ORDER BY updated_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list prices: %w", err)
	}
	defer rows.Close()
	items := make([]*domain.Price, 0)
	for rows.Next() {
		var item domain.Price
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Currency, &item.AmountCents, &item.CompareAtCents, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan price: %w", err)
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate prices: %w", err)
	}
	return items, total, nil
}
