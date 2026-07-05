package http

type upsertDocumentRequest struct {
    ID               *string  `json:"id" validate:"omitempty"`
    ProductID        string   `json:"product_id" validate:"required,uuid"`
    Slug             string   `json:"slug" validate:"required,max=180"`
    SKU              string   `json:"sku" validate:"omitempty,max=120"`
    Name             string   `json:"name" validate:"required,max=255"`
    ShortDescription string   `json:"short_description" validate:"omitempty,max=500"`
    CategorySlug     string   `json:"category_slug" validate:"omitempty,max=180"`
    BrandSlug        string   `json:"brand_slug" validate:"omitempty,max=180"`
    Tags             []string `json:"tags" validate:"omitempty,dive,max=64"`
    SpecsText        string   `json:"specs_text" validate:"omitempty,max=4000"`
    PriceCents       int64    `json:"price_cents"`
    Currency         string   `json:"currency" validate:"required,max=8"`
    Available        bool     `json:"available"`
}
