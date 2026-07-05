package http

type createPaymentRequest struct {
	OrderID        string `json:"order_id" validate:"required,uuid"`
	Provider       string `json:"provider" validate:"required,max=120"`
	Method         string `json:"method" validate:"required,max=80"`
	Currency       string `json:"currency" validate:"required,max=8"`
	AmountCents    int64  `json:"amount_cents" validate:"required"`
	TransactionRef string `json:"transaction_ref" validate:"omitempty,max=160"`
}

type confirmPaymentRequest struct {
	TransactionRef string `json:"transaction_ref" validate:"omitempty,max=160"`
}

type failPaymentRequest struct {
	Reason         string `json:"reason" validate:"required,max=255"`
	TransactionRef string `json:"transaction_ref" validate:"omitempty,max=160"`
}
