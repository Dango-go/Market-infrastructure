package application

import (
	"context"
	"fmt"
	"time"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/pricing/internal/domain"
	"github.com/google/uuid"
)

type PricingUseCase struct{ Deps }

func NewPricingUseCase(d Deps) *PricingUseCase { return &PricingUseCase{Deps: d} }

type UpsertPriceInput struct {
	ProductID      uuid.UUID
	Currency       string
	AmountCents    int64
	CompareAtCents int64
	Active         bool
}

type CreatePromotionInput struct {
	Name         string
	Code         string
	DiscountType string
	ValueCents   int64
	PercentOff   int
	Active       bool
	StartsAt     time.Time
	EndsAt       *time.Time
}

func (uc *PricingUseCase) UpsertPrice(ctx context.Context, input UpsertPriceInput, req RequestContext) (PriceView, error) {
	var out PriceView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		price, err := tx.Prices().GetByProductID(ctx, input.ProductID)
		now := uc.Clock.Now()
		if err != nil {
			if err != domain.ErrPriceNotFound {
				return err
			}
			price = domain.NewPrice(uc.IDs.New(), input.ProductID, input.Currency, input.AmountCents, input.CompareAtCents, input.Active, now)
			if err := tx.Prices().Create(ctx, price); err != nil {
				return err
			}
		} else {
			price.Update(input.Currency, input.AmountCents, input.CompareAtCents, input.Active, now)
			if err := tx.Prices().Update(ctx, price); err != nil {
				return err
			}
		}
		env, err := events.NewEnvelope(uc.IDs.New(), events.TopicPriceChanged, uc.Source, price.ProductID.String(), req.CorrelationID, now, events.PriceChanged{
			ProductID:      price.ProductID.String(),
			Currency:       price.Currency,
			AmountCents:    price.AmountCents,
			CompareAtCents: price.CompareAtCents,
			Active:         price.Active,
		})
		if err != nil {
			return fmt.Errorf("build price changed event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, env); err != nil {
			return err
		}
		out = toPriceView(price)
		return nil
	})
	return out, err
}

func (uc *PricingUseCase) GetPrice(ctx context.Context, productID uuid.UUID) (PriceView, error) {
	item, err := uc.Store.Prices().GetByProductID(ctx, productID)
	if err != nil {
		return PriceView{}, err
	}
	return toPriceView(item), nil
}

func (uc *PricingUseCase) ListPrices(ctx context.Context, limit, offset int32) ([]PriceView, int64, error) {
	items, total, err := uc.Store.Prices().List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]PriceView, 0, len(items))
	for _, item := range items {
		out = append(out, toPriceView(item))
	}
	return out, total, nil
}

func (uc *PricingUseCase) CreatePromotion(ctx context.Context, input CreatePromotionInput, req RequestContext) (PromotionView, error) {
	var out PromotionView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		item := domain.NewPromotion(uc.IDs.New(), input.Name, input.Code, input.DiscountType, input.ValueCents, input.PercentOff, input.Active, input.StartsAt, input.EndsAt, uc.Clock.Now())
		if err := tx.Promotions().Create(ctx, item); err != nil {
			return err
		}
		if item.Active {
			env, err := events.NewEnvelope(uc.IDs.New(), events.TopicPromotionStarted, uc.Source, item.ID.String(), req.CorrelationID, uc.Clock.Now(), events.PromotionStarted{
				PromotionID:   item.ID.String(),
				Code:          item.Code,
				DiscountType:  item.DiscountType,
				ValueCents:    item.ValueCents,
				PercentOff:    item.PercentOff,
				StartsAt:      item.StartsAt,
			})
			if err != nil {
				return fmt.Errorf("build promotion started event: %w", err)
			}
			if err := tx.Outbox().Enqueue(ctx, env); err != nil {
				return err
			}
		}
		out = toPromotionView(item)
		return nil
	})
	return out, err
}

func (uc *PricingUseCase) ListPromotions(ctx context.Context, limit, offset int32) ([]PromotionView, int64, error) {
	items, total, err := uc.Store.Promotions().List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]PromotionView, 0, len(items))
	for _, item := range items {
		out = append(out, toPromotionView(item))
	}
	return out, total, nil
}

func toPriceView(p *domain.Price) PriceView {
	return PriceView{
		ID:             p.ID,
		ProductID:      p.ProductID,
		Currency:       p.Currency,
		AmountCents:    p.AmountCents,
		CompareAtCents: p.CompareAtCents,
		Active:         p.Active,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func toPromotionView(p *domain.Promotion) PromotionView {
	return PromotionView{
		ID:           p.ID,
		Name:         p.Name,
		Code:         p.Code,
		DiscountType: p.DiscountType,
		ValueCents:   p.ValueCents,
		PercentOff:   p.PercentOff,
		Active:       p.Active,
		StartsAt:     p.StartsAt,
		EndsAt:       p.EndsAt,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

