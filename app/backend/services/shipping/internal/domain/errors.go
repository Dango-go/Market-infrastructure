package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrShipmentNotFound = apperr.NotFound("shipment_not_found", "shipment not found")
)
