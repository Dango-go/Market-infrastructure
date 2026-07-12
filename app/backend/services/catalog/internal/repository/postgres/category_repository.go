package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/catalog/internal/domain"
	"github.com/google/uuid"
)

type categoryRepository struct{ db pgxConn }

func (r *categoryRepository) Create(ctx context.Context, c *domain.Category) error {
	const q = `INSERT INTO categories (id, name, slug, description, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6)`
	if _, err := r.db.Exec(ctx, q, c.ID, c.Name, c.Slug, c.Description, c.CreatedAt, c.UpdatedAt); err != nil {
		if isUniqueViolation(err) { return domain.ErrSlugTaken }
		return fmt.Errorf("insert category: %w", err)
	}
	return nil
}

func (r *categoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, slug, description, created_at, updated_at FROM categories ORDER BY name ASC`)
	if err != nil { return nil, fmt.Errorf("list categories: %w", err) }
	defer rows.Close()
	items := make([]*domain.Category, 0)
	for rows.Next() {
		var item domain.Category
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil { return nil, fmt.Errorf("scan category: %w", err) }
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate categories: %w", err) }
	return items, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	var item domain.Category
	if err := r.db.QueryRow(ctx, `SELECT id, name, slug, description, created_at, updated_at FROM categories WHERE id = $1`, id).Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrCategoryNotFound }
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &item, nil
}
