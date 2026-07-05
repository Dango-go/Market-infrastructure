-- name: UpsertPrice :exec
INSERT INTO prices (id, product_id, currency, amount_cents, compare_at_cents, active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
