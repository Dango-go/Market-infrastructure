package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

const (
	constraintEmailUnique    = "uq_accounts_email"
	constraintUsernameUnique = "uq_accounts_username"
)

type accountRepository struct {
	db pgxConn
}

const accountColumns = `id, email, username, password_hash, status, email_verified, created_at, updated_at, deleted_at`

func (r *accountRepository) Create(ctx context.Context, a *domain.Account) error {
	const q = `
		INSERT INTO accounts (id, email, username, password_hash, status, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, q,
		a.ID, a.Email.String(), a.Username.String(), nullableHash(a.PasswordHash),
		a.Status, a.EmailVerified, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		if name, ok := isUniqueViolation(err); ok {
			switch name {
			case constraintEmailUnique:
				return domain.ErrEmailTaken
			case constraintUsernameUnique:
				return domain.ErrUsernameTaken
			}
		}
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

func (r *accountRepository) Update(ctx context.Context, a *domain.Account) error {
	const q = `
		UPDATE accounts
		SET email = $2, username = $3, password_hash = $4, status = $5,
		    email_verified = $6, updated_at = $7, deleted_at = $8
		WHERE id = $1`
	tag, err := r.db.Exec(ctx, q,
		a.ID, a.Email.String(), a.Username.String(), nullableHash(a.PasswordHash),
		a.Status, a.EmailVerified, a.UpdatedAt, a.DeletedAt,
	)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *accountRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	q := `SELECT ` + accountColumns + ` FROM accounts WHERE id = $1 AND deleted_at IS NULL`
	return r.scanOne(ctx, q, id)
}

func (r *accountRepository) GetByEmail(ctx context.Context, email domain.Email) (*domain.Account, error) {
	q := `SELECT ` + accountColumns + ` FROM accounts WHERE lower(email) = lower($1) AND deleted_at IS NULL`
	return r.scanOne(ctx, q, email.String())
}

func (r *accountRepository) GetByUsername(ctx context.Context, username domain.Username) (*domain.Account, error) {
	q := `SELECT ` + accountColumns + ` FROM accounts WHERE lower(username) = lower($1) AND deleted_at IS NULL`
	return r.scanOne(ctx, q, username.String())
}

func (r *accountRepository) ExistsByEmail(ctx context.Context, email domain.Email) (bool, error) {
	return r.exists(ctx, `SELECT EXISTS(SELECT 1 FROM accounts WHERE lower(email) = lower($1) AND deleted_at IS NULL)`, email.String())
}

func (r *accountRepository) ExistsByUsername(ctx context.Context, username domain.Username) (bool, error) {
	return r.exists(ctx, `SELECT EXISTS(SELECT 1 FROM accounts WHERE lower(username) = lower($1) AND deleted_at IS NULL)`, username.String())
}

func (r *accountRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.Account, error) {
	var (
		a         domain.Account
		email     string
		username  string
		hash      *string
		status    string
		deletedAt *time.Time
	)
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&a.ID, &email, &username, &hash, &status, &a.EmailVerified, &a.CreatedAt, &a.UpdatedAt, &deletedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("scan account: %w", err)
	}
	a.Email = domain.Email(email)
	a.Username = domain.Username(username)
	if hash != nil {
		a.PasswordHash = domain.PasswordHash(*hash)
	}
	a.Status = domain.AccountStatus(status)
	a.DeletedAt = deletedAt
	return &a, nil
}

func (r *accountRepository) exists(ctx context.Context, query string, args ...any) (bool, error) {
	var ok bool
	if err := r.db.QueryRow(ctx, query, args...).Scan(&ok); err != nil {
		return false, fmt.Errorf("exists check: %w", err)
	}
	return ok, nil
}

func nullableHash(h domain.PasswordHash) any {
	if h == "" {
		return nil
	}
	return h.String()
}
