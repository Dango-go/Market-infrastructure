package domain

import "unicode"

// Password complexity policy. Centralized here so every entry point (registration,
// future password reset/change) enforces identical rules.
const (
	minPasswordLength = 10
	maxPasswordLength = 128
)

// ValidatePassword enforces the platform password policy, returning ErrWeakPassword when
// the plaintext is too short, too long, or lacks character-class diversity.
func ValidatePassword(plaintext string) error {
	if len(plaintext) < minPasswordLength || len(plaintext) > maxPasswordLength {
		return ErrWeakPassword
	}
	var hasLetter, hasDigit, hasOther bool
	for _, r := range plaintext {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasOther = true
		}
	}
	// Require at least two distinct character classes.
	classes := 0
	for _, present := range []bool{hasLetter, hasDigit, hasOther} {
		if present {
			classes++
		}
	}
	if classes < 2 {
		return ErrWeakPassword
	}
	return nil
}
