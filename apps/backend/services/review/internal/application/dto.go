package application

import (
	"time"

	"github.com/google/uuid"
)

type ReviewView struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	AccountID uuid.UUID `json:"account_id"`
	Rating    int32     `json:"rating"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReviewSummaryView struct {
	ProductID     uuid.UUID `json:"product_id"`
	TotalReviews  int64     `json:"total_reviews"`
	AverageRating float64   `json:"average_rating"`
	Rating1Count  int64     `json:"rating_1_count"`
	Rating2Count  int64     `json:"rating_2_count"`
	Rating3Count  int64     `json:"rating_3_count"`
	Rating4Count  int64     `json:"rating_4_count"`
	Rating5Count  int64     `json:"rating_5_count"`
}
