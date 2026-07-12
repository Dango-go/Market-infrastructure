package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ New() uuid.UUID }

type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]Payment, int64, error)
	Update(ctx context.Context, payment *Payment) error
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

type Store interface {
	Payments() PaymentRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}
