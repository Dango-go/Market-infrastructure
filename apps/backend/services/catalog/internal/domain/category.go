package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewCategory(id uuid.UUID, name, slug, description string, now time.Time) *Category {
	return &Category{ID: id, Name: strings.TrimSpace(name), Slug: normalizeSlug(slug), Description: strings.TrimSpace(description), CreatedAt: now, UpdatedAt: now}
}
