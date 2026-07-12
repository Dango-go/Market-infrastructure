package domain

import (
	"time"

	"github.com/google/uuid"
)

// OAuthProvider identifies a supported external identity provider.
type OAuthProvider string

const (
	ProviderGitHub OAuthProvider = "github"
	ProviderGoogle OAuthProvider = "google"
	ProviderGitLab OAuthProvider = "gitlab"
)

// Valid reports whether the provider is supported.
func (p OAuthProvider) Valid() bool {
	switch p {
	case ProviderGitHub, ProviderGoogle, ProviderGitLab:
		return true
	default:
		return false
	}
}

// OAuthIdentity links an external provider account to a local Account, enabling social
// login. (provider, provider_user_id) is globally unique.
type OAuthIdentity struct {
	ID             uuid.UUID
	AccountID      uuid.UUID
	Provider       OAuthProvider
	ProviderUserID string
	Email          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// OAuthUserInfo is the normalized profile returned by a provider after token exchange.
type OAuthUserInfo struct {
	Provider       OAuthProvider
	ProviderUserID string
	Email          string
	Username       string
	EmailVerified  bool
}

// NewOAuthIdentity constructs a provider link for an account.
func NewOAuthIdentity(id, accountID uuid.UUID, provider OAuthProvider, providerUserID, email string, now time.Time) *OAuthIdentity {
	return &OAuthIdentity{
		ID:             id,
		AccountID:      accountID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		Email:          email,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
