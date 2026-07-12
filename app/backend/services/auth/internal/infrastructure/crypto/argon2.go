// Package crypto provides the argon2id implementation of the domain PasswordHasher port.
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// Argon2Params controls the argon2id cost. Defaults follow current OWASP guidance.
type Argon2Params struct {
	Memory      uint32 // KiB
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params returns sensible production defaults (64 MiB, t=3, p=2).
func DefaultArgon2Params() Argon2Params {
	return Argon2Params{Memory: 64 * 1024, Iterations: 3, Parallelism: 2, SaltLength: 16, KeyLength: 32}
}

// Argon2Hasher hashes and verifies passwords with argon2id, encoding the parameters and
// salt in the standard PHC string format so hashes remain self-describing and upgradable.
type Argon2Hasher struct {
	p Argon2Params
}

// NewArgon2Hasher builds a hasher with the given parameters.
func NewArgon2Hasher(p Argon2Params) *Argon2Hasher { return &Argon2Hasher{p: p} }

var _ domain.PasswordHasher = (*Argon2Hasher)(nil)

// Hash derives a PHC-encoded argon2id hash of the plaintext.
func (h *Argon2Hasher) Hash(plaintext string) (domain.PasswordHash, error) {
	salt := make([]byte, h.p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	key := argon2.IDKey([]byte(plaintext), salt, h.p.Iterations, h.p.Memory, h.p.Parallelism, h.p.KeyLength)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.p.Memory, h.p.Iterations, h.p.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	return domain.PasswordHash(encoded), nil
}

// Verify compares a plaintext against a PHC-encoded hash in constant time.
func (h *Argon2Hasher) Verify(hash domain.PasswordHash, plaintext string) (bool, error) {
	params, salt, key, err := decode(hash.String())
	if err != nil {
		return false, err
	}
	candidate := argon2.IDKey([]byte(plaintext), salt, params.Iterations, params.Memory, params.Parallelism, uint32(len(key)))
	return subtle.ConstantTimeCompare(key, candidate) == 1, nil
}

func decode(encoded string) (Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Argon2Params{}, nil, nil, errors.New("invalid argon2 hash format")
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("parse argon2 version: %w", err)
	}
	var p Argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism); err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("parse argon2 params: %w", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("decode salt: %w", err)
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("decode key: %w", err)
	}
	return p, salt, key, nil
}
