package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/cart/internal/domain"
	"github.com/google/uuid"
)

type cartItemRepository struct{ db pgxConn }

func (r *cartItemRepository) ListByCartID(ctx context.Context, cartID uuid.UUID) ([]domain.CartItem, error) {
	rows, err := r.db.Query(ctx, `SELECT id, cart_id, product_id, quantity, unit_price_cents, created_at, updated_at FROM cart_items WHERE cart_id = $1 ORDER BY created_at ASC`, cartID)
	if err != nil { return nil, fmt.Errorf("list cart items: %w", err) }
	defer rows.Close()
	items := make([]domain.CartItem, 0)
	for rows.Next() {
		var item domain.CartItem
		if err := rows.Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.UnitPriceCents, &item.CreatedAt, &item.UpdatedAt); err != nil { return nil, fmt.Errorf("scan cart item: %w", err) }
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate cart items: %w", err) }
	return items, nil
}

func (r *cartItemRepository) GetByCartProduct(ctx context.Context, cartID, productID uuid.UUID) (*domain.CartItem, error) {
	var item domain.CartItem
	if err := r.db.QueryRow(ctx, `SELECT id, cart_id, product_id, quantity, unit_price_cents, created_at, updated_at FROM cart_items WHERE cart_id = $1 AND product_id = $2`, cartID, productID).Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.UnitPriceCents, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrCartItemNotFound }
		return nil, fmt.Errorf("get cart item: %w", err)
	}
	return &item, nil
}

func (r *cartItemRepository) Create(ctx context.Context, item *domain.CartItem) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO cart_items (id, cart_id, product_id, quantity, unit_price_cents, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, item.ID, item.CartID, item.ProductID, item.Quantity, item.UnitPriceCents, item.CreatedAt, item.UpdatedAt); err != nil { return fmt.Errorf("insert cart item: %w", err) }
	return nil
}

func (r *cartItemRepository) Update(ctx context.Context, item *domain.CartItem) error {
	tag, err := r.db.Exec(ctx, `UPDATE cart_items SET quantity = $2, unit_price_cents = $3, updated_at = $4 WHERE id = $1`, item.ID, item.Quantity, item.UnitPriceCents, item.UpdatedAt)
	if err != nil { return fmt.Errorf("update cart item: %w", err) }
	if tag.RowsAffected() == 0 { return domain.ErrCartItemNotFound }
	return nil
}

func (r *cartItemRepository) Delete(ctx context.Context, itemID uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM cart_items WHERE id = $1`, itemID); err != nil { return fmt.Errorf("delete cart item: %w", err) }
	return nil
}

func (r *cartItemRepository) DeleteByCartID(ctx context.Context, cartID uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID); err != nil { return fmt.Errorf("clear cart items: %w", err) }
	return nil
}
