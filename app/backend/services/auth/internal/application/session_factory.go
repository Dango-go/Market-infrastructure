package application

import (
	"context"

	"github.com/embedded-market/backend/services/auth/internal/domain"
	"github.com/google/uuid"
)

// establishSession mints a refresh-token session and access token for an account, writing
// the session through the supplied (possibly transactional) store. It is the single place
// that defines what "being signed in" means, reused by registration, login, refresh and
// OAuth. The access token is signed in-memory (no I/O); only the session row is written.
// rotatedFrom records the predecessor session when this session is the product of a
// refresh-token rotation; pass nil for a fresh sign-in.
func establishSession(ctx context.Context, d Deps, store domain.Store, account *domain.Account, reqCtx RequestContext, rotatedFrom *uuid.UUID) (AuthResult, error) {
	now := d.Clock.Now()

	plaintextRefresh, refreshHash, err := d.Tokens.GenerateRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}

	sessionID := d.IDs.New()
	expiresAt := now.Add(d.Tokens.RefreshTTL())
	session := domain.NewSession(sessionID, account.ID, refreshHash, reqCtx.UserAgent, reqCtx.IPAddress, now, expiresAt)
	session.RotatedFrom = rotatedFrom
	if err := store.Sessions().Create(ctx, session); err != nil {
		return AuthResult{}, err
	}

	access, err := d.Tokens.IssueAccessToken(account, sessionID, now)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		Account:               newAccountView(account),
		AccessToken:           access.Value,
		AccessTokenExpiresAt:  access.ExpiresAt,
		RefreshToken:          plaintextRefresh,
		RefreshTokenExpiresAt: expiresAt,
	}, nil
}
