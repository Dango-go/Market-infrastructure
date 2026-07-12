package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/embedded-market/backend/services/catalog/internal/domain"
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
func (s *Store) Products() domain.ProductRepository { return &productRepository{db: s.db} }
func (s *Store) Categories() domain.CategoryRepository { return &categoryRepository{db: s.db} }
func (s *Store) Brands() domain.BrandRepository { return &brandRepository{db: s.db} }
func (s *Store) Outbox() domain.OutboxRepository { return &outboxRepository{db: s.db} }

func (s *Store) WithinTx(ctx context.Context, fn func(context.Context, domain.Store) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil { return fmt.Errorf("begin tx: %w", err) }
	txStore := &Store{pool: s.pool, db: tx}
	if err := fn(ctx, txStore); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
