package domain

import (
	"time"

	"github.com/google/uuid"
)

type ShipmentStatus string

const (
	ShipmentPending    ShipmentStatus = "pending"
	ShipmentPreparing  ShipmentStatus = "preparing"
	ShipmentInTransit  ShipmentStatus = "in_transit"
	ShipmentDelivered  ShipmentStatus = "delivered"
	ShipmentCancelled  ShipmentStatus = "cancelled"
	ShipmentReturned   ShipmentStatus = "returned"
)

type Shipment struct {
	ID                 uuid.UUID
	OrderID            uuid.UUID
	AccountID          uuid.UUID
	Status             ShipmentStatus
	Carrier            string
	ServiceLevel       string
	TrackingNumber     string
	DestinationAddress string
	Eta                *time.Time
	ShippedAt          *time.Time
	DeliveredAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewShipment(id, orderID, accountID uuid.UUID, carrier, serviceLevel, trackingNumber, destinationAddress string, eta *time.Time, now time.Time) *Shipment {
	return &Shipment{ID: id, OrderID: orderID, AccountID: accountID, Status: ShipmentPending, Carrier: carrier, ServiceLevel: serviceLevel, TrackingNumber: trackingNumber, DestinationAddress: destinationAddress, Eta: eta, CreatedAt: now, UpdatedAt: now}
}

func (s *Shipment) UpdateStatus(status ShipmentStatus, trackingNumber, carrier string, eta *time.Time, now time.Time) {
	s.Status = status
	if trackingNumber != "" { s.TrackingNumber = trackingNumber }
	if carrier != "" { s.Carrier = carrier }
	s.Eta = eta
	if status == ShipmentInTransit && s.ShippedAt == nil {
		shippedAt := now
		s.ShippedAt = &shippedAt
	}
	if status == ShipmentDelivered {
		deliveredAt := now
		s.DeliveredAt = &deliveredAt
	}
	s.UpdatedAt = now
}
