package domain

import "github.com/embedded-market/backend/pkg/apperr"

var (
	ErrDocumentNotFound = apperr.NotFound("search_document_not_found", "search document not found")
)
