package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/embedded-market/backend/services/search/internal/domain"
)

type pgxConn interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Store struct {
	pool *pgxpool.Pool
	db   pgxConn
}

var _ domain.Store = (*Store)(nil)

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool, db: pool} }
func (s *Store) Documents() domain.DocumentRepository { return &documentRepository{db: s.db} }

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

func tagsToCSV(tags []string) string {
	if len(tags) == 0 { return "" }
	out := tags[0]
	for i := 1; i < len(tags); i++ { out += "," + tags[i] }
	return out
}

func csvToTags(raw string) []string {
	if raw == "" { return []string{} }
	out := make([]string, 0)
	start := 0
	for i := 0; i <= len(raw); i++ {
		if i == len(raw) || raw[i] == ',' {
			if i > start { out = append(out, raw[start:i]) }
			start = i + 1
		}
	}
	return out
}

func countErr(err error, msg string) error {
	if err != nil { return fmt.Errorf("%s: %w", msg, err) }
	return nil
}
