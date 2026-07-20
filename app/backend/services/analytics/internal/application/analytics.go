package application

import (
	"context"
	"strings"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/services/analytics/internal/domain"
	"github.com/google/uuid"
)

type AnalyticsUseCase struct{ Deps }

func NewAnalyticsUseCase(d Deps) *AnalyticsUseCase { return &AnalyticsUseCase{Deps: d} }

type TrackEventInput struct {
	AccountID *uuid.UUID
	SessionID string
	ProductID *uuid.UUID
	EventType string
	Path      string
	Referrer  string
	Query     string
	UserAgent string
}

func (uc *AnalyticsUseCase) TrackEvent(ctx context.Context, input TrackEventInput) error {
	eventType := domain.EventType(strings.TrimSpace(input.EventType))
	if !isValidEventType(eventType) {
		return domain.ErrInvalidEventType
	}
	if strings.TrimSpace(input.SessionID) == "" {
		return apperr.Invalid("invalid_session_id", "session_id is required")
	}
	eventID, err := uuid.Parse(uc.IDs.NewUUID())
	if err != nil {
		return apperr.Internal("invalid_generated_id", "failed to generate analytics event id")
	}
	return uc.Store.Events().Create(ctx, domain.NewEvent(eventID, input.AccountID, strings.TrimSpace(input.SessionID), input.ProductID, eventType, strings.TrimSpace(input.Path), strings.TrimSpace(input.Referrer), strings.TrimSpace(input.Query), strings.TrimSpace(input.UserAgent), uc.Clock.Now()))
}

func (uc *AnalyticsUseCase) Overview(ctx context.Context, days int32) (OverviewView, error) {
	if days < 1 || days > 365 {
		return OverviewView{}, apperr.Invalid("invalid_days", "days must be between 1 and 365")
	}
	item, err := uc.Store.Events().Overview(ctx, days)
	if err != nil {
		return OverviewView{}, err
	}
	return OverviewView(item), nil
}

func (uc *AnalyticsUseCase) TopProducts(ctx context.Context, days, limit int32) ([]TopProductView, error) {
	if days < 1 || days > 365 {
		return nil, apperr.Invalid("invalid_days", "days must be between 1 and 365")
	}
	if limit < 1 {
		return nil, apperr.Invalid("invalid_limit", "limit must be greater than zero")
	}
	items, err := uc.Store.Events().TopProducts(ctx, days, limit)
	if err != nil {
		return nil, err
	}
	out := make([]TopProductView, 0, len(items))
	for _, item := range items {
		out = append(out, TopProductView(item))
	}
	return out, nil
}

func isValidEventType(eventType domain.EventType) bool {
	switch eventType {
	case domain.EventPageView, domain.EventProductView, domain.EventSearch, domain.EventAddToCart, domain.EventBeginCheckout, domain.EventPurchase:
		return true
	default:
		return false
	}
}
