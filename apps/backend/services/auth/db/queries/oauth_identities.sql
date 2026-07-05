-- name: CreateOAuthIdentity :exec
INSERT INTO oauth_identities (id, account_id, provider, provider_user_id, email, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetOAuthIdentityByProviderUserID :one
SELECT id, account_id, provider, provider_user_id, email, created_at, updated_at
FROM oauth_identities
WHERE provider = $1 AND provider_user_id = $2;
