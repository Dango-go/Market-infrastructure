package http

type upsertProfileRequest struct {
	ProductID    string   `json:"product_id" validate:"required,uuid"`
	Slug         string   `json:"slug" validate:"required,max=180"`
	CategorySlug string   `json:"category_slug" validate:"omitempty,max=180"`
	BrandSlug    string   `json:"brand_slug" validate:"omitempty,max=180"`
	Tags         []string `json:"tags" validate:"omitempty,dive,max=64"`
	PriceCents   int64    `json:"price_cents"`
	Available    bool     `json:"available"`
}

type trackEventRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Type      string `json:"type" validate:"required,max=32"`
}
