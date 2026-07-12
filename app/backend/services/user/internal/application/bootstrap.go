package application

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/user/internal/domain"
	"github.com/google/uuid"
)

// BootstrapUseCase provisions a local profile from the auth-service user.registered event.
type BootstrapUseCase struct{ Deps }

func NewBootstrapUseCase(d Deps) *BootstrapUseCase { return &BootstrapUseCase{Deps: d} }

func (uc *BootstrapUseCase) Handle(ctx context.Context, env events.Envelope) error {
	var payload events.UserRegistered
	if err := env.Decode(&payload); err != nil {
		return apperr.Invalid("invalid_user_registered_event", "the event payload is invalid").WithCause(err)
	}
	accountID, err := uuid.Parse(payload.AccountID)
	if err != nil {
		return apperr.Invalid("invalid_account_id", "the account id in the event is invalid").WithCause(err)
	}
	registeredAt := payload.RegisteredAt
	if registeredAt.IsZero() {
		registeredAt = uc.Clock.Now()
	}

	return uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		processed, err := tx.ProcessedEvents().Mark(ctx, env.ID, string(env.Type), uc.Clock.Now())
		if err != nil {
			return err
		}
		if !processed {
			return nil
		}

		profile := domain.NewProfile(accountID, payload.Email, payload.Username, registeredAt)
		created, err := tx.Profiles().CreateIfMissing(ctx, profile)
		if err != nil {
			return err
		}

		prefs := domain.NewPreferences(accountID, registeredAt)
		_, err = tx.Preferences().CreateIfMissing(ctx, prefs)
		if err != nil {
			return err
		}

		if !created {
			return nil
		}

		envOut, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicUserProfileCreated,
			uc.Source,
			accountID.String(),
			env.CorrelationID,
			uc.Clock.Now(),
			events.UserProfileCreated{
				AccountID:   accountID.String(),
				Email:       payload.Email,
				Username:    payload.Username,
				DisplayName: profile.DisplayName,
				CreatedAt:   registeredAt,
			},
		)
		if err != nil {
			return fmt.Errorf("build profile created event: %w", err)
		}
		return tx.Outbox().Enqueue(ctx, envOut)
	})
}
