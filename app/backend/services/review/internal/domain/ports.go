package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Clock interface { Now() time.Time }

type IDGenerator interface { New() uuid.UUID }

type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*Review, error)
	ListByProductID(ctx context.Context, productID uuid.UUID, limit, offset int32) ([]Review, int64, error)
	ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]Review, int64, error)
	Update(ctx context.Context, review *Review) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetSummaryByProductID(ctx context.Context, productID uuid.UUID) (ReviewSummary, error)
}

type Store interface { Reviews() ReviewRepository }
