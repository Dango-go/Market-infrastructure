package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrReviewNotFound      = apperr.NotFound("review_not_found", "review not found")
	ErrReviewAlreadyExists = apperr.Conflict("review_exists", "a review for this product already exists")
)
