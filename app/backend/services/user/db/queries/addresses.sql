-- name: CreateAddress :exec
INSERT INTO addresses (id, account_id, label, recipient_name, line1, line2, city, region, postal_code, country_code, phone, is_default_shipping, is_default_billing, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: GetAddressByID :one
SELECT id, account_id, label, recipient_name, line1, line2, city, region, postal_code, country_code, phone, is_default_shipping, is_default_billing, created_at, updated_at, deleted_at
FROM addresses
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListAddressesByAccountID :many
SELECT id, account_id, label, recipient_name, line1, line2, city, region, postal_code, country_code, phone, is_default_shipping, is_default_billing, created_at, updated_at, deleted_at
FROM addresses
WHERE account_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAddressesByAccountID :one
SELECT COUNT(*) FROM addresses WHERE account_id = $1 AND deleted_at IS NULL;

-- name: UpdateAddress :execrows
UPDATE addresses
SET label = $2, recipient_name = $3, line1 = $4, line2 = $5, city = $6, region = $7, postal_code = $8, country_code = $9, phone = $10, is_default_shipping = $11, is_default_billing = $12, updated_at = $13, deleted_at = $14
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteAddress :execrows
UPDATE addresses SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL;

-- name: ClearDefaultShippingByAccount :exec
UPDATE addresses SET is_default_shipping = FALSE, updated_at = now() WHERE account_id = $1 AND deleted_at IS NULL AND is_default_shipping = TRUE;

-- name: ClearDefaultBillingByAccount :exec
UPDATE addresses SET is_default_billing = FALSE, updated_at = now() WHERE account_id = $1 AND deleted_at IS NULL AND is_default_billing = TRUE;
