-- name: EnqueueOutboxEvent :exec
INSERT INTO outbox (id, type, version, source, subject, correlation_id, occurred_at, data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FetchUnpublishedOutbox :many
SELECT id, type, version, source, subject, correlation_id, occurred_at, data
FROM outbox
WHERE published_at IS NULL
ORDER BY occurred_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxPublished :exec
UPDATE outbox SET published_at = $2 WHERE id = ANY($1::uuid[]);
