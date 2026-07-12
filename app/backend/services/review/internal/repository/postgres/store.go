package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/embedded-market/backend/services/review/internal/domain"
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
func (s *Store) Reviews() domain.ReviewRepository { return &reviewRepository{db: s.db} }

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func countErr(err error, msg string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}
