package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrProductNotFound  = apperr.NotFound("product_not_found", "product not found")
	ErrCategoryNotFound = apperr.NotFound("category_not_found", "category not found")
	ErrBrandNotFound    = apperr.NotFound("brand_not_found", "brand not found")
	ErrSlugTaken        = apperr.Conflict("slug_taken", "slug is already in use")
)
