package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/embedded-market/backend/services/user/internal/domain"
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

func (s *Store) Profiles() domain.ProfileRepository { return &profileRepository{db: s.db} }
func (s *Store) Addresses() domain.AddressRepository { return &addressRepository{db: s.db} }
func (s *Store) Preferences() domain.PreferencesRepository { return &preferencesRepository{db: s.db} }
func (s *Store) ProcessedEvents() domain.ProcessedEventRepository { return &processedEventRepository{db: s.db} }
func (s *Store) Outbox() domain.OutboxRepository { return &outboxRepository{db: s.db} }

func (s *Store) WithinTx(ctx context.Context, fn func(ctx context.Context, tx domain.Store) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
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

func isUniqueViolation(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return pgErr.ConstraintName, true
	}
	return "", false
}

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
