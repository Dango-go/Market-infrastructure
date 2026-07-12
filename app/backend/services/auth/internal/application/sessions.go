package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// SessionsUseCase serves the authenticated session-management endpoints: listing active
// sessions, revoking one, and revoking all others ("sign out everywhere").
type SessionsUseCase struct {
	Deps
}

func NewSessionsUseCase(d Deps) *SessionsUseCase { return &SessionsUseCase{Deps: d} }

// List returns the account's active sessions (paginated), flagging the caller's current
// session so the UI can label it.
func (uc *SessionsUseCase) List(ctx context.Context, accountID, currentSessionID uuid.UUID, limit, offset int32) ([]SessionView, int64, error) {
	sessions, total, err := uc.Store.Sessions().ListActiveByAccount(ctx, accountID, uc.Clock.Now(), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	views := make([]SessionView, 0, len(sessions))
	for _, s := range sessions {
		views = append(views, newSessionView(s, currentSessionID))
	}
	return views, total, nil
}

// Revoke revokes a single session owned by the account. A session belonging to another
// account is reported as not found to avoid leaking its existence.
func (uc *SessionsUseCase) Revoke(ctx context.Context, accountID, sessionID uuid.UUID) error {
	session, err := uc.Store.Sessions().GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if session.AccountID != accountID {
		return domain.ErrSessionNotFound
	}
	return uc.Store.Sessions().Revoke(ctx, sessionID, uc.Clock.Now())
}

// RevokeAll revokes every session for the account (e.g. after a password change or a
// "sign out everywhere" action).
func (uc *SessionsUseCase) RevokeAll(ctx context.Context, accountID uuid.UUID) (int64, error) {
	return uc.Store.Sessions().RevokeAllByAccount(ctx, accountID, uc.Clock.Now())
}
