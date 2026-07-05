-- name: CreateReservation :exec
INSERT INTO reservations (id, product_id, warehouse_id, reference, quantity, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
