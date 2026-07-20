package application

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/user/internal/domain"
	"github.com/google/uuid"
)

type AddressUseCase struct{ Deps }

func NewAddressUseCase(d Deps) *AddressUseCase { return &AddressUseCase{Deps: d} }

func (uc *AddressUseCase) List(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]AddressView, int64, error) {
	items, total, err := uc.Store.Addresses().ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	views := make([]AddressView, 0, len(items))
	for _, item := range items {
		views = append(views, addressToView(item))
	}
	return views, total, nil
}

type CreateAddressInput struct {
	Label             string
	RecipientName     string
	Line1             string
	Line2             string
	City              string
	Region            string
	PostalCode        string
	CountryCode       string
	Phone             string
	IsDefaultShipping bool
	IsDefaultBilling  bool
}

func (uc *AddressUseCase) Create(ctx context.Context, accountID uuid.UUID, input CreateAddressInput, req RequestContext) (AddressView, error) {
	var out AddressView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if input.IsDefaultShipping {
			if err := tx.Addresses().ClearDefaultShipping(ctx, accountID); err != nil {
				return err
			}
		}
		if input.IsDefaultBilling {
			if err := tx.Addresses().ClearDefaultBilling(ctx, accountID); err != nil {
				return err
			}
		}
		addr := domain.NewAddress(uc.IDs.New(), accountID, input.Label, input.RecipientName, input.Line1, input.Line2, input.City, input.Region, input.PostalCode, input.CountryCode, input.Phone, input.IsDefaultShipping, input.IsDefaultBilling, uc.Clock.Now())
		if err := tx.Addresses().Create(ctx, addr); err != nil {
			return err
		}
		envOut, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicUserAddressesUpdated,
			uc.Source,
			accountID.String(),
			req.CorrelationID,
			uc.Clock.Now(),
			events.UserAddressUpdated{
				AccountID:         accountID.String(),
				AddressID:         addr.ID.String(),
				Action:            "created",
				Label:             addr.Label,
				IsDefaultShipping: addr.IsDefaultShipping,
				IsDefaultBilling:  addr.IsDefaultBilling,
				UpdatedAt:         addr.UpdatedAt,
			},
		)
		if err != nil {
			return fmt.Errorf("build address created event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, envOut); err != nil {
			return err
		}
		out = addressToView(addr)
		return nil
	})
	return out, err
}

type UpdateAddressInput struct {
	Label             *string
	RecipientName     *string
	Line1             *string
	Line2             *string
	City              *string
	Region            *string
	PostalCode        *string
	CountryCode       *string
	Phone             *string
	IsDefaultShipping *bool
	IsDefaultBilling  *bool
}

func (uc *AddressUseCase) Update(ctx context.Context, accountID, addressID uuid.UUID, input UpdateAddressInput, req RequestContext) (AddressView, error) {
	var out AddressView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		addr, err := tx.Addresses().GetByID(ctx, addressID)
		if err != nil {
			return err
		}
		if addr.AccountID != accountID {
			return domain.ErrAddressNotFound
		}
		addr.Update(input.Label, input.RecipientName, input.Line1, input.Line2, input.City, input.Region, input.PostalCode, input.CountryCode, input.Phone, input.IsDefaultShipping, input.IsDefaultBilling, uc.Clock.Now())
		if addr.IsDefaultShipping {
			if err := tx.Addresses().ClearDefaultShipping(ctx, accountID); err != nil {
				return err
			}
		}
		if addr.IsDefaultBilling {
			if err := tx.Addresses().ClearDefaultBilling(ctx, accountID); err != nil {
				return err
			}
		}
		if err := tx.Addresses().Update(ctx, addr); err != nil {
			return err
		}
		envOut, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicUserAddressesUpdated,
			uc.Source,
			accountID.String(),
			req.CorrelationID,
			uc.Clock.Now(),
			events.UserAddressUpdated{
				AccountID:         accountID.String(),
				AddressID:         addr.ID.String(),
				Action:            "updated",
				Label:             addr.Label,
				IsDefaultShipping: addr.IsDefaultShipping,
				IsDefaultBilling:  addr.IsDefaultBilling,
				UpdatedAt:         addr.UpdatedAt,
			},
		)
		if err != nil {
			return fmt.Errorf("build address updated event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, envOut); err != nil {
			return err
		}
		out = addressToView(addr)
		return nil
	})
	return out, err
}

func (uc *AddressUseCase) Delete(ctx context.Context, accountID, addressID uuid.UUID, req RequestContext) error {
	return uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		addr, err := tx.Addresses().GetByID(ctx, addressID)
		if err != nil {
			return err
		}
		if addr.AccountID != accountID {
			return domain.ErrAddressNotFound
		}
		if err := tx.Addresses().Delete(ctx, addressID, uc.Clock.Now()); err != nil {
			return err
		}
		envOut, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicUserAddressesUpdated,
			uc.Source,
			accountID.String(),
			req.CorrelationID,
			uc.Clock.Now(),
			events.UserAddressUpdated{AccountID: accountID.String(), AddressID: addressID.String(), Action: "deleted", UpdatedAt: uc.Clock.Now()},
		)
		if err != nil {
			return fmt.Errorf("build address deleted event: %w", err)
		}
		return tx.Outbox().Enqueue(ctx, envOut)
	})
}

func (uc *AddressUseCase) SetDefaultShipping(ctx context.Context, accountID, addressID uuid.UUID, req RequestContext) error {
	return uc.setDefault(ctx, accountID, addressID, true, req)
}

func (uc *AddressUseCase) SetDefaultBilling(ctx context.Context, accountID, addressID uuid.UUID, req RequestContext) error {
	return uc.setDefault(ctx, accountID, addressID, false, req)
}

func (uc *AddressUseCase) setDefault(ctx context.Context, accountID, addressID uuid.UUID, shipping bool, req RequestContext) error {
	return uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		addr, err := tx.Addresses().GetByID(ctx, addressID)
		if err != nil {
			return err
		}
		if addr.AccountID != accountID {
			return domain.ErrAddressNotFound
		}
		if shipping {
			if err := tx.Addresses().ClearDefaultShipping(ctx, accountID); err != nil {
				return err
			}
			addr.IsDefaultShipping = true
		} else {
			if err := tx.Addresses().ClearDefaultBilling(ctx, accountID); err != nil {
				return err
			}
			addr.IsDefaultBilling = true
		}
		addr.UpdatedAt = uc.Clock.Now()
		if err := tx.Addresses().Update(ctx, addr); err != nil {
			return err
		}
		envOut, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicUserAddressesUpdated,
			uc.Source,
			accountID.String(),
			req.CorrelationID,
			uc.Clock.Now(),
			events.UserAddressUpdated{AccountID: accountID.String(), AddressID: addressID.String(), Action: "default_set", IsDefaultShipping: addr.IsDefaultShipping, IsDefaultBilling: addr.IsDefaultBilling, UpdatedAt: addr.UpdatedAt},
		)
		if err != nil {
			return fmt.Errorf("build address default event: %w", err)
		}
		return tx.Outbox().Enqueue(ctx, envOut)
	})
}

func addressToView(a *domain.Address) AddressView {
	return AddressView{
		ID:                a.ID,
		AccountID:         a.AccountID,
		Label:             a.Label,
		RecipientName:     a.RecipientName,
		Line1:             a.Line1,
		Line2:             a.Line2,
		City:              a.City,
		Region:            a.Region,
		PostalCode:        a.PostalCode,
		CountryCode:       a.CountryCode,
		Phone:             a.Phone,
		IsDefaultShipping: a.IsDefaultShipping,
		IsDefaultBilling:  a.IsDefaultBilling,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}
}
