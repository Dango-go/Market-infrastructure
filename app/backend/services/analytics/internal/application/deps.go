package application

import "github.com/embedded-market/backend/services/analytics/internal/domain"

type Deps struct {
	Store domain.Store
	Clock domain.Clock
	IDs   IDGenerator
}

type IDGenerator interface { NewUUID() string }
