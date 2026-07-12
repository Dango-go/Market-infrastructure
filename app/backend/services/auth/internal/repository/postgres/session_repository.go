package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

type sessionRepository struct {
	db pgxConn
}

const sessionColumns = `id, account_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, rotated_from, created_at, last_used_at`

func (r *sessionRepository) Create(ctx context.Context, s *domain.Session) error {
	const q = `
		INSERT INTO sessions (id, account_id, refresh_token_hash, user_agent, ip_address, expires_at, rotated_from, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, q,
		s.ID, s.AccountID, s.RefreshTokenHash, s.UserAgent, s.IPAddress, s.ExpiresAt, s.RotatedFrom, s.CreatedAt, s.LastUsedAt,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	q := `SELECT ` + sessionColumns + ` FROM sessions WHERE id = $1`
	return r.scanOne(ctx, q, id)
}

func (r *sessionRepository) GetByRefreshHash(ctx context.Context, hash string) (*domain.Session, error) {
	q := `SELECT ` + sessionColumns + ` FROM sessions WHERE refresh_token_hash = $1`
	return r.scanOne(ctx, q, hash)
}

func (r *sessionRepository) ListActiveByAccount(ctx context.Context, accountID uuid.UUID, now time.Time, limit, offset int32) ([]*domain.Session, int64, error) {
	const countQ = `SELECT COUNT(*) FROM sessions WHERE account_id = $1 AND revoked_at IS NULL AND expires_at > $2`
	var total int64
	if err := r.db.QueryRow(ctx, countQ, accountID, now).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sessions: %w", err)
	}

	q := `SELECT ` + sessionColumns + `
		FROM sessions
		WHERE account_id = $1 AND revoked_at IS NULL AND expires_at > $2
		ORDER BY last_used_at DESC
		LIMIT $3 OFFSET $4`
	rows, err := r.db.Query(ctx, q, accountID, now, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		s, err := scanSessionRow(rows)
		if err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate sessions: %w", err)
	}
	return sessions, total, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id uuid.UUID, now time.Time) error {
	const q = `UPDATE sessions SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL`
	tag, err := r.db.Exec(ctx, q, id, now)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	// Zero rows simply means it was already revoked or absent — revoke is idempotent.
	_ = tag
	return nil
}

func (r *sessionRepository) RevokeAllByAccount(ctx context.Context, accountID uuid.UUID, now time.Time) (int64, error) {
	const q = `UPDATE sessions SET revoked_at = $2 WHERE account_id = $1 AND revoked_at IS NULL`
	tag, err := r.db.Exec(ctx, q, accountID, now)
	if err != nil {
		return 0, fmt.Errorf("revoke account sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *sessionRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.Session, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("query session: %w", err)
		}
		return nil, domain.ErrSessionNotFound
	}
	return scanSessionRow(rows)
}

// rowScanner is satisfied by pgx.Rows for a single row.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanSessionRow(row rowScanner) (*domain.Session, error) {
	var s domain.Session
	if err := row.Scan(
		&s.ID, &s.AccountID, &s.RefreshTokenHash, &s.UserAgent, &s.IPAddress,
		&s.ExpiresAt, &s.RevokedAt, &s.RotatedFrom, &s.CreatedAt, &s.LastUsedAt,
	); err != nil {
		return nil, fmt.Errorf("scan session: %w", err)
	}
	return &s, nil
}
