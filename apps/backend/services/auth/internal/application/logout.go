package application

import (
	"context"
	"errors"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// LogoutUseCase revokes the session associated with a presented refresh token. It is
// idempotent: revoking an unknown or already-revoked token succeeds silently so clients
// can always "log out" without leaking session existence.
type LogoutUseCase struct {
	Deps
}

func NewLogoutUseCase(d Deps) *LogoutUseCase { return &LogoutUseCase{Deps: d} }

// Execute revokes the session matching the refresh token hash, if any.
func (uc *LogoutUseCase) Execute(ctx context.Context, in RefreshInput) error {
	if in.RefreshToken == "" {
		return nil
	}
	hash := uc.Tokens.HashRefreshToken(in.RefreshToken)
	session, err := uc.Store.Sessions().GetByRefreshHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return nil
		}
		return err
	}
	return uc.Store.Sessions().Revoke(ctx, session.ID, uc.Clock.Now())
}
