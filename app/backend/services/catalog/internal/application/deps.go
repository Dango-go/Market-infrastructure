package application

import "github.com/embedded-market/backend/services/catalog/internal/domain"

type Deps struct {
	Store  domain.Store
	Clock  domain.Clock
	IDs    domain.IDGenerator
	Source string
}
