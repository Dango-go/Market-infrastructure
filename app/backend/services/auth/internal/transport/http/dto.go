// Package http is the auth service transport layer: Gin handlers, request/response DTOs
// and routing. It depends on the application layer and translates between HTTP and use
// cases — no business logic lives here.
package http

import (
	"time"

	"github.com/embedded-market/backend/services/auth/internal/application"
)

// --- Requests -------------------------------------------------------------------------

type registerRequest struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=10,max=128"`
}

type loginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // email or username
	Password   string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// --- Responses ------------------------------------------------------------------------

type accountResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	Status        string    `json:"status"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

func toAccountResponse(v application.AccountView) accountResponse {
	return accountResponse{
		ID:            v.ID.String(),
		Email:         v.Email,
		Username:      v.Username,
		Status:        v.Status,
		EmailVerified: v.EmailVerified,
		CreatedAt:     v.CreatedAt,
	}
}

type authResponse struct {
	Account               accountResponse `json:"account"`
	TokenType             string          `json:"token_type"`
	AccessToken           string          `json:"access_token"`
	AccessTokenExpiresAt  time.Time       `json:"access_token_expires_at"`
	RefreshToken          string          `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time       `json:"refresh_token_expires_at"`
}

func toAuthResponse(r application.AuthResult) authResponse {
	return authResponse{
		Account:               toAccountResponse(r.Account),
		TokenType:             "Bearer",
		AccessToken:           r.AccessToken,
		AccessTokenExpiresAt:  r.AccessTokenExpiresAt,
		RefreshToken:          r.RefreshToken,
		RefreshTokenExpiresAt: r.RefreshTokenExpiresAt,
	}
}

type sessionResponse struct {
	ID         string    `json:"id"`
	UserAgent  string    `json:"user_agent"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Current    bool      `json:"current"`
}

func toSessionResponse(v application.SessionView) sessionResponse {
	return sessionResponse{
		ID:         v.ID.String(),
		UserAgent:  v.UserAgent,
		IPAddress:  v.IPAddress,
		CreatedAt:  v.CreatedAt,
		LastUsedAt: v.LastUsedAt,
		ExpiresAt:  v.ExpiresAt,
		Current:    v.Current,
	}
}
