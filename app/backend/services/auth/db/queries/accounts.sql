-- name: CreateAccount :exec
INSERT INTO accounts (id, email, username, password_hash, status, email_verified, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpdateAccount :execrows
UPDATE accounts
SET email = $2, username = $3, password_hash = $4, status = $5,
    email_verified = $6, updated_at = $7, deleted_at = $8
WHERE id = $1;

-- name: GetAccountByID :one
SELECT id, email, username, password_hash, status, email_verified, created_at, updated_at, deleted_at
FROM accounts
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetAccountByEmail :one
SELECT id, email, username, password_hash, status, email_verified, created_at, updated_at, deleted_at
FROM accounts
WHERE lower(email) = lower($1) AND deleted_at IS NULL;

-- name: GetAccountByUsername :one
SELECT id, email, username, password_hash, status, email_verified, created_at, updated_at, deleted_at
FROM accounts
WHERE lower(username) = lower($1) AND deleted_at IS NULL;

-- name: ExistsAccountByEmail :one
SELECT EXISTS(SELECT 1 FROM accounts WHERE lower(email) = lower($1) AND deleted_at IS NULL);

-- name: ExistsAccountByUsername :one
SELECT EXISTS(SELECT 1 FROM accounts WHERE lower(username) = lower($1) AND deleted_at IS NULL);
