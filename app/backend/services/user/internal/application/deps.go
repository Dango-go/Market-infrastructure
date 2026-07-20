package application

import "github.com/embedded-market/backend/services/user/internal/domain"

// Deps contains the ports shared by all user-service use cases.
type Deps struct {
	Store  domain.Store
	Clock  domain.Clock
	IDs    domain.IDGenerator
	Source string
}
