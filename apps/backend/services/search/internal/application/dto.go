package application

import (
	"time"

	"github.com/google/uuid"
)

type DocumentView struct {
	ID               uuid.UUID `json:"id"`
	ProductID        uuid.UUID `json:"product_id"`
	Slug             string    `json:"slug"`
	SKU              string    `json:"sku"`
	Name             string    `json:"name"`
	ShortDescription string    `json:"short_description"`
	CategorySlug     string    `json:"category_slug"`
	BrandSlug        string    `json:"brand_slug"`
	Tags             []string  `json:"tags"`
	SpecsText        string    `json:"specs_text"`
	PriceCents       int64     `json:"price_cents"`
	Currency         string    `json:"currency"`
	Available        bool      `json:"available"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
