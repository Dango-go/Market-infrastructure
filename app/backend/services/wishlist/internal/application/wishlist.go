package application

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/wishlist/internal/domain"
	"github.com/google/uuid"
)

type WishlistUseCase struct{ Deps }

func NewWishlistUseCase(d Deps) *WishlistUseCase { return &WishlistUseCase{Deps: d} }

func (uc *WishlistUseCase) List(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]WishlistView, int64, error) {
	items, total, err := uc.Store.Wishlist().ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	views := make([]WishlistView, 0, len(items))
	for _, item := range items {
		views = append(views, WishlistView{AccountID: item.AccountID, ProductID: item.ProductID, AddedAt: item.AddedAt})
	}
	return views, total, nil
}

func (uc *WishlistUseCase) Add(ctx context.Context, accountID, productID uuid.UUID, req RequestContext) error {
	return uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		item := domain.NewWishlistItem(accountID, productID, uc.Clock.Now())
		created, err := tx.Wishlist().Add(ctx, item)
		if err != nil {
			return err
		}
		if !created {
			return nil
		}
		envOut, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicUserWishlistUpdated,
			uc.Source,
			accountID.String(),
			req.CorrelationID,
			uc.Clock.Now(),
			events.UserWishlistUpdated{AccountID: accountID.String(), ProductID: productID.String(), Action: "added", UpdatedAt: item.AddedAt},
		)
		if err != nil {
			return fmt.Errorf("build wishlist added event: %w", err)
		}
		return tx.Outbox().Enqueue(ctx, envOut)
	})
}

func (uc *WishlistUseCase) Remove(ctx context.Context, accountID, productID uuid.UUID, req RequestContext) error {
	return uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if err := tx.Wishlist().Remove(ctx, accountID, productID); err != nil {
			return err
		}
		now := uc.Clock.Now()
		envOut, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicUserWishlistUpdated,
			uc.Source,
			accountID.String(),
			req.CorrelationID,
			now,
			events.UserWishlistUpdated{AccountID: accountID.String(), ProductID: productID.String(), Action: "removed", UpdatedAt: now},
		)
		if err != nil {
			return fmt.Errorf("build wishlist removed event: %w", err)
		}
		return tx.Outbox().Enqueue(ctx, envOut)
	})
}
