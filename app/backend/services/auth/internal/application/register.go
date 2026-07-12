package application

import (
	"context"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// RegisterUseCase creates a new password-backed identity, establishes an initial session
// and emits user.registered via the transactional outbox.
type RegisterUseCase struct {
	Deps
}

func NewRegisterUseCase(d Deps) *RegisterUseCase { return &RegisterUseCase{Deps: d} }

// Execute validates and normalizes the input, enforces uniqueness and the password
// policy, then atomically persists the account, its first session and the outbox event.
func (uc *RegisterUseCase) Execute(ctx context.Context, in RegisterInput, reqCtx RequestContext) (AuthResult, error) {
	email, err := domain.NewEmail(in.Email)
	if err != nil {
		return AuthResult{}, err
	}
	username, err := domain.NewUsername(in.Username)
	if err != nil {
		return AuthResult{}, err
	}
	if err := domain.ValidatePassword(in.Password); err != nil {
		return AuthResult{}, err
	}

	// Fast-path uniqueness checks for friendly errors; the unique constraints in the
	// repository remain the source of truth under concurrency.
	if taken, err := uc.Store.Accounts().ExistsByEmail(ctx, email); err != nil {
		return AuthResult{}, err
	} else if taken {
		return AuthResult{}, domain.ErrEmailTaken
	}
	if taken, err := uc.Store.Accounts().ExistsByUsername(ctx, username); err != nil {
		return AuthResult{}, err
	} else if taken {
		return AuthResult{}, domain.ErrUsernameTaken
	}

	hash, err := uc.Hasher.Hash(in.Password)
	if err != nil {
		return AuthResult{}, err
	}

	now := uc.Clock.Now()
	account := domain.NewAccount(uc.IDs.New(), email, username, hash, now)

	var result AuthResult
	err = uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if err := tx.Accounts().Create(ctx, account); err != nil {
			return err
		}
		res, err := establishSession(ctx, uc.Deps, tx, account, reqCtx, nil)
		if err != nil {
			return err
		}
		result = res

		env, err := events.NewEnvelope(
			uc.IDs.New(), events.TopicUserRegistered, uc.Source,
			account.ID.String(), reqCtx.CorrelationID, now,
			events.UserRegistered{
				AccountID:    account.ID.String(),
				Email:        account.Email.String(),
				Username:     account.Username.String(),
				RegisteredAt: now,
			},
		)
		if err != nil {
			return err
		}
		return tx.Outbox().Enqueue(ctx, env)
	})
	if err != nil {
		return AuthResult{}, err
	}
	return result, nil
}
