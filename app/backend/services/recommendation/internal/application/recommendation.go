package application

import (
	"context"
	"strings"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/services/recommendation/internal/domain"
	"github.com/google/uuid"
)

type RecommendationUseCase struct{ Deps }

func NewRecommendationUseCase(d Deps) *RecommendationUseCase { return &RecommendationUseCase{Deps: d} }

type UpsertProfileInput struct {
	ProductID    uuid.UUID
	Slug         string
	CategorySlug string
	BrandSlug    string
	Tags         []string
	PriceCents   int64
	Available    bool
}

type TrackEventInput struct {
	ProductID uuid.UUID
	AccountID *uuid.UUID
	Type      string
}

func (uc *RecommendationUseCase) UpsertProfile(ctx context.Context, input UpsertProfileInput) (ProductProfileView, error) {
	if strings.TrimSpace(input.Slug) == "" {
		return ProductProfileView{}, apperr.Invalid("invalid_slug", "slug is required")
	}
	if input.PriceCents < 0 {
		return ProductProfileView{}, apperr.Invalid("invalid_price", "price_cents must be zero or greater")
	}
	now := uc.Clock.Now()
	profile, err := uc.Store.Profiles().GetByProductID(ctx, input.ProductID)
	if err != nil && err != domain.ErrProfileNotFound {
		return ProductProfileView{}, err
	}
	if err == domain.ErrProfileNotFound {
		profile = domain.NewProductProfile(input.ProductID, strings.TrimSpace(input.Slug), strings.TrimSpace(input.CategorySlug), strings.TrimSpace(input.BrandSlug), cleanTags(input.Tags), input.PriceCents, input.Available, now)
	} else {
		profile.Update(strings.TrimSpace(input.Slug), strings.TrimSpace(input.CategorySlug), strings.TrimSpace(input.BrandSlug), cleanTags(input.Tags), input.PriceCents, input.Available, now)
	}
	if err := uc.Store.Profiles().Upsert(ctx, profile); err != nil {
		return ProductProfileView{}, err
	}
	return toProfileView(profile), nil
}

func (uc *RecommendationUseCase) Related(ctx context.Context, productID uuid.UUID, limit int32) ([]RecommendationView, error) {
	items, err := uc.Store.Profiles().Related(ctx, productID, limit)
	if err != nil {
		return nil, err
	}
	return toRecommendationViews(items), nil
}

func (uc *RecommendationUseCase) Trending(ctx context.Context, limit int32) ([]RecommendationView, error) {
	items, err := uc.Store.Profiles().Trending(ctx, limit)
	if err != nil {
		return nil, err
	}
	return toRecommendationViews(items), nil
}

func (uc *RecommendationUseCase) TrackEvent(ctx context.Context, input TrackEventInput) error {
	eventType := domain.EventType(strings.TrimSpace(input.Type))
	if eventType != domain.EventView && eventType != domain.EventFavorite && eventType != domain.EventCart && eventType != domain.EventPurchase {
		return apperr.Invalid("invalid_event_type", "event type must be one of: view, favorite, cart, purchase")
	}
	return uc.Store.Events().Create(ctx, domain.NewEvent(uc.IDs.New(), input.ProductID, input.AccountID, eventType, uc.Clock.Now()))
}

func cleanTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		if value := strings.TrimSpace(tag); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func toRecommendationViews(items []domain.Recommendation) []RecommendationView {
	out := make([]RecommendationView, 0, len(items))
	for _, item := range items {
		out = append(out, RecommendationView{ProductID: item.ProductID, Score: item.Score, Reason: item.Reason})
	}
	return out
}

func toProfileView(item *domain.ProductProfile) ProductProfileView {
	return ProductProfileView{ProductID: item.ProductID, Slug: item.Slug, CategorySlug: item.CategorySlug, BrandSlug: item.BrandSlug, Tags: item.Tags, PriceCents: item.PriceCents, Available: item.Available}
}
