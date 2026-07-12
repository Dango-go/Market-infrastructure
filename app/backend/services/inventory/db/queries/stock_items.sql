-- name: UpsertStockItemSeed :exec
INSERT INTO stock_items (id, product_id, warehouse_id, on_hand, reserved, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (product_id, warehouse_id) DO NOTHING;
