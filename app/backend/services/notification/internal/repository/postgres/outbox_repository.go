package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type outboxRepository struct{ db pgxConn }

func (r *outboxRepository) Enqueue(ctx context.Context, env events.Envelope) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO outbox (id, type, version, source, subject, correlation_id, occurred_at, data) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, env.ID, string(env.Type), env.Version, env.Source, env.Subject, env.CorrelationID, env.OccurredAt, env.Data); err != nil { return fmt.Errorf("enqueue outbox: %w", err) }
	return nil
}

func (r *outboxRepository) FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error) {
	rows, err := r.db.Query(ctx, `SELECT id, type, version, source, subject, correlation_id, occurred_at, data FROM outbox WHERE published_at IS NULL ORDER BY occurred_at ASC LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil { return nil, fmt.Errorf("fetch outbox: %w", err) }
	defer rows.Close()
	items := make([]events.Envelope, 0)
	for rows.Next() {
		var env events.Envelope
		var typ string
		if err := rows.Scan(&env.ID, &typ, &env.Version, &env.Source, &env.Subject, &env.CorrelationID, &env.OccurredAt, &env.Data); err != nil { return nil, fmt.Errorf("scan outbox: %w", err) }
		env.Type = events.Topic(typ)
		items = append(items, env)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate outbox: %w", err) }
	return items, nil
}

func (r *outboxRepository) MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error {
	if len(ids) == 0 { return nil }
	if _, err := r.db.Exec(ctx, `UPDATE outbox SET published_at = $2 WHERE id = ANY($1::uuid[])`, ids, publishedAt); err != nil { return fmt.Errorf("mark published: %w", err) }
	return nil
}
