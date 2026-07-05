package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	AccountID uuid.UUID
	Rating    int32
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ReviewSummary struct {
	ProductID     uuid.UUID
	TotalReviews  int64
	AverageRating float64
	Rating1Count  int64
	Rating2Count  int64
	Rating3Count  int64
	Rating4Count  int64
	Rating5Count  int64
}

func NewReview(id, productID, accountID uuid.UUID, rating int32, title, body string, now time.Time) *Review {
	return &Review{ID: id, ProductID: productID, AccountID: accountID, Rating: rating, Title: title, Body: body, CreatedAt: now, UpdatedAt: now}
}

func (r *Review) Update(rating int32, title, body string, now time.Time) {
	r.Rating = rating
	r.Title = title
	r.Body = body
	r.UpdatedAt = now
}
