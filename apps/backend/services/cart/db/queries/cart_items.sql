-- name: CreateCartItem :exec
INSERT INTO cart_items (id, cart_id, product_id, quantity, unit_price_cents, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);
