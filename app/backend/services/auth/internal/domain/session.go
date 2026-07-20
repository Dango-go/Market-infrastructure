package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an issued refresh-token grant. The refresh token itself is opaque
// and never stored; only its hash is persisted, so a database leak cannot be replayed.
type Session struct {
	ID                uuid.UUID
	AccountID         uuid.UUID
	RefreshTokenHash  string // SHA-256 of the opaque refresh token
	UserAgent         string
	IPAddress         string
	ExpiresAt         time.Time
	RevokedAt         *time.Time
	RotatedFrom       *uuid.UUID // previous session id when this one is a rotation
	CreatedAt         time.Time
	LastUsedAt        time.Time
}

// NewSession constructs an active session bound to a hashed refresh token.
func NewSession(id, accountID uuid.UUID, refreshHash, userAgent, ip string, now, expiresAt time.Time) *Session {
	return &Session{
		ID:               id,
		AccountID:        accountID,
		RefreshTokenHash: refreshHash,
		UserAgent:        userAgent,
		IPAddress:        ip,
		ExpiresAt:        expiresAt,
		CreatedAt:        now,
		LastUsedAt:       now,
	}
}

// IsActive reports whether the session can still be used at the given instant.
func (s *Session) IsActive(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}

// Revoke marks the session revoked at the given instant (idempotent).
func (s *Session) Revoke(now time.Time) {
	if s.RevokedAt == nil {
		s.RevokedAt = &now
	}
}

// EnsureUsable returns a domain error if the session is revoked or expired.
func (s *Session) EnsureUsable(now time.Time) error {
	if s.RevokedAt != nil {
		return ErrSessionRevoked
	}
	if !now.Before(s.ExpiresAt) {
		return ErrSessionExpired
	}
	return nil
}
