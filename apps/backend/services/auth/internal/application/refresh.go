package application

import (
	"context"
	"errors"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// RefreshUseCase rotates a refresh token: it validates the presented opaque token, revokes
// the old session, and issues a brand-new session and access token. Rotation with reuse
// detection limits the blast radius of a stolen refresh token.
type RefreshUseCase struct {
	Deps
}

func NewRefreshUseCase(d Deps) *RefreshUseCase { return &RefreshUseCase{Deps: d} }

// Execute looks up the session by the token hash. If the matched session is already
// revoked, that signals token reuse: every session for the account is revoked and the
// request is rejected. Otherwise the session is rotated atomically.
func (uc *RefreshUseCase) Execute(ctx context.Context, in RefreshInput, reqCtx RequestContext) (AuthResult, error) {
	if in.RefreshToken == "" {
		return AuthResult{}, domain.ErrInvalidRefresh
	}
	hash := uc.Tokens.HashRefreshToken(in.RefreshToken)

	session, err := uc.Store.Sessions().GetByRefreshHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return AuthResult{}, domain.ErrInvalidRefresh
		}
		return AuthResult{}, err
	}

	now := uc.Clock.Now()

	// Reuse detection: presenting an already-revoked token revokes the whole family.
	if session.RevokedAt != nil {
		_, _ = uc.Store.Sessions().RevokeAllByAccount(ctx, session.AccountID, now)
		return AuthResult{}, domain.ErrSessionRevoked
	}
	if err := session.EnsureUsable(now); err != nil {
		return AuthResult{}, err
	}

	account, err := uc.Store.Accounts().GetByID(ctx, session.AccountID)
	if err != nil {
		return AuthResult{}, err
	}
	if err := account.EnsureCanAuthenticate(); err != nil {
		return AuthResult{}, err
	}

	var result AuthResult
	err = uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if err := tx.Sessions().Revoke(ctx, session.ID, now); err != nil {
			return err
		}
		res, err := establishSession(ctx, uc.Deps, tx, account, reqCtx, &session.ID)
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	if err != nil {
		return AuthResult{}, err
	}
	return result, nil
}
