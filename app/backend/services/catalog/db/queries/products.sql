-- name: CreateProduct :exec
INSERT INTO products (id, category_id, brand_id, slug, sku, name, short_description, description, datasheet_url, image_url, status, featured, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: GetProductBySlug :one
SELECT id, category_id, brand_id, slug, sku, name, short_description, description, datasheet_url, image_url, status, featured, created_by, created_at, updated_at
FROM products
WHERE slug = $1;
