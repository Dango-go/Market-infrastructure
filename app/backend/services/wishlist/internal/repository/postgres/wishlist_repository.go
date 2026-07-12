package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/wishlist/internal/domain"
	"github.com/google/uuid"
)

type wishlistRepository struct{ db pgxConn }

func (r *wishlistRepository) Add(ctx context.Context, item *domain.WishlistItem) (bool, error) {
	const q = `INSERT INTO wishlist_items (account_id, product_id, added_at)
VALUES ($1, $2, $3)
ON CONFLICT (account_id, product_id) DO NOTHING`
	tag, err := r.db.Exec(ctx, q, item.AccountID, item.ProductID, item.AddedAt)
	if err != nil {
		return false, fmt.Errorf("insert wishlist item: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (r *wishlistRepository) Remove(ctx context.Context, accountID, productID uuid.UUID) error {
	const q = `DELETE FROM wishlist_items WHERE account_id = $1 AND product_id = $2`
	if _, err := r.db.Exec(ctx, q, accountID, productID); err != nil {
		return fmt.Errorf("delete wishlist item: %w", err)
	}
	return nil
}

func (r *wishlistRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]*domain.WishlistItem, int64, error) {
	const countQ = `SELECT COUNT(*) FROM wishlist_items WHERE account_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQ, accountID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count wishlist items: %w", err)
	}
	const q = `SELECT account_id, product_id, added_at FROM wishlist_items WHERE account_id = $1 ORDER BY added_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, accountID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list wishlist items: %w", err)
	}
	defer rows.Close()
	items := make([]*domain.WishlistItem, 0)
	for rows.Next() {
		var item domain.WishlistItem
		if err := rows.Scan(&item.AccountID, &item.ProductID, &item.AddedAt); err != nil {
			return nil, 0, fmt.Errorf("scan wishlist item: %w", err)
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate wishlist items: %w", err)
	}
	return items, total, nil
}
