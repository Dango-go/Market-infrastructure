package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Profile is the customer profile aggregate owned by the user service.
type Profile struct {
	AccountID      uuid.UUID
	Email          string
	Username       string
	DisplayName    string
	Bio            string
	Phone          string
	AvatarURL      string
	Locale         string
	Timezone       string
	MarketingOptIn bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

func NewProfile(accountID uuid.UUID, email, username string, now time.Time) *Profile {
	return &Profile{
		AccountID:      accountID,
		Email:          strings.TrimSpace(strings.ToLower(email)),
		Username:       strings.TrimSpace(username),
		DisplayName:    strings.TrimSpace(username),
		Locale:         "en",
		Timezone:       "UTC",
		CreatedAt:      now,
		UpdatedAt:      now,
		MarketingOptIn: false,
	}
}

func (p *Profile) IsDeleted() bool { return p.DeletedAt != nil }

func (p *Profile) Update(displayName, bio, phone, avatarURL, locale, timezone *string, marketingOptIn *bool, now time.Time) {
	if displayName != nil {
		p.DisplayName = strings.TrimSpace(*displayName)
	}
	if bio != nil {
		p.Bio = strings.TrimSpace(*bio)
	}
	if phone != nil {
		p.Phone = strings.TrimSpace(*phone)
	}
	if avatarURL != nil {
		p.AvatarURL = strings.TrimSpace(*avatarURL)
	}
	if locale != nil {
		p.Locale = strings.TrimSpace(*locale)
	}
	if timezone != nil {
		p.Timezone = strings.TrimSpace(*timezone)
	}
	if marketingOptIn != nil {
		p.MarketingOptIn = *marketingOptIn
	}
	p.UpdatedAt = now
}

func (p *Profile) GetAccountID() uuid.UUID      { return p.AccountID }
func (p *Profile) GetEmail() string             { return p.Email }
func (p *Profile) GetUsername() string          { return p.Username }
func (p *Profile) GetDisplayName() string       { return p.DisplayName }
func (p *Profile) GetBio() string               { return p.Bio }
func (p *Profile) GetPhone() string             { return p.Phone }
func (p *Profile) GetAvatarURL() string         { return p.AvatarURL }
func (p *Profile) GetLocale() string            { return p.Locale }
func (p *Profile) GetTimezone() string          { return p.Timezone }
func (p *Profile) GetMarketingOptIn() bool      { return p.MarketingOptIn }
func (p *Profile) GetCreatedAt() time.Time      { return p.CreatedAt }
func (p *Profile) GetUpdatedAt() time.Time      { return p.UpdatedAt }
