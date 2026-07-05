-- name: CreatePromotion :exec
INSERT INTO promotions (id, name, code, discount_type, value_cents, percent_off, active, starts_at, ends_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
