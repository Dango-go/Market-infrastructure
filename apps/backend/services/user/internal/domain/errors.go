package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrProfileNotFound    = apperr.NotFound("profile_not_found", "profile not found")
	ErrAddressNotFound    = apperr.NotFound("address_not_found", "address not found")
	ErrPreferencesNotFound = apperr.NotFound("preferences_not_found", "preferences not found")
)
