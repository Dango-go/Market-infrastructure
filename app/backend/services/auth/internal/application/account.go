package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// AccountUseCase serves authenticated identity queries (the "who am I" endpoint).
type AccountUseCase struct {
	Deps
}

func NewAccountUseCase(d Deps) *AccountUseCase { return &AccountUseCase{Deps: d} }

// Get returns the current account view by id.
func (uc *AccountUseCase) Get(ctx context.Context, id uuid.UUID) (AccountView, error) {
	account, err := uc.Store.Accounts().GetByID(ctx, id)
	if err != nil {
		return AccountView{}, err
	}
	if account.IsDeleted() {
		return AccountView{}, domain.ErrAccountNotFound
	}
	return newAccountView(account), nil
}
