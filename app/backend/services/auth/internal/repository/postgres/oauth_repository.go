package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

type oauthRepository struct {
	db pgxConn
}

const oauthColumns = `id, account_id, provider, provider_user_id, email, created_at, updated_at`

func (r *oauthRepository) Create(ctx context.Context, o *domain.OAuthIdentity) error {
	const q = `
		INSERT INTO oauth_identities (id, account_id, provider, provider_user_id, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, q,
		o.ID, o.AccountID, string(o.Provider), o.ProviderUserID, o.Email, o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		if _, ok := isUniqueViolation(err); ok {
			return domain.ErrOAuthAlreadyLinked
		}
		return fmt.Errorf("insert oauth identity: %w", err)
	}
	return nil
}

func (r *oauthRepository) GetByProviderUserID(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.OAuthIdentity, error) {
	q := `SELECT ` + oauthColumns + ` FROM oauth_identities WHERE provider = $1 AND provider_user_id = $2`
	var (
		o        domain.OAuthIdentity
		provName string
	)
	err := r.db.QueryRow(ctx, q, string(provider), providerUserID).Scan(
		&o.ID, &o.AccountID, &provName, &o.ProviderUserID, &o.Email, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrOAuthIdentityNotFound
		}
		return nil, fmt.Errorf("scan oauth identity: %w", err)
	}
	o.Provider = domain.OAuthProvider(provName)
	return &o, nil
}
