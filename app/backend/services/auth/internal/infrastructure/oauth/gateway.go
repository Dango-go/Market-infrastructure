// Package oauth implements the domain OAuthProviderGateway over golang.org/x/oauth2 for
// GitHub, Google and GitLab, normalizing each provider's profile into OAuthUserInfo.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// ProviderCredentials are the per-provider OAuth client settings loaded from config.
type ProviderCredentials struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Config holds credentials for every supported provider. Providers with empty client ids
// are treated as not configured and rejected at runtime.
type Config struct {
	GitHub ProviderCredentials
	Google ProviderCredentials
	GitLab ProviderCredentials
}

type providerSpec struct {
	endpoint    oauth2.Endpoint
	scopes      []string
	userInfoURL string
	fetch       func(ctx context.Context, client *http.Client) (domain.OAuthUserInfo, error)
}

// Gateway is the concrete OAuthProviderGateway.
type Gateway struct {
	configs map[domain.OAuthProvider]*oauth2.Config
	specs   map[domain.OAuthProvider]providerSpec
	client  *http.Client
}

var _ domain.OAuthProviderGateway = (*Gateway)(nil)

// NewGateway wires the provider configs and HTTP client.
func NewGateway(cfg Config) *Gateway {
	g := &Gateway{
		configs: map[domain.OAuthProvider]*oauth2.Config{},
		specs:   map[domain.OAuthProvider]providerSpec{},
		client:  &http.Client{Timeout: 10 * time.Second},
	}

	g.register(domain.ProviderGitHub, cfg.GitHub, providerSpec{
		endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
		scopes: []string{"read:user", "user:email"},
		fetch:  fetchGitHub,
	})
	g.register(domain.ProviderGoogle, cfg.Google, providerSpec{
		endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
		scopes: []string{"openid", "email", "profile"},
		fetch:  fetchGoogle,
	})
	g.register(domain.ProviderGitLab, cfg.GitLab, providerSpec{
		endpoint: oauth2.Endpoint{
			AuthURL:  "https://gitlab.com/oauth/authorize",
			TokenURL: "https://gitlab.com/oauth/token",
		},
		scopes: []string{"read_user"},
		fetch:  fetchGitLab,
	})
	return g
}

func (g *Gateway) register(provider domain.OAuthProvider, creds ProviderCredentials, spec providerSpec) {
	if creds.ClientID == "" {
		return
	}
	g.configs[provider] = &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		RedirectURL:  creds.RedirectURL,
		Endpoint:     spec.endpoint,
		Scopes:       spec.scopes,
	}
	g.specs[provider] = spec
}

// AuthCodeURL builds the consent URL for the given state.
func (g *Gateway) AuthCodeURL(provider domain.OAuthProvider, state string) (string, error) {
	cfg, ok := g.configs[provider]
	if !ok {
		return "", domain.ErrUnsupportedProvider
	}
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// Exchange swaps the code for a token and fetches the normalized profile.
func (g *Gateway) Exchange(ctx context.Context, provider domain.OAuthProvider, code string) (domain.OAuthUserInfo, error) {
	cfg, ok := g.configs[provider]
	if !ok {
		return domain.OAuthUserInfo{}, domain.ErrUnsupportedProvider
	}
	spec := g.specs[provider]

	ctx = context.WithValue(ctx, oauth2.HTTPClient, g.client)
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return domain.OAuthUserInfo{}, fmt.Errorf("oauth exchange: %w", err)
	}
	httpClient := cfg.Client(ctx, tok)

	info, err := spec.fetch(ctx, httpClient)
	if err != nil {
		return domain.OAuthUserInfo{}, err
	}
	info.Provider = provider
	return info, nil
}

func getJSON(ctx context.Context, client *http.Client, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("provider userinfo returned status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func fetchGitHub(ctx context.Context, client *http.Client) (domain.OAuthUserInfo, error) {
	var profile struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
	}
	if err := getJSON(ctx, client, "https://api.github.com/user", &profile); err != nil {
		return domain.OAuthUserInfo{}, err
	}

	email, verified := profile.Email, profile.Email != ""
	if email == "" {
		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := getJSON(ctx, client, "https://api.github.com/user/emails", &emails); err == nil {
			for _, e := range emails {
				if e.Primary {
					email, verified = e.Email, e.Verified
					break
				}
			}
		}
	}
	return domain.OAuthUserInfo{
		ProviderUserID: strconv.FormatInt(profile.ID, 10),
		Email:          email,
		Username:       profile.Login,
		EmailVerified:  verified,
	}, nil
}

func fetchGoogle(ctx context.Context, client *http.Client) (domain.OAuthUserInfo, error) {
	var profile struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := getJSON(ctx, client, "https://openidconnect.googleapis.com/v1/userinfo", &profile); err != nil {
		return domain.OAuthUserInfo{}, err
	}
	return domain.OAuthUserInfo{
		ProviderUserID: profile.Sub,
		Email:          profile.Email,
		Username:       profile.Name,
		EmailVerified:  profile.EmailVerified,
	}, nil
}

func fetchGitLab(ctx context.Context, client *http.Client) (domain.OAuthUserInfo, error) {
	var profile struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := getJSON(ctx, client, "https://gitlab.com/api/v4/user", &profile); err != nil {
		return domain.OAuthUserInfo{}, err
	}
	return domain.OAuthUserInfo{
		ProviderUserID: strconv.FormatInt(profile.ID, 10),
		Email:          profile.Email,
		Username:       profile.Username,
		EmailVerified:  profile.Email != "",
	}, nil
}
