package domain

import (
	"context"
	"time"
)

type Clock interface { Now() time.Time }

type IDGenerator interface { New() [16]byte }

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	Overview(ctx context.Context, days int32) (Overview, error)
	TopProducts(ctx context.Context, days, limit int32) ([]TopProduct, error)
}

type Store interface { Events() EventRepository }
