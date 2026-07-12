-- name: CreateWarehouse :exec
INSERT INTO warehouses (id, code, name, location, active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);
