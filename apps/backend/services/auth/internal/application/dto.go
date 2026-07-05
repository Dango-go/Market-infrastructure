// Package application holds the auth use cases — the orchestration of domain entities and
// ports that fulfils each business operation. It depends only on the domain layer and
// shared contract packages, never on Gin, pgx or Kafka.
package application

import (
	"time"

	"github.com/embedded-market/backend/services/auth/internal/domain"
	"github.com/google/uuid"
)

// RequestContext carries cross-cutting request metadata into a use case: the correlation
// id (propagated to emitted events) and client attributes recorded on sessions.
type RequestContext struct {
	CorrelationID string
	UserAgent     string
	IPAddress     string
}

// RegisterInput is the application-level input for account registration.
type RegisterInput struct {
	Email    string
	Username string
	Password string
}

// LoginInput is the application-level input for password login.
type LoginInput struct {
	Identifier string // email or username
	Password   string
}

// RefreshInput carries the opaque refresh token presented for rotation.
type RefreshInput struct {
	RefreshToken string
}

// AccountView is the read-model of an account returned to callers.
type AccountView struct {
	ID            uuid.UUID
	Email         string
	Username      string
	Status        string
	EmailVerified bool
	CreatedAt     time.Time
}

func newAccountView(a *domain.Account) AccountView {
	return AccountView{
		ID:            a.ID,
		Email:         a.Email.String(),
		Username:      a.Username.String(),
		Status:        string(a.Status),
		EmailVerified: a.EmailVerified,
		CreatedAt:     a.CreatedAt,
	}
}

// AuthResult is the outcome of any operation that establishes a session: the access and
// refresh tokens plus the authenticated account.
type AuthResult struct {
	Account               AccountView
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

// SessionView is the read-model of a session for the session-management endpoints.
type SessionView struct {
	ID         uuid.UUID
	UserAgent  string
	IPAddress  string
	CreatedAt  time.Time
	LastUsedAt time.Time
	ExpiresAt  time.Time
	Current    bool
}

func newSessionView(s *domain.Session, currentID uuid.UUID) SessionView {
	return SessionView{
		ID:         s.ID,
		UserAgent:  s.UserAgent,
		IPAddress:  s.IPAddress,
		CreatedAt:  s.CreatedAt,
		LastUsedAt: s.LastUsedAt,
		ExpiresAt:  s.ExpiresAt,
		Current:    s.ID == currentID,
	}
}

// Deps is the dependency set shared by the use cases, injected once at composition time.
type Deps struct {
	Store   domain.Store
	Hasher  domain.PasswordHasher
	Tokens  domain.TokenService
	OAuth   domain.OAuthProviderGateway
	Clock   domain.Clock
	IDs     domain.IDGenerator
	Source  string // event source name, e.g. "auth-service"
}
