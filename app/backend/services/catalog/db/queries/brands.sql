-- name: CreateBrand :exec
INSERT INTO brands (id, name, slug, description, country_code, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListBrands :many
SELECT id, name, slug, description, country_code, created_at, updated_at
FROM brands
ORDER BY name ASC;
