package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Brand struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	CountryCode string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewBrand(id uuid.UUID, name, slug, description, countryCode string, now time.Time) *Brand {
	return &Brand{ID: id, Name: strings.TrimSpace(name), Slug: normalizeSlug(slug), Description: strings.TrimSpace(description), CountryCode: strings.ToUpper(strings.TrimSpace(countryCode)), CreatedAt: now, UpdatedAt: now}
}
