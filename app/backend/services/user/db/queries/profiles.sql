-- name: CreateProfileIfMissing :execrows
INSERT INTO profiles (account_id, email, username, display_name, bio, phone, avatar_url, locale, timezone, marketing_opt_in, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (account_id) DO NOTHING;

-- name: GetProfileByAccountID :one
SELECT account_id, email, username, display_name, bio, phone, avatar_url, locale, timezone, marketing_opt_in, created_at, updated_at, deleted_at
FROM profiles
WHERE account_id = $1 AND deleted_at IS NULL;

-- name: UpdateProfile :execrows
UPDATE profiles
SET email = $2, username = $3, display_name = $4, bio = $5, phone = $6, avatar_url = $7, locale = $8, timezone = $9, marketing_opt_in = $10, updated_at = $11, deleted_at = $12
WHERE account_id = $1 AND deleted_at IS NULL;
