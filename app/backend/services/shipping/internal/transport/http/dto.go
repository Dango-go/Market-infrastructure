package http

type createShipmentRequest struct {
	OrderID             string  `json:"order_id" validate:"required,uuid"`
	Carrier             string  `json:"carrier" validate:"required,max=120"`
	ServiceLevel        string  `json:"service_level" validate:"required,max=120"`
	TrackingNumber      string  `json:"tracking_number" validate:"omitempty,max=120"`
	DestinationAddress  string  `json:"destination_address" validate:"required,max=500"`
	Eta                 *string `json:"eta"`
}

type updateShipmentStatusRequest struct {
	Status         string  `json:"status" validate:"required,oneof=pending preparing in_transit delivered cancelled returned"`
	Carrier        string  `json:"carrier" validate:"omitempty,max=120"`
	TrackingNumber string  `json:"tracking_number" validate:"omitempty,max=120"`
	Eta            *string `json:"eta"`
}
