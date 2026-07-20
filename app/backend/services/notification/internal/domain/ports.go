package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ New() uuid.UUID }

type TemplateRepository interface {
	Create(ctx context.Context, template *Template) error
	GetByID(ctx context.Context, id uuid.UUID) (*Template, error)
	List(ctx context.Context, limit, offset int32) ([]Template, int64, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]Notification, int64, error)
	Update(ctx context.Context, notification *Notification) error
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

type Store interface {
	Templates() TemplateRepository
	Notifications() NotificationRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}
