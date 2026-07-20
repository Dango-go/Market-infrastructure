package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/shipping/internal/domain"
	"github.com/google/uuid"
)

type ShippingUseCase struct{ Deps }

func NewShippingUseCase(d Deps) *ShippingUseCase { return &ShippingUseCase{Deps: d} }

type CreateShipmentInput struct {
	OrderID            uuid.UUID
	Carrier            string
	ServiceLevel       string
	TrackingNumber     string
	DestinationAddress string
	Eta                *time.Time
}

type UpdateShipmentStatusInput struct {
	Status         string
	Carrier        string
	TrackingNumber string
	Eta            *time.Time
}

func (uc *ShippingUseCase) Create(ctx context.Context, accountID uuid.UUID, input CreateShipmentInput, req RequestContext) (ShipmentView, error) {
	if strings.TrimSpace(input.Carrier) == "" {
		return ShipmentView{}, apperr.Invalid("invalid_carrier", "carrier is required")
	}
	if strings.TrimSpace(input.ServiceLevel) == "" {
		return ShipmentView{}, apperr.Invalid("invalid_service_level", "service_level is required")
	}
	if strings.TrimSpace(input.DestinationAddress) == "" {
		return ShipmentView{}, apperr.Invalid("invalid_destination_address", "destination_address is required")
	}

	var out ShipmentView
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		now := uc.Clock.Now()
		shipment := domain.NewShipment(
			uc.IDs.New(),
			input.OrderID,
			accountID,
			strings.TrimSpace(input.Carrier),
			strings.TrimSpace(input.ServiceLevel),
			strings.TrimSpace(input.TrackingNumber),
			strings.TrimSpace(input.DestinationAddress),
			input.Eta,
			now,
		)
		if err := tx.Shipments().Create(ctx, shipment); err != nil {
			return err
		}
		env, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicShipmentCreated,
			uc.Source,
			shipment.ID.String(),
			req.CorrelationID,
			now,
			events.ShipmentCreated{
				ShipmentID:      shipment.ID.String(),
				OrderID:         shipment.OrderID.String(),
				AccountID:       shipment.AccountID.String(),
				Status:          string(shipment.Status),
				Carrier:         shipment.Carrier,
				ServiceLevel:    shipment.ServiceLevel,
				TrackingNumber:  shipment.TrackingNumber,
				DestinationAddress: shipment.DestinationAddress,
			},
		)
		if err != nil {
			return fmt.Errorf("build shipment created event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, env); err != nil {
			return err
		}
		out = toShipmentView(shipment)
		return nil
	})
	return out, err
}

func (uc *ShippingUseCase) GetByID(ctx context.Context, accountID, shipmentID uuid.UUID) (ShipmentView, error) {
	shipment, err := uc.Store.Shipments().GetByID(ctx, shipmentID)
	if err != nil {
		return ShipmentView{}, err
	}
	if shipment.AccountID != accountID {
		return ShipmentView{}, domain.ErrShipmentNotFound
	}
	return toShipmentView(shipment), nil
}

func (uc *ShippingUseCase) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]ShipmentView, int64, error) {
	items, total, err := uc.Store.Shipments().ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ShipmentView, 0, len(items))
	for i := range items {
		out = append(out, toShipmentView(&items[i]))
	}
	return out, total, nil
}

func (uc *ShippingUseCase) UpdateStatus(ctx context.Context, accountID, shipmentID uuid.UUID, input UpdateShipmentStatusInput, req RequestContext) (ShipmentView, error) {
	status, err := parseShipmentStatus(input.Status)
	if err != nil {
		return ShipmentView{}, err
	}

	var out ShipmentView
	err = uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		shipment, err := tx.Shipments().GetByID(ctx, shipmentID)
		if err != nil {
			return err
		}
		if shipment.AccountID != accountID {
			return domain.ErrShipmentNotFound
		}
		shipment.UpdateStatus(status, strings.TrimSpace(input.TrackingNumber), strings.TrimSpace(input.Carrier), input.Eta, uc.Clock.Now())
		if err := tx.Shipments().Update(ctx, shipment); err != nil {
			return err
		}
		env, err := events.NewEnvelope(
			uc.IDs.New(),
			events.TopicShipmentStatusUpdated,
			uc.Source,
			shipment.ID.String(),
			req.CorrelationID,
			uc.Clock.Now(),
			events.ShipmentStatusUpdated{
				ShipmentID:      shipment.ID.String(),
				OrderID:         shipment.OrderID.String(),
				AccountID:       shipment.AccountID.String(),
				Status:          string(shipment.Status),
				Carrier:         shipment.Carrier,
				TrackingNumber:  shipment.TrackingNumber,
			},
		)
		if err != nil {
			return fmt.Errorf("build shipment status event: %w", err)
		}
		if err := tx.Outbox().Enqueue(ctx, env); err != nil {
			return err
		}
		out = toShipmentView(shipment)
		return nil
	})
	return out, err
}

func parseShipmentStatus(raw string) (domain.ShipmentStatus, error) {
	status := domain.ShipmentStatus(strings.ToLower(strings.TrimSpace(raw)))
	switch status {
	case domain.ShipmentPending, domain.ShipmentPreparing, domain.ShipmentInTransit, domain.ShipmentDelivered, domain.ShipmentCancelled, domain.ShipmentReturned:
		return status, nil
	default:
		return "", apperr.Invalid("invalid_status", "status must be one of: pending, preparing, in_transit, delivered, cancelled, returned")
	}
}

func toShipmentView(item *domain.Shipment) ShipmentView {
	return ShipmentView{
		ID:                 item.ID,
		OrderID:            item.OrderID,
		AccountID:          item.AccountID,
		Status:             string(item.Status),
		Carrier:            item.Carrier,
		ServiceLevel:       item.ServiceLevel,
		TrackingNumber:     item.TrackingNumber,
		DestinationAddress: item.DestinationAddress,
		Eta:                item.Eta,
		ShippedAt:          item.ShippedAt,
		DeliveredAt:        item.DeliveredAt,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}
