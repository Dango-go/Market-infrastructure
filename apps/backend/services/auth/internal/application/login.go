package application

import (
	"context"
	"errors"
	"strings"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// LoginUseCase authenticates an identity by email-or-username and password, then
// establishes a session.
type LoginUseCase struct {
	Deps
}

func NewLoginUseCase(d Deps) *LoginUseCase { return &LoginUseCase{Deps: d} }

// Execute resolves the account, verifies the password in constant-ish time, checks the
// account is allowed to sign in, and issues tokens. All credential failures collapse to a
// single opaque error to avoid account enumeration.
func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput, reqCtx RequestContext) (AuthResult, error) {
	account, err := uc.resolve(ctx, in.Identifier)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			// Spend a hash to keep timing uniform whether or not the account exists.
			_, _ = uc.Hasher.Verify("$argon2id$v=19$m=65536,t=3,p=1$AAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", in.Password)
			return AuthResult{}, domain.ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	if !account.HasPassword() {
		return AuthResult{}, domain.ErrInvalidCredentials
	}
	ok, err := uc.Hasher.Verify(account.PasswordHash, in.Password)
	if err != nil {
		return AuthResult{}, err
	}
	if !ok {
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	if err := account.EnsureCanAuthenticate(); err != nil {
		return AuthResult{}, err
	}

	var result AuthResult
	err = uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		res, err := establishSession(ctx, uc.Deps, tx, account, reqCtx, nil)
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	if err != nil {
		return AuthResult{}, err
	}
	return result, nil
}

// resolve looks the account up by email when the identifier looks like an email address,
// otherwise by username.
func (uc *LoginUseCase) resolve(ctx context.Context, identifier string) (*domain.Account, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, apperr.Invalid("missing_identifier", "an email or username is required")
	}
	if strings.Contains(identifier, "@") {
		email, err := domain.NewEmail(identifier)
		if err != nil {
			return nil, domain.ErrInvalidCredentials
		}
		return uc.Store.Accounts().GetByEmail(ctx, email)
	}
	username, err := domain.NewUsername(identifier)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	return uc.Store.Accounts().GetByUsername(ctx, username)
}
