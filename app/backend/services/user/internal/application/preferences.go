package application

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/user/internal/domain"
	"github.com/google/uuid"
)

type PreferencesUseCase struct{ Deps }

func NewPreferencesUseCase(d Deps) *PreferencesUseCase { return &PreferencesUseCase{Deps: d} }

func (uc *PreferencesUseCase) Get(ctx context.Context, accountID uuid.UUID) (PreferencesView, error) {
	prefs, err := uc.Store.Preferences().GetByAccountID(ctx, accountID)
	if err != nil {
		return PreferencesView{}, err
	}
	return preferencesToView(prefs), nil
}

type UpdatePreferencesInput struct {
	Currency          *string
	Language          *string
	EmailNotifications *bool
	SMSNotifications  *bool
	PushNotifications  *bool
	MarketingOptIn    *bool
}

func (uc *PreferencesUseCase) Update(ctx context.Context, accountID uuid.UUID, input UpdatePreferencesInput, req RequestContext) (PreferencesView, error) {
	var out PreferencesView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		prefs, err := tx.Preferences().GetByAccountID(ctx, accountID)
		if err != nil {
			return err
		}
		prefs.Update(input.Currency, input.Language, input.EmailNotifications, input.SMSNotifications, input.PushNotifications, input.MarketingOptIn, uc.Clock.Now())
		if err := tx.Preferences().Update(ctx, prefs); err != nil {
			return err
		}
		envOut, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicUserPreferencesUpdated,
			uc.Source,
			accountID.String(),
			req.CorrelationID,
			uc.Clock.Now(),
			events.UserPreferencesUpdated{
				AccountID:         accountID.String(),
				Currency:          prefs.Currency,
				Language:          prefs.Language,
				EmailNotifications: prefs.EmailNotifications,
				SMSNotifications:  prefs.SMSNotifications,
				PushNotifications: prefs.PushNotifications,
				MarketingOptIn:    prefs.MarketingOptIn,
				UpdatedAt:         prefs.UpdatedAt,
			},
		)
		if err != nil {
			return fmt.Errorf("build preferences updated event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, envOut); err != nil {
			return err
		}
		out = preferencesToView(prefs)
		return nil
	})
	return out, err
}

func preferencesToView(p *domain.Preferences) PreferencesView {
	return PreferencesView{
		AccountID:         p.AccountID,
		Currency:          p.Currency,
		Language:          p.Language,
		EmailNotifications: p.EmailNotifications,
		SMSNotifications:  p.SMSNotifications,
		PushNotifications:  p.PushNotifications,
		MarketingOptIn:    p.MarketingOptIn,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}
