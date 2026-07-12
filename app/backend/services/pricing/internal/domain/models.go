package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Price struct {
	ID             uuid.UUID
	ProductID      uuid.UUID
	Currency       string
	AmountCents    int64
	CompareAtCents int64
	Active         bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewPrice(id, productID uuid.UUID, currency string, amountCents, compareAtCents int64, active bool, now time.Time) *Price {
	return &Price{
		ID:             id,
		ProductID:      productID,
		Currency:       strings.ToUpper(strings.TrimSpace(currency)),
		AmountCents:    amountCents,
		CompareAtCents: compareAtCents,
		Active:         active,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (p *Price) Update(currency string, amountCents, compareAtCents int64, active bool, now time.Time) {
	p.Currency = strings.ToUpper(strings.TrimSpace(currency))
	p.AmountCents = amountCents
	p.CompareAtCents = compareAtCents
	p.Active = active
	p.UpdatedAt = now
}

type Promotion struct {
	ID           uuid.UUID
	Name         string
	Code         string
	DiscountType string
	ValueCents   int64
	PercentOff   int
	Active       bool
	StartsAt     time.Time
	EndsAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewPromotion(id uuid.UUID, name, code, discountType string, valueCents int64, percentOff int, active bool, startsAt time.Time, endsAt *time.Time, now time.Time) *Promotion {
	return &Promotion{
		ID:           id,
		Name:         strings.TrimSpace(name),
		Code:         strings.ToUpper(strings.TrimSpace(code)),
		DiscountType: strings.TrimSpace(discountType),
		ValueCents:   valueCents,
		PercentOff:   percentOff,
		Active:       active,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
