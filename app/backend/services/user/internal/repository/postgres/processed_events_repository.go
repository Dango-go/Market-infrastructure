package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/embedded-market/backend/services/user/internal/domain"
	"github.com/google/uuid"
)

type processedEventRepository struct{ db pgxConn }

func (r *processedEventRepository) Mark(ctx context.Context, eventID uuid.UUID, topic string, processedAt time.Time) (bool, error) {
	const q = `INSERT INTO processed_events (id, topic, processed_at) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`
	tag, err := r.db.Exec(ctx, q, eventID, topic, processedAt)
	if err != nil {
		return false, fmt.Errorf("mark processed event: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

var _ = domain.ErrProfileNotFound
