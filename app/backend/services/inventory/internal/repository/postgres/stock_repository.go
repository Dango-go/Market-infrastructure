package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/embedded-market/backend/services/inventory/internal/domain"
	"github.com/google/uuid"
)

type stockRepository struct{ db pgxConn }

func (r *stockRepository) CreateOrGet(ctx context.Context, item *domain.StockItem) (*domain.StockItem, error) {
	_, err := r.db.Exec(ctx, `INSERT INTO stock_items (id, product_id, warehouse_id, on_hand, reserved, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (product_id, warehouse_id) DO NOTHING`, item.ID, item.ProductID, item.WarehouseID, item.OnHand, item.Reserved, item.CreatedAt, item.UpdatedAt)
	if err != nil { return nil, fmt.Errorf("insert stock item: %w", err) }
	return r.GetByProductWarehouse(ctx, item.ProductID, item.WarehouseID)
}

func (r *stockRepository) GetByProductWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) (*domain.StockItem, error) {
	var item domain.StockItem
	if err := r.db.QueryRow(ctx, `SELECT id, product_id, warehouse_id, on_hand, reserved, created_at, updated_at FROM stock_items WHERE product_id = $1 AND warehouse_id = $2`, productID, warehouseID).Scan(&item.ID, &item.ProductID, &item.WarehouseID, &item.OnHand, &item.Reserved, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrStockItemNotFound }
		return nil, fmt.Errorf("get stock item: %w", err)
	}
	return &item, nil
}

func (r *stockRepository) List(ctx context.Context, filters domain.StockFilters, limit, offset int32) ([]*domain.StockItem, int64, error) {
	where := []string{"1=1"}
	args := make([]any, 0)
	idx := 1
	if filters.ProductID != nil {
		where = append(where, fmt.Sprintf("product_id = $%d", idx))
		args = append(args, *filters.ProductID)
		idx++
	}
	if filters.WarehouseID != nil {
		where = append(where, fmt.Sprintf("warehouse_id = $%d", idx))
		args = append(args, *filters.WarehouseID)
		idx++
	}
	base := ` FROM stock_items WHERE ` + strings.Join(where, ` AND `)
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*)`+base, args...).Scan(&total); err != nil { return nil, 0, fmt.Errorf("count stock items: %w", err) }
	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.db.Query(ctx, fmt.Sprintf(`SELECT id, product_id, warehouse_id, on_hand, reserved, created_at, updated_at%s ORDER BY updated_at DESC LIMIT $%d OFFSET $%d`, base, idx, idx+1), listArgs...)
	if err != nil { return nil, 0, fmt.Errorf("list stock items: %w", err) }
	defer rows.Close()
	items := make([]*domain.StockItem, 0)
	for rows.Next() {
		var item domain.StockItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.WarehouseID, &item.OnHand, &item.Reserved, &item.CreatedAt, &item.UpdatedAt); err != nil { return nil, 0, fmt.Errorf("scan stock item: %w", err) }
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("iterate stock items: %w", err) }
	return items, total, nil
}

func (r *stockRepository) Update(ctx context.Context, item *domain.StockItem) error {
	tag, err := r.db.Exec(ctx, `UPDATE stock_items SET on_hand = $2, reserved = $3, updated_at = $4 WHERE id = $1`, item.ID, item.OnHand, item.Reserved, item.UpdatedAt)
	if err != nil { return fmt.Errorf("update stock item: %w", err) }
	if tag.RowsAffected() == 0 { return domain.ErrStockItemNotFound }
	return nil
}
