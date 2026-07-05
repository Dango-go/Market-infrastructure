package domain

import (
	"context"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/google/uuid"
)

// Clock abstracts the current time so use cases stay deterministic and testable.
type Clock interface {
	Now() time.Time
}

// PasswordHasher hashes and verifies plaintext passwords. Implemented with argon2id in
// the infrastructure layer; the plaintext never leaves this boundary.
type PasswordHasher interface {
	Hash(plaintext string) (PasswordHash, error)
	Verify(hash PasswordHash, plaintext string) (bool, error)
}

// IssuedToken bundles a signed access token with its absolute expiry.
type IssuedToken struct {
	Value     string
	ExpiresAt time.Time
}

// TokenService issues access tokens and mints/hashes opaque refresh tokens. All crypto
// lives in the implementation; the application reasons only about these values.
type TokenService interface {
	IssueAccessToken(account *Account, sessionID uuid.UUID, now time.Time) (IssuedToken, error)
	// GenerateRefreshToken returns the plaintext token (handed to the client once) and
	// its hash (the only form persisted).
	GenerateRefreshToken() (plaintext string, hash string, err error)
	HashRefreshToken(plaintext string) string
	// RefreshTTL is the lifetime applied to new sessions.
	RefreshTTL() time.Duration
}

// AccountRepository persists identity aggregates. Implementations translate unique
// constraint violations into ErrEmailTaken / ErrUsernameTaken.
type AccountRepository interface {
	Create(ctx context.Context, a *Account) error
	Update(ctx context.Context, a *Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByEmail(ctx context.Context, email Email) (*Account, error)
	GetByUsername(ctx context.Context, username Username) (*Account, error)
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
	ExistsByUsername(ctx context.Context, username Username) (bool, error)
}

// SessionRepository persists refresh-token sessions.
type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	GetByRefreshHash(ctx context.Context, hash string) (*Session, error)
	ListActiveByAccount(ctx context.Context, accountID uuid.UUID, now time.Time, limit, offset int32) ([]*Session, int64, error)
	Revoke(ctx context.Context, id uuid.UUID, now time.Time) error
	RevokeAllByAccount(ctx context.Context, accountID uuid.UUID, now time.Time) (int64, error)
}

// OAuthRepository persists external-identity links.
type OAuthRepository interface {
	Create(ctx context.Context, o *OAuthIdentity) error
	GetByProviderUserID(ctx context.Context, provider OAuthProvider, providerUserID string) (*OAuthIdentity, error)
}

// OutboxRepository implements the transactional outbox: events are enqueued in the same
// transaction as the state change, then drained to Kafka by a relay.
type OutboxRepository interface {
	Enqueue(ctx context.Context, env events.Envelope) error
	FetchUnpublished(ctx context.Context, limit int32) ([]events.Envelope, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID, publishedAt time.Time) error
}

// Store aggregates the repositories and provides transactional execution. WithinTx runs
// fn against a transactional Store; all repository calls made through the supplied Store
// share one transaction and commit or roll back atomically.
type Store interface {
	Accounts() AccountRepository
	Sessions() SessionRepository
	OAuthIdentities() OAuthRepository
	Outbox() OutboxRepository
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Store) error) error
}

// OAuthProviderGateway brokers the provider authorization-code flow. Implementations wrap
// golang.org/x/oauth2 per provider.
type OAuthProviderGateway interface {
	// AuthCodeURL builds the provider consent URL for the given opaque state.
	AuthCodeURL(provider OAuthProvider, state string) (string, error)
	// Exchange swaps an authorization code for the normalized external profile.
	Exchange(ctx context.Context, provider OAuthProvider, code string) (OAuthUserInfo, error)
}

// IDGenerator generates UUID identifiers (v7 for time-ordered primary keys).
type IDGenerator interface {
	New() uuid.UUID
}
