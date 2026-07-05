// Package postgres implements the domain repository ports over pgx. The SQL executed here
// is the same SQL declared in db/queries (the sqlc authority) — running `sqlc generate`
// produces a typed query layer these repositories can adopt without changing call sites.
//
// Store wires the repositories together and provides transactional execution: WithinTx
// runs a function against a Store whose repositories all share one pgx.Tx, so a use case's
// writes (including the transactional outbox) commit or roll back atomically.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// pgxConn is the subset of pgx behaviour shared by *pgxpool.Pool and pgx.Tx, letting the
// repositories run identically inside or outside a transaction.
type pgxConn interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Store is the composition of all auth repositories over a single connection or tx.
type Store struct {
	pool *pgxpool.Pool
	db   pgxConn
}

var _ domain.Store = (*Store)(nil)

// NewStore builds a non-transactional Store backed by the connection pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, db: pool}
}

func (s *Store) Accounts() domain.AccountRepository    { return &accountRepository{db: s.db} }
func (s *Store) Sessions() domain.SessionRepository    { return &sessionRepository{db: s.db} }
func (s *Store) OAuthIdentities() domain.OAuthRepository { return &oauthRepository{db: s.db} }
func (s *Store) Outbox() domain.OutboxRepository       { return &outboxRepository{db: s.db} }

// WithinTx begins a transaction, runs fn against a tx-scoped Store, and commits on success
// or rolls back on error (or panic).
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

// isUniqueViolation reports whether err is a Postgres unique-constraint violation and, if
// so, returns the violated constraint name.
func isUniqueViolation(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return pgErr.ConstraintName, true
	}
	return "", false
}

// isNoRows reports whether err signals an empty result set.
func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
