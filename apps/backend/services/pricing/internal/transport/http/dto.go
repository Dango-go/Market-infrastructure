package http

type upsertPriceRequest struct {
	ProductID      string `json:"product_id" validate:"required,uuid"`
	Currency       string `json:"currency" validate:"required,max=8"`
	AmountCents    int64  `json:"amount_cents" validate:"required"`
	CompareAtCents int64  `json:"compare_at_cents"`
	Active         bool   `json:"active"`
}

type createPromotionRequest struct {
	Name         string  `json:"name" validate:"required,max=160"`
	Code         string  `json:"code" validate:"required,max=80"`
	DiscountType string  `json:"discount_type" validate:"required,oneof=fixed percent"`
	ValueCents   int64   `json:"value_cents"`
	PercentOff   int     `json:"percent_off"`
	Active       bool    `json:"active"`
	StartsAt     string  `json:"starts_at" validate:"required"`
	EndsAt       *string `json:"ends_at"`
}
