package application

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/user/internal/domain"
	"github.com/google/uuid"
)

type ProfileUseCase struct{ Deps }

func NewProfileUseCase(d Deps) *ProfileUseCase { return &ProfileUseCase{Deps: d} }

func (uc *ProfileUseCase) Get(ctx context.Context, accountID uuid.UUID) (ProfileView, error) {
	p, err := uc.Store.Profiles().GetByAccountID(ctx, accountID)
	if err != nil {
		return ProfileView{}, err
	}
	return profileToView(p), nil
}

type UpdateProfileInput struct {
	DisplayName    *string
	Bio            *string
	Phone          *string
	AvatarURL      *string
	Locale         *string
	Timezone       *string
	MarketingOptIn *bool
}

func (uc *ProfileUseCase) Update(ctx context.Context, accountID uuid.UUID, input UpdateProfileInput, req RequestContext) (ProfileView, error) {
	var out ProfileView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		p, err := tx.Profiles().GetByAccountID(ctx, accountID)
		if err != nil {
			return err
		}
		p.Update(input.DisplayName, input.Bio, input.Phone, input.AvatarURL, input.Locale, input.Timezone, input.MarketingOptIn, uc.Clock.Now())
		if err := tx.Profiles().Update(ctx, p); err != nil {
			return err
		}

		envOut, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicUserProfileUpdated,
			uc.Source,
			accountID.String(),
			req.CorrelationID,
			uc.Clock.Now(),
			events.UserProfileUpdated{
				AccountID:      accountID.String(),
				DisplayName:    p.DisplayName,
				AvatarURL:      p.AvatarURL,
				Locale:         p.Locale,
				Timezone:       p.Timezone,
				MarketingOptIn: p.MarketingOptIn,
				UpdatedAt:      p.UpdatedAt,
			},
		)
		if err != nil {
			return fmt.Errorf("build profile updated event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, envOut); err != nil {
			return err
		}
		out = profileToView(p)
		return nil
	})
	return out, err
}

func profileToView(p *domain.Profile) ProfileView {
	return ProfileView{
		AccountID:      p.AccountID,
		Email:          p.Email,
		Username:       p.Username,
		DisplayName:    p.DisplayName,
		Bio:            p.Bio,
		Phone:          p.Phone,
		AvatarURL:      p.AvatarURL,
		Locale:         p.Locale,
		Timezone:       p.Timezone,
		MarketingOptIn: p.MarketingOptIn,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
