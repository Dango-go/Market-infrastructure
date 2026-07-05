package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/order/internal/domain"
	"github.com/google/uuid"
)

type orderRepository struct{ db pgxConn }

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO orders (id, account_id, cart_id, status, currency, subtotal_cents, shipping_cents, total_cents, delivery_method, delivery_address, customer_note, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, order.ID, order.AccountID, order.CartID, string(order.Status), order.Currency, order.SubtotalCents, order.ShippingCents, order.TotalCents, order.DeliveryMethod, order.DeliveryAddress, order.CustomerNote, order.CreatedAt, order.UpdatedAt); err != nil { return fmt.Errorf("insert order: %w", err) }
	return nil
}

func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return r.getOne(ctx, `SELECT id, account_id, cart_id, status, currency, subtotal_cents, shipping_cents, total_cents, delivery_method, delivery_address, customer_note, created_at, updated_at FROM orders WHERE id = $1`, id)
}

func (r *orderRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]domain.Order, int64, error) {
	rows, err := r.db.Query(ctx, `SELECT id, account_id, cart_id, status, currency, subtotal_cents, shipping_cents, total_cents, delivery_method, delivery_address, customer_note, created_at, updated_at, COUNT(*) OVER() AS total_count FROM orders WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, accountID, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list orders: %w", err) }
	defer rows.Close()
	items := make([]domain.Order, 0)
	var total int64
	for rows.Next() {
		var item domain.Order
		var status string
		if err := rows.Scan(&item.ID, &item.AccountID, &item.CartID, &status, &item.Currency, &item.SubtotalCents, &item.ShippingCents, &item.TotalCents, &item.DeliveryMethod, &item.DeliveryAddress, &item.CustomerNote, &item.CreatedAt, &item.UpdatedAt, &total); err != nil { return nil, 0, fmt.Errorf("scan order: %w", err) }
		item.Status = domain.OrderStatus(status)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("iterate orders: %w", err) }
	return items, total, nil
}

func (r *orderRepository) Update(ctx context.Context, order *domain.Order) error {
	tag, err := r.db.Exec(ctx, `UPDATE orders SET status = $2, subtotal_cents = $3, shipping_cents = $4, total_cents = $5, delivery_method = $6, delivery_address = $7, customer_note = $8, updated_at = $9 WHERE id = $1`, order.ID, string(order.Status), order.SubtotalCents, order.ShippingCents, order.TotalCents, order.DeliveryMethod, order.DeliveryAddress, order.CustomerNote, order.UpdatedAt)
	if err != nil { return fmt.Errorf("update order: %w", err) }
	if tag.RowsAffected() == 0 { return domain.ErrOrderNotFound }
	return nil
}

func (r *orderRepository) getOne(ctx context.Context, query string, arg any) (*domain.Order, error) {
	var item domain.Order
	var status string
	if err := r.db.QueryRow(ctx, query, arg).Scan(&item.ID, &item.AccountID, &item.CartID, &status, &item.Currency, &item.SubtotalCents, &item.ShippingCents, &item.TotalCents, &item.DeliveryMethod, &item.DeliveryAddress, &item.CustomerNote, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrOrderNotFound }
		return nil, fmt.Errorf("get order: %w", err)
	}
	item.Status = domain.OrderStatus(status)
	return &item, nil
}
