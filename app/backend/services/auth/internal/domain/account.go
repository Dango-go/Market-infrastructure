// Package domain is the heart of the auth bounded context: identity entities, value
// objects, domain errors and the port interfaces the application layer depends on. It
// imports no framework, database or transport package.
package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/google/uuid"
)

// AccountStatus is the lifecycle state of an identity.
type AccountStatus string

const (
	StatusPendingVerification AccountStatus = "pending_verification"
	StatusActive              AccountStatus = "active"
	StatusSuspended           AccountStatus = "suspended"
	StatusDeactivated         AccountStatus = "deactivated"
)

// Valid reports whether the status is a recognized value.
func (s AccountStatus) Valid() bool {
	switch s {
	case StatusPendingVerification, StatusActive, StatusSuspended, StatusDeactivated:
		return true
	default:
		return false
	}
}

// CanAuthenticate reports whether an account in this state may obtain tokens.
func (s AccountStatus) CanAuthenticate() bool {
	return s == StatusActive || s == StatusPendingVerification
}

var (
	emailRe    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	usernameRe = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9_-]{1,30}[a-zA-Z0-9])$`)
)

// Email is a validated, normalized email value object.
type Email string

// NewEmail normalizes (trim + lowercase) and validates an email address.
func NewEmail(raw string) (Email, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" || len(v) > 254 || !emailRe.MatchString(v) {
		return "", ErrInvalidEmail
	}
	return Email(v), nil
}

func (e Email) String() string { return string(e) }

// Username is a validated, normalized handle value object (3–32 chars, url-safe).
type Username string

// NewUsername normalizes (trim) and validates a username handle.
func NewUsername(raw string) (Username, error) {
	v := strings.TrimSpace(raw)
	if !usernameRe.MatchString(v) {
		return "", ErrInvalidUsername
	}
	return Username(v), nil
}

func (u Username) String() string { return string(u) }

// PasswordHash is an opaque, already-hashed credential. The plaintext never enters the
// domain except transiently through the PasswordHasher port.
type PasswordHash string

func (p PasswordHash) String() string { return string(p) }

// Account is the aggregate root of the identity context.
type Account struct {
	ID            uuid.UUID
	Email         Email
	Username      Username
	PasswordHash  PasswordHash // empty for OAuth-only accounts
	Status        AccountStatus
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// NewAccount constructs a freshly-registered, password-backed account in the
// pending-verification state. id and now are injected to keep the domain deterministic.
func NewAccount(id uuid.UUID, email Email, username Username, hash PasswordHash, now time.Time) *Account {
	return &Account{
		ID:            id,
		Email:         email,
		Username:      username,
		PasswordHash:  hash,
		Status:        StatusPendingVerification,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// HasPassword reports whether the account carries a local password credential.
func (a *Account) HasPassword() bool { return a.PasswordHash != "" }

// IsDeleted reports whether the account is soft-deleted.
func (a *Account) IsDeleted() bool { return a.DeletedAt != nil }

// VerifyEmail marks the email verified and promotes a pending account to active.
func (a *Account) VerifyEmail(now time.Time) {
	a.EmailVerified = true
	if a.Status == StatusPendingVerification {
		a.Status = StatusActive
	}
	a.UpdatedAt = now
}

// EnsureCanAuthenticate guards token issuance against suspended/deactivated/deleted
// accounts, returning a domain error otherwise.
func (a *Account) EnsureCanAuthenticate() error {
	if a.IsDeleted() {
		return ErrAccountNotFound
	}
	if !a.Status.CanAuthenticate() {
		return apperr.Forbidden("account_not_active", "this account cannot sign in (%s)", a.Status)
	}
	return nil
}
