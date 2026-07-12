package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/order/internal/domain"
	"github.com/google/uuid"
)

type orderItemRepository struct{ db pgxConn }

func (r *orderItemRepository) ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	rows, err := r.db.Query(ctx, `SELECT id, order_id, product_id, quantity, unit_price_cents, created_at, updated_at FROM order_items WHERE order_id = $1 ORDER BY created_at ASC`, orderID)
	if err != nil { return nil, fmt.Errorf("list order items: %w", err) }
	defer rows.Close()
	items := make([]domain.OrderItem, 0)
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.UnitPriceCents, &item.CreatedAt, &item.UpdatedAt); err != nil { return nil, fmt.Errorf("scan order item: %w", err) }
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate order items: %w", err) }
	return items, nil
}

func (r *orderItemRepository) Create(ctx context.Context, item *domain.OrderItem) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO order_items (id, order_id, product_id, quantity, unit_price_cents, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, item.ID, item.OrderID, item.ProductID, item.Quantity, item.UnitPriceCents, item.CreatedAt, item.UpdatedAt); err != nil { return fmt.Errorf("insert order item: %w", err) }
	return nil
}
