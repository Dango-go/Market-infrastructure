# Auth Service

Bounded context: **Identity & Access**. Owns credentials, sessions and external-identity
links. The only service permitted to mint access tokens; it holds the RS256 private key
and publishes the public key as a JWKS so every other service verifies tokens locally.

## Entities & value objects

| Type | Kind | Notes |
|------|------|-------|
| `Account` | aggregate root | id, email, username, password_hash (nullable for OAuth-only), status, email_verified, timestamps, soft delete |
| `Session` | entity | refresh-token grant; stores only the SHA-256 of the opaque refresh token, `rotated_from` for rotation lineage |
| `OAuthIdentity` | entity | links `(provider, provider_user_id)` → account |
| `Email`, `Username`, `PasswordHash` | value objects | validated & normalized on construction |
| `AccountStatus` | value object | pending_verification → active / suspended / deactivated |

Password policy (length + character-class diversity) and `AccountStatus.CanAuthenticate`
live in the domain, so every entry point enforces identical rules.

## Database schema (PostgreSQL)

- `accounts` — UUID PK, partial unique indexes on `lower(email)` / `lower(username)`
  scoped to non-deleted rows (a handle frees up after soft delete), status CHECK.
- `sessions` — UUID PK, FK→accounts (cascade), unique `refresh_token_hash`, partial index
  on active sessions per account, self-FK `rotated_from`.
- `oauth_identities` — UUID PK, FK→accounts, unique `(provider, provider_user_id)`.
- `outbox` — transactional outbox (id, type, payload jsonb, occurred_at, published_at).
- `schema_migrations` — applied-migration ledger (embedded migrations run at startup).

## Use cases (application layer)

`Register`, `Login`, `Refresh` (rotation + reuse detection), `Logout`, `Sessions`
(list / revoke one / revoke all), `Account` (current identity), `OAuth` (begin / complete
with find-or-provision-or-link). Multi-write operations run inside `Store.WithinTx`, which
also enqueues the outbox event in the same transaction.

## Repository interfaces (ports, owned by domain)

`AccountRepository`, `SessionRepository`, `OAuthRepository`, `OutboxRepository`, plus
`Store` (transactional composition). Infrastructure ports: `PasswordHasher` (argon2id),
`TokenService` (RS256 + opaque refresh), `OAuthProviderGateway`, `Clock`, `IDGenerator`.

## REST endpoints (`/api/v1/auth`)

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/register` | – | create account, return tokens |
| POST | `/login` | – | password login |
| POST | `/refresh` | – (token) | rotate refresh token |
| POST | `/logout` | – (token) | revoke session |
| GET | `/me` | Bearer | current account |
| GET | `/sessions` | Bearer | list active sessions (paginated) |
| DELETE | `/sessions/:id` | Bearer | revoke one session |
| DELETE | `/sessions` | Bearer | revoke all sessions |
| GET | `/oauth/:provider` | – | begin OAuth (github/google/gitlab) |
| GET | `/oauth/:provider/callback` | – | complete OAuth |
| GET | `/.well-known/jwks.json` | – | public verification keys |
| GET | `/healthz`, `/readyz` | – | liveness / readiness |

Pagination (`page`/`page_size`), validation (422 + field details), and the stable error
envelope come from shared `pkg`.

## Kafka events

Publishes **`user.registered`** (via the outbox relay) when an account is created by
registration or first OAuth login. Consumed downstream by the User service (provision
profile) and Notification service (welcome). The auth service consumes no events.

## Interactions with other services

- **All services & gateway** verify access tokens locally against the JWKS — no synchronous
  call to auth on the hot path.
- **User service** reacts to `user.registered` to create the public profile.
- **Notification service** reacts to `user.registered` for onboarding.

## Folder structure

```
services/auth/
├── cmd/auth/main.go                 # composition root
├── internal/
│   ├── config/                      # env-driven configuration + key loading
│   ├── domain/                      # entities, VOs, errors, ports
│   ├── application/                 # use cases + DTOs
│   ├── infrastructure/
│   │   ├── crypto/                  # argon2id hasher
│   │   ├── token/                   # RS256 issuer + JWKS
│   │   ├── oauth/                   # provider gateway (x/oauth2)
│   │   ├── events/                  # outbox relay → Kafka
│   │   └── system/                  # clock, uuidv7
│   ├── repository/postgres/         # pgx implementations of the ports
│   └── transport/http/              # gin handlers, DTOs, router
└── db/{migrations,queries,sqlc.yaml}
```
