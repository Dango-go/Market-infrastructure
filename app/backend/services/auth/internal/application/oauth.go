package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// OAuthUseCase implements social login over the authorization-code flow: it builds the
// provider consent URL and, on callback, exchanges the code, then links or provisions a
// local account before establishing a session.
type OAuthUseCase struct {
	Deps
}

func NewOAuthUseCase(d Deps) *OAuthUseCase { return &OAuthUseCase{Deps: d} }

// Begin validates the provider and returns the consent URL for the supplied opaque state.
// CSRF protection (binding state to the user agent) is enforced at the transport layer.
func (uc *OAuthUseCase) Begin(provider domain.OAuthProvider, state string) (string, error) {
	if !provider.Valid() {
		return "", domain.ErrUnsupportedProvider
	}
	url, err := uc.OAuth.AuthCodeURL(provider, state)
	if err != nil {
		return "", domain.ErrOAuthExchange
	}
	return url, nil
}

// Complete exchanges the authorization code, finds or provisions the local account and
// its provider link, and issues tokens.
func (uc *OAuthUseCase) Complete(ctx context.Context, provider domain.OAuthProvider, code string, reqCtx RequestContext) (AuthResult, error) {
	if !provider.Valid() {
		return AuthResult{}, domain.ErrUnsupportedProvider
	}
	info, err := uc.OAuth.Exchange(ctx, provider, code)
	if err != nil {
		return AuthResult{}, domain.ErrOAuthExchange
	}

	// Already linked: authenticate the existing account.
	identity, err := uc.Store.OAuthIdentities().GetByProviderUserID(ctx, provider, info.ProviderUserID)
	if err != nil && !errors.Is(err, domain.ErrOAuthIdentityNotFound) {
		return AuthResult{}, err
	}
	if identity != nil {
		account, err := uc.Store.Accounts().GetByID(ctx, identity.AccountID)
		if err != nil {
			return AuthResult{}, err
		}
		if err := account.EnsureCanAuthenticate(); err != nil {
			return AuthResult{}, err
		}
		return uc.establish(ctx, account, reqCtx)
	}

	// Link to an existing account with the same verified email, if present.
	if email, emailErr := domain.NewEmail(info.Email); emailErr == nil {
		existing, getErr := uc.Store.Accounts().GetByEmail(ctx, email)
		if getErr != nil && !errors.Is(getErr, domain.ErrAccountNotFound) {
			return AuthResult{}, getErr
		}
		if existing != nil {
			return uc.linkAndAuthenticate(ctx, existing, provider, info, reqCtx)
		}
	}

	// Otherwise provision a fresh OAuth-only account.
	return uc.provision(ctx, provider, info, reqCtx)
}

func (uc *OAuthUseCase) linkAndAuthenticate(ctx context.Context, account *domain.Account, provider domain.OAuthProvider, info domain.OAuthUserInfo, reqCtx RequestContext) (AuthResult, error) {
	if err := account.EnsureCanAuthenticate(); err != nil {
		return AuthResult{}, err
	}
	now := uc.Clock.Now()
	identity := domain.NewOAuthIdentity(uc.IDs.New(), account.ID, provider, info.ProviderUserID, info.Email, now)

	var result AuthResult
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if err := tx.OAuthIdentities().Create(ctx, identity); err != nil {
			return err
		}
		res, err := establishSession(ctx, uc.Deps, tx, account, reqCtx, nil)
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	return result, err
}

func (uc *OAuthUseCase) provision(ctx context.Context, provider domain.OAuthProvider, info domain.OAuthUserInfo, reqCtx RequestContext) (AuthResult, error) {
	email, err := domain.NewEmail(info.Email)
	if err != nil {
		return AuthResult{}, domain.ErrInvalidEmail
	}
	username, err := uc.uniqueUsername(ctx, info.Username, info.Email)
	if err != nil {
		return AuthResult{}, err
	}

	now := uc.Clock.Now()
	account := domain.NewAccount(uc.IDs.New(), email, username, "", now)
	if info.EmailVerified {
		account.VerifyEmail(now) // promotes to active
	}
	identity := domain.NewOAuthIdentity(uc.IDs.New(), account.ID, provider, info.ProviderUserID, info.Email, now)

	var result AuthResult
	err = uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		if err := tx.Accounts().Create(ctx, account); err != nil {
			return err
		}
		if err := tx.OAuthIdentities().Create(ctx, identity); err != nil {
			return err
		}
		res, err := establishSession(ctx, uc.Deps, tx, account, reqCtx, nil)
		if err != nil {
			return err
		}
		result = res

		env, err := events.NewEnvelope(
			uc.IDs.New(), events.TopicUserRegistered, uc.Source,
			account.ID.String(), reqCtx.CorrelationID, now,
			events.UserRegistered{
				AccountID:    account.ID.String(),
				Email:        account.Email.String(),
				Username:     account.Username.String(),
				RegisteredAt: now,
			},
		)
		if err != nil {
			return err
		}
		return tx.Outbox().Enqueue(ctx, env)
	})
	return result, err
}

func (uc *OAuthUseCase) establish(ctx context.Context, account *domain.Account, reqCtx RequestContext) (AuthResult, error) {
	var result AuthResult
	err := uc.Store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		res, err := establishSession(ctx, uc.Deps, tx, account, reqCtx, nil)
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	return result, err
}

// uniqueUsername derives a valid, available username from the provider handle (or the
// email local-part), appending a short disambiguator on collision.
func (uc *OAuthUseCase) uniqueUsername(ctx context.Context, candidate, email string) (domain.Username, error) {
	base := strings.TrimSpace(candidate)
	if base == "" {
		base, _, _ = strings.Cut(email, "@")
	}
	base = sanitizeHandle(base)
	if base == "" {
		base = "user"
	}

	for attempt := 0; attempt < 5; attempt++ {
		raw := base
		if attempt > 0 {
			raw = fmt.Sprintf("%s-%s", base, uc.IDs.New().String()[:6])
		}
		username, err := domain.NewUsername(raw)
		if err != nil {
			base = "user"
			continue
		}
		taken, err := uc.Store.Accounts().ExistsByUsername(ctx, username)
		if err != nil {
			return "", err
		}
		if !taken {
			return username, nil
		}
	}
	return "", domain.ErrUsernameTaken
}

func sanitizeHandle(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ' || r == '.':
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-_")
	if len(out) > 32 {
		out = out[:32]
	}
	return out
}
