package application

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/cart/internal/domain"
	"github.com/google/uuid"
)

type CartUseCase struct{ Deps }

func NewCartUseCase(d Deps) *CartUseCase { return &CartUseCase{Deps: d} }

type AddItemInput struct {
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
}

func (uc *CartUseCase) GetOrCreateActive(ctx context.Context, accountID uuid.UUID) (CartView, error) {
	cart, err := uc.Store.Carts().GetActiveByAccountID(ctx, accountID)
	if err != nil {
		if err != domain.ErrCartNotFound {
			return CartView{}, err
		}
		now := uc.Clock.Now()
		cart = domain.NewCart(uc.IDs.New(), accountID, "USD", now)
		if err := uc.Store.Carts().Create(ctx, cart); err != nil {
			return CartView{}, err
		}
	}
	items, err := uc.Store.Items().ListByCartID(ctx, cart.ID)
	if err != nil {
		return CartView{}, err
	}
	cart.Recalculate(items, cart.UpdatedAt)
	return toCartView(cart), nil
}

func (uc *CartUseCase) AddItem(ctx context.Context, accountID uuid.UUID, input AddItemInput, req RequestContext) (CartView, error) {
	if input.Quantity <= 0 {
		return CartView{}, apperr.Invalid("invalid_quantity", "quantity must be greater than zero")
	}
	if input.UnitPriceCents < 0 {
		return CartView{}, apperr.Invalid("invalid_unit_price", "unit_price_cents must be zero or greater")
	}

	var out CartView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		cart, err := ensureActiveCart(ctx, tx, uc, accountID)
		if err != nil {
			return err
		}
		item, err := tx.Items().GetByCartProduct(ctx, cart.ID, input.ProductID)
		now := uc.Clock.Now()
		if err != nil {
			if err != domain.ErrCartItemNotFound {
				return err
			}
			item = domain.NewCartItem(uc.IDs.New(), cart.ID, input.ProductID, input.Quantity, input.UnitPriceCents, now)
			if err := tx.Items().Create(ctx, item); err != nil {
				return err
			}
		} else {
			item.Quantity += input.Quantity
			item.UnitPriceCents = input.UnitPriceCents
			item.UpdatedAt = now
			if err := tx.Items().Update(ctx, item); err != nil {
				return err
			}
		}
		return refreshCartAndBuild(ctx, tx, uc, cart, req, &out)
	})
	return out, err
}

func (uc *CartUseCase) UpdateItem(ctx context.Context, accountID, productID uuid.UUID, quantity int32, req RequestContext) (CartView, error) {
	var out CartView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		cart, err := tx.Carts().GetActiveByAccountID(ctx, accountID)
		if err != nil {
			return err
		}
		item, err := tx.Items().GetByCartProduct(ctx, cart.ID, productID)
		if err != nil {
			return err
		}
		if quantity <= 0 {
			if err := tx.Items().Delete(ctx, item.ID); err != nil {
				return err
			}
		} else {
			item.Quantity = quantity
			item.UpdatedAt = uc.Clock.Now()
			if err := tx.Items().Update(ctx, item); err != nil {
				return err
			}
		}
		return refreshCartAndBuild(ctx, tx, uc, cart, req, &out)
	})
	return out, err
}

func (uc *CartUseCase) Clear(ctx context.Context, accountID uuid.UUID, req RequestContext) (CartView, error) {
	var out CartView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		cart, err := tx.Carts().GetActiveByAccountID(ctx, accountID)
		if err != nil {
			return err
		}
		if err := tx.Items().DeleteByCartID(ctx, cart.ID); err != nil {
			return err
		}
		return refreshCartAndBuild(ctx, tx, uc, cart, req, &out)
	})
	return out, err
}

func (uc *CartUseCase) Checkout(ctx context.Context, accountID uuid.UUID, req RequestContext) (CartView, error) {
	var out CartView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		cart, err := tx.Carts().GetActiveByAccountID(ctx, accountID)
		if err != nil {
			return err
		}
		items, err := tx.Items().ListByCartID(ctx, cart.ID)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return apperr.Invalid("empty_cart", "cannot checkout an empty cart")
		}
		now := uc.Clock.Now()
		cart.Recalculate(items, now)
		cart.Checkout(now)
		if err := tx.Carts().Update(ctx, cart); err != nil {
			return err
		}
		payloadItems := make([]events.CartCheckedOutItem, 0, len(items))
		for _, item := range items {
			payloadItems = append(payloadItems, events.CartCheckedOutItem{
				ProductID:      item.ProductID.String(),
				Quantity:       item.Quantity,
				UnitPriceCents: item.UnitPriceCents,
			})
		}
		env, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicCartCheckedOut,
			uc.Source,
			cart.ID.String(),
			req.CorrelationID,
			now,
			events.CartCheckedOut{
				CartID:        cart.ID.String(),
				AccountID:     cart.AccountID.String(),
				Currency:      cart.Currency,
				SubtotalCents: cart.SubtotalCents,
				Items:         payloadItems,
			},
		)
		if err != nil {
			return fmt.Errorf("build cart checked out event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, env); err != nil {
			return err
		}
		out = toCartView(cart)
		return nil
	})
	return out, err
}

func ensureActiveCart(ctx context.Context, tx domain.Store, uc *CartUseCase, accountID uuid.UUID) (*domain.Cart, error) {
	cart, err := tx.Carts().GetActiveByAccountID(ctx, accountID)
	if err != nil {
		if err != domain.ErrCartNotFound {
			return nil, err
		}
		now := uc.Clock.Now()
		cart = domain.NewCart(uc.IDs.New(), accountID, "USD", now)
		if err := tx.Carts().Create(ctx, cart); err != nil {
			return nil, err
		}
	}
	return cart, nil
}

func refreshCartAndBuild(ctx context.Context, tx domain.Store, uc *CartUseCase, cart *domain.Cart, req RequestContext, out *CartView) error {
	items, err := tx.Items().ListByCartID(ctx, cart.ID)
	if err != nil {
		return err
	}
	cart.Recalculate(items, uc.Clock.Now())
	if err := tx.Carts().Update(ctx, cart); err != nil {
		return err
	}
	*out = toCartView(cart)
	return nil
}

func toCartView(cart *domain.Cart) CartView {
	items := make([]CartItemView, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, CartItemView{
			ID:             item.ID,
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: int64(item.Quantity) * item.UnitPriceCents,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return CartView{
		ID:            cart.ID,
		AccountID:     cart.AccountID,
		Status:        string(cart.Status),
		Currency:      cart.Currency,
		SubtotalCents: cart.SubtotalCents,
		CreatedAt:     cart.CreatedAt,
		UpdatedAt:     cart.UpdatedAt,
		CheckedOutAt:  cart.CheckedOutAt,
		Items:         items,
	}
}
