package config

import (
	"crypto/rsa"
	"fmt"
	"net/url"
	"os"

	"github.com/golang-jwt/jwt/v5"

	baseconfig "github.com/embedded-market/backend/pkg/config"
)

type Config struct {
	baseconfig.Base
	JWT      JWTConfig
	Upstream UpstreamConfig
}

type JWTConfig struct {
	PublicKeyPEM  string `env:"JWT_PUBLIC_KEY_PEM"`
	PublicKeyFile string `env:"JWT_PUBLIC_KEY_FILE"`
	Issuer        string `env:"JWT_ISSUER" envDefault:"embedded-market-auth"`
	Audience      string `env:"JWT_AUDIENCE" envDefault:"embedded-market"`

	publicKey *rsa.PublicKey
}

func (j JWTConfig) PublicKey() *rsa.PublicKey { return j.publicKey }

type UpstreamConfig struct {
	AuthURL           string `env:"AUTH_SERVICE_URL,required"`
	UserURL           string `env:"USER_SERVICE_URL,required"`
	CatalogURL        string `env:"CATALOG_SERVICE_URL,required"`
	InventoryURL      string `env:"INVENTORY_SERVICE_URL,required"`
	PricingURL        string `env:"PRICING_SERVICE_URL,required"`
	CartURL           string `env:"CART_SERVICE_URL,required"`
	OrderURL          string `env:"ORDER_SERVICE_URL,required"`
	ShippingURL       string `env:"SHIPPING_SERVICE_URL,required"`
	PaymentURL        string `env:"PAYMENT_SERVICE_URL,required"`
	NotificationURL   string `env:"NOTIFICATION_SERVICE_URL,required"`
	WishlistURL       string `env:"WISHLIST_SERVICE_URL,required"`
	SearchURL         string `env:"SEARCH_SERVICE_URL,required"`
	ReviewURL         string `env:"REVIEW_SERVICE_URL,required"`
	RecommendationURL string `env:"RECOMMENDATION_SERVICE_URL,required"`
	AnalyticsURL      string `env:"ANALYTICS_SERVICE_URL,required"`

	Auth           *url.URL
	User           *url.URL
	Catalog        *url.URL
	Inventory      *url.URL
	Pricing        *url.URL
	Cart           *url.URL
	Order          *url.URL
	Shipping       *url.URL
	Payment        *url.URL
	Notification   *url.URL
	Wishlist       *url.URL
	Search         *url.URL
	Review         *url.URL
	Recommendation *url.URL
	Analytics      *url.URL
}

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.ServiceName = "api-gateway"
	if err := baseconfig.Load(cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "api-gateway"
	}
	pemBytes, err := resolvePublicKeyPEM(cfg.JWT)
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parse JWT public key: %w", err)
	}
	cfg.JWT.publicKey = key
	if err := cfg.Upstream.parse(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func resolvePublicKeyPEM(j JWTConfig) ([]byte, error) {
	if j.PublicKeyPEM != "" {
		return []byte(j.PublicKeyPEM), nil
	}
	if j.PublicKeyFile != "" {
		data, err := os.ReadFile(j.PublicKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read JWT public key file: %w", err)
		}
		return data, nil
	}
	return nil, fmt.Errorf("a JWT verification key is required: set JWT_PUBLIC_KEY_PEM or JWT_PUBLIC_KEY_FILE")
}

func (u *UpstreamConfig) parse() error {
	var err error
	if u.Auth, err = parseURL(u.AuthURL, "AUTH_SERVICE_URL"); err != nil { return err }
	if u.User, err = parseURL(u.UserURL, "USER_SERVICE_URL"); err != nil { return err }
	if u.Catalog, err = parseURL(u.CatalogURL, "CATALOG_SERVICE_URL"); err != nil { return err }
	if u.Inventory, err = parseURL(u.InventoryURL, "INVENTORY_SERVICE_URL"); err != nil { return err }
	if u.Pricing, err = parseURL(u.PricingURL, "PRICING_SERVICE_URL"); err != nil { return err }
	if u.Cart, err = parseURL(u.CartURL, "CART_SERVICE_URL"); err != nil { return err }
	if u.Order, err = parseURL(u.OrderURL, "ORDER_SERVICE_URL"); err != nil { return err }
	if u.Shipping, err = parseURL(u.ShippingURL, "SHIPPING_SERVICE_URL"); err != nil { return err }
	if u.Payment, err = parseURL(u.PaymentURL, "PAYMENT_SERVICE_URL"); err != nil { return err }
	if u.Notification, err = parseURL(u.NotificationURL, "NOTIFICATION_SERVICE_URL"); err != nil { return err }
	if u.Wishlist, err = parseURL(u.WishlistURL, "WISHLIST_SERVICE_URL"); err != nil { return err }
	if u.Search, err = parseURL(u.SearchURL, "SEARCH_SERVICE_URL"); err != nil { return err }
	if u.Review, err = parseURL(u.ReviewURL, "REVIEW_SERVICE_URL"); err != nil { return err }
	if u.Recommendation, err = parseURL(u.RecommendationURL, "RECOMMENDATION_SERVICE_URL"); err != nil { return err }
	if u.Analytics, err = parseURL(u.AnalyticsURL, "ANALYTICS_SERVICE_URL"); err != nil { return err }
	return nil
}

func parseURL(raw, name string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid %s: %q", name, raw)
	}
	return parsed, nil
}
