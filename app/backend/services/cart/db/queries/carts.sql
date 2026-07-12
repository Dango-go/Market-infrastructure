-- name: CreateCart :exec
INSERT INTO carts (id, account_id, status, currency, subtotal_cents, created_at, updated_at, checked_out_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
