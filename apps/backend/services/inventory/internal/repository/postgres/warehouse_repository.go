package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/inventory/internal/domain"
	"github.com/google/uuid"
)

type warehouseRepository struct{ db pgxConn }

func (r *warehouseRepository) Create(ctx context.Context, w *domain.Warehouse) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO warehouses (id, code, name, location, active, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, w.ID, w.Code, w.Name, w.Location, w.Active, w.CreatedAt, w.UpdatedAt); err != nil { return fmt.Errorf("insert warehouse: %w", err) }
	return nil
}

func (r *warehouseRepository) List(ctx context.Context) ([]*domain.Warehouse, error) {
	rows, err := r.db.Query(ctx, `SELECT id, code, name, location, active, created_at, updated_at FROM warehouses ORDER BY code ASC`)
	if err != nil { return nil, fmt.Errorf("list warehouses: %w", err) }
	defer rows.Close()
	items := make([]*domain.Warehouse, 0)
	for rows.Next() {
		var item domain.Warehouse
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Location, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil { return nil, fmt.Errorf("scan warehouse: %w", err) }
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate warehouses: %w", err) }
	return items, nil
}

func (r *warehouseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	var item domain.Warehouse
	if err := r.db.QueryRow(ctx, `SELECT id, code, name, location, active, created_at, updated_at FROM warehouses WHERE id = $1`, id).Scan(&item.ID, &item.Code, &item.Name, &item.Location, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrWarehouseNotFound }
		return nil, fmt.Errorf("get warehouse: %w", err)
	}
	return &item, nil
}
