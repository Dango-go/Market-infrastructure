package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/catalog/internal/domain"
	"github.com/google/uuid"
)

type brandRepository struct{ db pgxConn }

func (r *brandRepository) Create(ctx context.Context, b *domain.Brand) error {
	const q = `INSERT INTO brands (id, name, slug, description, country_code, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	if _, err := r.db.Exec(ctx, q, b.ID, b.Name, b.Slug, b.Description, b.CountryCode, b.CreatedAt, b.UpdatedAt); err != nil {
		if isUniqueViolation(err) { return domain.ErrSlugTaken }
		return fmt.Errorf("insert brand: %w", err)
	}
	return nil
}

func (r *brandRepository) List(ctx context.Context) ([]*domain.Brand, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, slug, description, country_code, created_at, updated_at FROM brands ORDER BY name ASC`)
	if err != nil { return nil, fmt.Errorf("list brands: %w", err) }
	defer rows.Close()
	items := make([]*domain.Brand, 0)
	for rows.Next() {
		var item domain.Brand
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.CountryCode, &item.CreatedAt, &item.UpdatedAt); err != nil { return nil, fmt.Errorf("scan brand: %w", err) }
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate brands: %w", err) }
	return items, nil
}

func (r *brandRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Brand, error) {
	var item domain.Brand
	if err := r.db.QueryRow(ctx, `SELECT id, name, slug, description, country_code, created_at, updated_at FROM brands WHERE id = $1`, id).Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.CountryCode, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrBrandNotFound }
		return nil, fmt.Errorf("get brand: %w", err)
	}
	return &item, nil
}
