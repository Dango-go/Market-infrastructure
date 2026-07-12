package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrProfileNotFound = apperr.NotFound("recommendation_profile_not_found", "recommendation profile not found")
)
