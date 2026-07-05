package application

import (
	"time"

	"github.com/google/uuid"
)

type RequestContext struct {
	CorrelationID string
	AccountID     uuid.UUID
}

type ShipmentView struct {
	ID                 uuid.UUID  `json:"id"`
	OrderID            uuid.UUID  `json:"order_id"`
	AccountID          uuid.UUID  `json:"account_id"`
	Status             string     `json:"status"`
	Carrier            string     `json:"carrier"`
	ServiceLevel       string     `json:"service_level"`
	TrackingNumber     string     `json:"tracking_number,omitempty"`
	DestinationAddress string     `json:"destination_address"`
	Eta                *time.Time `json:"eta,omitempty"`
	ShippedAt          *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt        *time.Time `json:"delivered_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
