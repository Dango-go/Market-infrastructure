package application

import "github.com/embedded-market/backend/services/review/internal/domain"

type Deps struct {
	Store domain.Store
	Clock domain.Clock
	IDs   domain.IDGenerator
}
