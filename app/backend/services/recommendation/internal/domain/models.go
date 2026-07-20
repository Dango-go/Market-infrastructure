package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProductProfile struct {
	ProductID     uuid.UUID
	Slug          string
	CategorySlug  string
	BrandSlug     string
	Tags          []string
	PriceCents    int64
	Available     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Recommendation struct {
	ProductID uuid.UUID
	Score     float64
	Reason    string
}

type EventType string

const (
	EventView     EventType = "view"
	EventFavorite EventType = "favorite"
	EventCart     EventType = "cart"
	EventPurchase EventType = "purchase"
)

type Event struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	AccountID *uuid.UUID
	Type      EventType
	CreatedAt time.Time
}

func NewProductProfile(productID uuid.UUID, slug, categorySlug, brandSlug string, tags []string, priceCents int64, available bool, now time.Time) *ProductProfile {
	return &ProductProfile{ProductID: productID, Slug: slug, CategorySlug: categorySlug, BrandSlug: brandSlug, Tags: tags, PriceCents: priceCents, Available: available, CreatedAt: now, UpdatedAt: now}
}

func (p *ProductProfile) Update(slug, categorySlug, brandSlug string, tags []string, priceCents int64, available bool, now time.Time) {
	p.Slug = slug
	p.CategorySlug = categorySlug
	p.BrandSlug = brandSlug
	p.Tags = tags
	p.PriceCents = priceCents
	p.Available = available
	p.UpdatedAt = now
}

func NewEvent(id, productID uuid.UUID, accountID *uuid.UUID, eventType EventType, now time.Time) *Event {
	return &Event{ID: id, ProductID: productID, AccountID: accountID, Type: eventType, CreatedAt: now}
}
