package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/recommendation/internal/domain"
)

type eventRepository struct{ db pgxConn }

func (r *eventRepository) Create(ctx context.Context, event *domain.Event) error {
	_, err := r.db.Exec(ctx, `INSERT INTO recommendation_events (id, product_id, account_id, type, created_at) VALUES ($1,$2,$3,$4,$5)`, event.ID, event.ProductID, event.AccountID, event.Type, event.CreatedAt)
	if err != nil { return fmt.Errorf("create event: %w", err) }
	return nil
}
