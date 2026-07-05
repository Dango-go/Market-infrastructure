-- name: CreateSession :exec
INSERT INTO sessions (id, account_id, refresh_token_hash, user_agent, ip_address, expires_at, rotated_from, created_at, last_used_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetSessionByID :one
SELECT id, account_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, rotated_from, created_at, last_used_at
FROM sessions
WHERE id = $1;

-- name: GetSessionByRefreshHash :one
SELECT id, account_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, rotated_from, created_at, last_used_at
FROM sessions
WHERE refresh_token_hash = $1;

-- name: CountActiveSessionsByAccount :one
SELECT COUNT(*)
FROM sessions
WHERE account_id = $1 AND revoked_at IS NULL AND expires_at > $2;

-- name: ListActiveSessionsByAccount :many
SELECT id, account_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, rotated_from, created_at, last_used_at
FROM sessions
WHERE account_id = $1 AND revoked_at IS NULL AND expires_at > $2
ORDER BY last_used_at DESC
LIMIT $3 OFFSET $4;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeAllSessionsByAccount :execrows
UPDATE sessions SET revoked_at = $2 WHERE account_id = $1 AND revoked_at IS NULL;
