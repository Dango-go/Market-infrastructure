package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/cart/internal/domain"
	"github.com/google/uuid"
)

type cartRepository struct{ db pgxConn }

func (r *cartRepository) Create(ctx context.Context, cart *domain.Cart) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO carts (id, account_id, status, currency, subtotal_cents, created_at, updated_at, checked_out_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, cart.ID, cart.AccountID, string(cart.Status), cart.Currency, cart.SubtotalCents, cart.CreatedAt, cart.UpdatedAt, cart.CheckedOutAt); err != nil { return fmt.Errorf("insert cart: %w", err) }
	return nil
}

func (r *cartRepository) GetActiveByAccountID(ctx context.Context, accountID uuid.UUID) (*domain.Cart, error) {
	return r.getOne(ctx, `SELECT id, account_id, status, currency, subtotal_cents, created_at, updated_at, checked_out_at FROM carts WHERE account_id = $1 AND status = 'active' ORDER BY created_at DESC LIMIT 1`, accountID)
}

func (r *cartRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cart, error) {
	return r.getOne(ctx, `SELECT id, account_id, status, currency, subtotal_cents, created_at, updated_at, checked_out_at FROM carts WHERE id = $1`, id)
}

func (r *cartRepository) Update(ctx context.Context, cart *domain.Cart) error {
	tag, err := r.db.Exec(ctx, `UPDATE carts SET status = $2, currency = $3, subtotal_cents = $4, updated_at = $5, checked_out_at = $6 WHERE id = $1`, cart.ID, string(cart.Status), cart.Currency, cart.SubtotalCents, cart.UpdatedAt, cart.CheckedOutAt)
	if err != nil { return fmt.Errorf("update cart: %w", err) }
	if tag.RowsAffected() == 0 { return domain.ErrCartNotFound }
	return nil
}

func (r *cartRepository) getOne(ctx context.Context, query string, arg any) (*domain.Cart, error) {
	var item domain.Cart
	var status string
	if err := r.db.QueryRow(ctx, query, arg).Scan(&item.ID, &item.AccountID, &status, &item.Currency, &item.SubtotalCents, &item.CreatedAt, &item.UpdatedAt, &item.CheckedOutAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrCartNotFound }
		return nil, fmt.Errorf("get cart: %w", err)
	}
	item.Status = domain.CartStatus(status)
	return &item, nil
}
