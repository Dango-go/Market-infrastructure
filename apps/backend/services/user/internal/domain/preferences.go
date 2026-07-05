package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Preferences stores commerce personalization settings.
type Preferences struct {
	AccountID         uuid.UUID
	Currency          string
	Language          string
	EmailNotifications bool
	SMSNotifications  bool
	PushNotifications bool
	MarketingOptIn    bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewPreferences(accountID uuid.UUID, now time.Time) *Preferences {
	return &Preferences{
		AccountID:         accountID,
		Currency:          "USD",
		Language:          "en",
		EmailNotifications: true,
		SMSNotifications:  false,
		PushNotifications: false,
		MarketingOptIn:    false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (p *Preferences) Update(currency, language *string, emailNotifications, smsNotifications, pushNotifications, marketingOptIn *bool, now time.Time) {
	if currency != nil {
		p.Currency = strings.ToUpper(strings.TrimSpace(*currency))
	}
	if language != nil {
		p.Language = strings.TrimSpace(*language)
	}
	if emailNotifications != nil {
		p.EmailNotifications = *emailNotifications
	}
	if smsNotifications != nil {
		p.SMSNotifications = *smsNotifications
	}
	if pushNotifications != nil {
		p.PushNotifications = *pushNotifications
	}
	if marketingOptIn != nil {
		p.MarketingOptIn = *marketingOptIn
	}
	p.UpdatedAt = now
}

func (p *Preferences) GetAccountID() uuid.UUID      { return p.AccountID }
func (p *Preferences) GetCurrency() string          { return p.Currency }
func (p *Preferences) GetLanguage() string          { return p.Language }
func (p *Preferences) GetEmailNotifications() bool  { return p.EmailNotifications }
func (p *Preferences) GetSMSNotifications() bool    { return p.SMSNotifications }
func (p *Preferences) GetPushNotifications() bool   { return p.PushNotifications }
func (p *Preferences) GetMarketingOptIn() bool      { return p.MarketingOptIn }
func (p *Preferences) GetCreatedAt() time.Time      { return p.CreatedAt }
func (p *Preferences) GetUpdatedAt() time.Time      { return p.UpdatedAt }
