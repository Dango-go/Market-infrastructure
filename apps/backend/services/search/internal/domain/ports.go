package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ New() uuid.UUID }

type DocumentRepository interface {
	Upsert(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*Document, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, filters SearchFilters, limit, offset int32) ([]Document, int64, error)
	Suggest(ctx context.Context, query string, limit int32) ([]string, error)
}

type Store interface {
	Documents() DocumentRepository
}
