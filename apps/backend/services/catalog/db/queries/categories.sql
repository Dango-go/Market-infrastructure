-- name: CreateCategory :exec
INSERT INTO categories (id, name, slug, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListCategories :many
SELECT id, name, slug, description, created_at, updated_at
FROM categories
ORDER BY name ASC;
