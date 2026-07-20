-- name: MarkProcessedEvent :execrows
INSERT INTO processed_events (id, topic, processed_at)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO NOTHING;
