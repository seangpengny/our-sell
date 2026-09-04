# Our Sell backend

A small, production-oriented authentication backend written in Go with Fiber. The initial feature set is deliberately narrow: users, authentication, opaque rotating refresh sessions, email verification, password recovery, rate limiting, and operational health endpoints.

## Stack

- Go 1.27.1
- Fiber v3.5.0
- PostgreSQL 18 with pgx v5.10.0 and sqlc v1.31.1
- Redis 8 with go-redis v9
- Argon2id password hashes
- HS256 JWT access tokens through golang-jwt/jwt v5
- goose migrations
- slog JSON logging
- Docker Compose

The pinned versions were checked against upstream release pages while this project was created. Keep dependencies current with normal review and vulnerability scanning.

## Architecture

The request path stays explicit:

~~~text
Fiber route -> middleware -> handler -> service -> repository -> PostgreSQL
                                      |             |
                                      |             +-> sqlc-generated pgx queries
                                      +-> Redis / Mailer
~~~

Business logic does not receive a Fiber context. The auth service can therefore be reused by another transport or a background worker. Realtime has a small concurrency-safe hub in internal/realtime; a future WebSocket adapter can call the same services without moving business logic into the transport.

Feature code lives with the feature. Shared packages are limited to infrastructure and platform concerns:

~~~text
cmd/api                  process entry point and graceful shutdown
internal/app             dependency wiring, routes, error handling
internal/auth            authentication, sessions, tokens, auth HTTP layer
internal/user            user model, repository, and user service boundary
internal/middleware      request ID, logging, recovery, CORS, security, rate limiting
internal/database        PostgreSQL pool and Redis client setup
internal/platform        public errors, response envelope, validation, auth context
internal/realtime        transport-neutral hub and future WebSocket boundary
db/migrations             goose SQL migrations
db/queries                SQL source for sqlc
db/sqlc                   generated code; never edit by hand
api/openapi.yaml          versioned API contract
~~~

## Quick start

Requirements: Go 1.27+, Docker Desktop, and Docker Compose.

~~~bash
cp .env.example .env
docker compose up -d postgres redis
make migrate-up
go run ./cmd/api
~~~

The API is then available at http://localhost:8080.

To run the complete local stack in containers:

~~~bash
make start
~~~

Useful Make targets:

~~~text
make start     # build and start PostgreSQL, Redis, and the API
make stop      # stop and remove the containers; keep database volumes
make status    # show container status
make logs      # follow API logs
~~~

The Compose file intentionally does not run migrations automatically. Run migrations as an explicit deployment step so schema changes are observable and reversible.

## Environment

.env.example is safe development configuration only. Important settings are:

- DATABASE_URL, REDIS_URL: required connection URLs.
- JWT_SECRET: required 32+ byte signing secret; use a secret manager in production.
- JWT_ISSUER, JWT_AUDIENCE: pinned JWT validation context.
- ACCESS_TOKEN_TTL: validated between 10 and 30 minutes; default 15 minutes.
- REFRESH_TOKEN_TTL: default 30 days; refresh tokens are not JWTs.
- CORS_ALLOWED_ORIGINS: comma-separated explicit origins; * is rejected.
- TRUSTED_PROXY_CIDRS and TRUST_FORWARDED_HEADERS: only enable forwarded-IP parsing for known proxies.
- COOKIE_SECURE, COOKIE_SAMESITE, COOKIE_DOMAIN: refresh-cookie policy. Production requires COOKIE_SECURE=true.
- DB_MAX_CONNS, DB_MIN_CONNS, DB_MAX_CONN_LIFETIME, DB_MAX_CONN_IDLE_TIME, DB_HEALTH_CHECK_PERIOD: pool controls.
- REQUEST_BODY_LIMIT: global request body cap, default 1 MiB.
- SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD: reserved for a real mailer adapter; the default build uses a token-blind development mailer.

Configuration is validated before opening external connections. No secrets are committed.

## Database and sqlc

Migrations are in db/migrations and use goose annotations.

~~~bash
make migrate-up
make migrate-down
make sqlc
~~~

PostgreSQL is the source of truth. Redis is used for distributed rate limiting and is not required to store users or authentication sessions.

The schema uses UUID primary keys generated in Go with UUIDv7 where supported by the UUID library. Refresh token hashes and token history are stored as BYTEA; raw tokens never reach the database.

## Development commands

~~~bash
make dev          # start local PostgreSQL/Redis and run the API
make build        # compile bin/our-sell-api
make test         # unit tests
make test-race    # race detector
make test-integration # Testcontainers PostgreSQL/Redis smoke test
make lint         # golangci-lint
make fmt          # gofmt
make sqlc         # regenerate db/sqlc
make docker-up
make docker-down
~~~

The normal verification set is:

~~~bash
go mod tidy
go fmt ./...
go vet ./...
golangci-lint run
go test ./...
go test -race ./...
go build -o bin/our-sell-api ./cmd/api
~~~

## Authentication design

### Registration and login

Registration normalizes email, bounds input, validates password length, stores only an Argon2id encoded hash, and creates a single-use email-verification token. Login returns a short-lived JWT access token and sets an opaque refresh token in an HttpOnly cookie. Unknown-account and wrong-password cases use the same public error.

Argon2id parameters are encoded with each hash so the work factor can be upgraded later. password.VerifyPassword uses constant-time comparison and rejects malformed encodings.

### Access tokens

The JWT contains only sub, session_id, role, iat, exp, iss, and aud. Validation explicitly requires HS256, the configured issuer and audience, an expiration, a valid issued-at value, and UUID-shaped subject/session IDs. The signing secret is never placed in claims.

### Refresh sessions

Refresh tokens are 32 cryptographically random bytes rendered as hex. Only SHA-256 hashes are stored. A refresh rotates the token inside a PostgreSQL transaction:

~~~text
presented token -> lock token/session -> mark old token used
                -> update current session hash -> insert new token history
~~~

Reusing a previously used token revokes every session for that user. This makes refresh-token replay visible to the security boundary and limits damage from token theft. Logout revokes the current server-side session and clears the cookie; logout-all revokes every active session.

GET /api/v1/me/sessions returns only safe metadata, and DELETE /api/v1/me/sessions/:sessionID includes the authenticated user ID in its SQL predicate to prevent IDOR.

### Password recovery and verification

Forgot-password always returns the same message for valid and unknown addresses. Reset and verification tokens are random, hashed, expiring, and single-use. Password reset, token consumption, and session revocation are one database transaction. Password change verifies the current hash and revokes other sessions while preserving the current session.

The default mailer is token-blind and only represents a development integration boundary. Replace internal/email.LogMailer with an SMTP/provider adapter implementing email.Mailer; do not log tokens in the adapter.

### Cookies and CSRF

The refresh cookie is HttpOnly, path-scoped to /api/v1/auth, and uses configurable SameSite behavior. The default development setting is Lax; production should use HTTPS and choose Strict or Lax based on the frontend deployment. SameSite is a defense-in-depth CSRF control, not a complete CSRF architecture. If the frontend requires cross-site credentialed requests, use an explicit CSRF token/origin-check design before changing to SameSite=None; Secure is mandatory for None and production.

## HTTP API

The complete contract is in api/openapi.yaml.

Start the API, then open [http://localhost:8080/docs](http://localhost:8080/docs) for the interactive Swagger UI. The same embedded contract is available at [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml).

Use `Try it out` to call endpoints. For protected endpoints, call login first, copy the returned access token, click `Authorize`, and enter `Bearer <access-token>`. Refresh accepts the refresh cookie automatically in a browser or a `refresh_token` JSON body for non-browser clients.

~~~text
GET    /health
GET    /ready
GET    /docs
GET    /openapi.yaml

POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
POST   /api/v1/auth/logout-all
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
POST   /api/v1/auth/verify-email
POST   /api/v1/auth/resend-verification
POST   /api/v1/auth/change-password
GET    /api/v1/auth/me
GET    /api/v1/me/sessions
DELETE /api/v1/me/sessions/:sessionID
~~~

Responses use either { "data": ... } or { "message": ... }. Errors use { "error": { "code": ..., "message": ..., "fields": ... } }. Internal SQL, stack, path, hash, cookie, and token details are never sent to clients.

## Security and operations

- Request IDs are accepted only when bounded to a safe character set; otherwise a random ID is generated and echoed in X-Request-ID.
- Structured logs include request ID, route, status, duration, and user ID when available. Passwords, bearer tokens, cookies, and reset/verification tokens are never logged.
- CORS allows only configured origins and credentials are never combined with wildcard origins.
- Rate limits are Redis-backed and distributed. Auth endpoints have stricter per-IP and per-email-hash limits; a production Redis outage fails closed by default.
- Security headers include X-Content-Type-Options, Referrer-Policy, Permissions-Policy, and Cache-Control: no-store. HSTS is emitted only in production. A CSP is intentionally not emitted because this service serves JSON, not HTML.
- Fiber's trusted-proxy configuration is disabled by default. Forwarded headers are considered only when explicitly enabled with proxy CIDRs.
- Docker builds a static binary, copies no source or development secrets into the runtime image, and runs as a non-root user.
- SIGINT/SIGTERM stops accepting traffic through Fiber's graceful shutdown, then closes the PostgreSQL and Redis clients.

Do not treat this starter as a claim of perfect security. Add dependency scanning, secret rotation, TLS termination, audit retention, alerting, backups, and a threat-model review before production deployment.

## Realtime roadmap

internal/realtime.Hub is an in-process, transport-neutral abstraction. When realtime becomes product-critical, add a Fiber WebSocket adapter that authenticates through the same token/session rules and use Redis Pub/Sub or NATS for cross-instance fan-out. Do not make Redis authoritative for user or session state, and do not add NATS until there is a real event-driven use case.
