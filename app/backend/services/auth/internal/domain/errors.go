package domain

import "github.com/embedded-market/backend/pkg/apperr"

// Domain errors are expressed in the shared apperr taxonomy so the transport layer maps
// them to HTTP without the domain knowing anything about HTTP. They are package-level
// values because they carry no per-instance state.
var (
	ErrInvalidEmail    = apperr.Invalid("invalid_email", "the email address is invalid")
	ErrInvalidUsername = apperr.Invalid("invalid_username", "the username must be 3–32 characters of letters, digits, '-' or '_'")
	ErrWeakPassword    = apperr.Invalid("weak_password", "the password does not meet complexity requirements")

	ErrEmailTaken    = apperr.Conflict("email_taken", "an account with this email already exists")
	ErrUsernameTaken = apperr.Conflict("username_taken", "this username is already taken")

	ErrAccountNotFound    = apperr.NotFound("account_not_found", "account not found")
	ErrInvalidCredentials = apperr.Unauthorized("invalid_credentials", "email or password is incorrect")

	ErrSessionNotFound = apperr.NotFound("session_not_found", "session not found")
	ErrSessionRevoked  = apperr.Unauthorized("session_revoked", "the session has been revoked")
	ErrSessionExpired  = apperr.Unauthorized("session_expired", "the session has expired")
	ErrInvalidRefresh  = apperr.Unauthorized("invalid_refresh_token", "the refresh token is invalid")

	ErrOAuthIdentityNotFound = apperr.NotFound("oauth_identity_not_found", "no linked account for this provider identity")
	ErrOAuthAlreadyLinked    = apperr.Conflict("oauth_already_linked", "this provider identity is already linked to an account")
	ErrUnsupportedProvider   = apperr.Invalid("unsupported_provider", "the OAuth provider is not supported")
	ErrOAuthStateMismatch  = apperr.Unauthorized("oauth_state_mismatch", "the OAuth state is invalid or expired")
	ErrOAuthExchange       = apperr.Unavailable("oauth_exchange_failed", "failed to complete the OAuth exchange")
)
