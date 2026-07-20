-- name: CreatePreferencesIfMissing :execrows
INSERT INTO preferences (account_id, currency, language, email_notifications, sms_notifications, push_notifications, marketing_opt_in, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (account_id) DO NOTHING;

-- name: GetPreferencesByAccountID :one
SELECT account_id, currency, language, email_notifications, sms_notifications, push_notifications, marketing_opt_in, created_at, updated_at
FROM preferences
WHERE account_id = $1;

-- name: UpdatePreferences :execrows
UPDATE preferences
SET currency = $2, language = $3, email_notifications = $4, sms_notifications = $5, push_notifications = $6, marketing_opt_in = $7, updated_at = $8
WHERE account_id = $1;
