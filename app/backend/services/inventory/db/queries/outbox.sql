-- name: EnqueueOutboxEvent :exec
INSERT INTO outbox (id, type, version, source, subject, correlation_id, occurred_at, data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
