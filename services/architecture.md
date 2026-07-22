# IAM Service Architecture

This service is organized as a small hexagonal application:

- `domain` defines the core business concepts and errors.
- `usecase` contains application logic and orchestration.
- `port` defines the interfaces the application depends on.
- `adapter` contains concrete implementations for infrastructure concerns.
- `transport/http` contains the HTTP adapter built on Echo.
- `bootstrap` wires the application together.
- `cmd/iam/main.go` is the thin launcher that starts the bootstrapped app.

## Folder Structure

```text
services
├── go.mod
├── cmd
│   └── iam
│       └── main.go
├── internal
│   └── iam
│       ├── bootstrap
│       │   ├── bootstrap.go
│       │   └── bootstrap_test.go
│       ├── adapter
│       │   ├── repository
│       │   │   ├── memory
│       │   │   │   ├── user_repository.go
│       │   │   │   └── user_repository_test.go
│       │   │   └── postgres
│       │   │       ├── db.go
│       │   │       ├── seed.go
│       │   │       ├── user_repository.go
│       │   │       ├── permission_repository.go
│       │   │       ├── refresh_token_repository.go
│       │   │       ├── login_audit_repository.go
│       │   │       └── migrations/
│       │   │           ├── 000001_init.up.sql
│       │   │           └── 000001_init.down.sql
│       │   └── security
│       │       ├── bcrypt.go
│       │       └── jwt.go
│       ├── domain
│       │   ├── errors.go
│       │   ├── role.go
│       │   └── user.go
│       ├── port
│       │   ├── repository/
│       │   │   ├── user_repository.go
│       │   │   ├── role_repository.go
│       │   │   ├── permission_repository.go
│       │   │   ├── refresh_token_repository.go
│       │   │   └── login_audit_repository.go
│       │   ├── event_publisher.go
│       │   ├── jwks_provider.go
│       │   ├── password_hasher.go
│       │   └── token_signer.go
│       ├── transport
│       │   └── http
│       │       ├── handler/
│       │       │   ├── handler.go
│       │       │   ├── auth_handler.go
│       │       │   ├── user_handler.go
│       │       │   ├── role_handler.go
│       │       │   ├── permission_handler.go
│       │       │   ├── errors.go
│       │       │   └── types.go
│       │       ├── middleware/
│       │       │   ├── middleware.go
│       │       │   ├── errors.go
│       │       │   └── middleware_test.go
│       │       └── router.go
│       └── usecase
│           ├── auth.go
│           └── auth_test.go
├── .env.example
├── architecture.md
└── Dockerfile.iam
```

## Layer Boundaries

### Domain

The `internal/iam/domain` package is the innermost layer. It contains:

- user, role, refresh token, session, API key, and login audit models
- permissions as typed string constants in `resource.action` format
- default system roles (owner, admin, supervisor, cashier, viewer)
- shared business errors with error codes (e.g. `AUTH001`–`AUTH006`)

Domain code should stay free of transport, persistence, and framework concerns.

### Ports

The `internal/iam/port` package defines the contracts used by the application layer:

- `repository/UserRepository`
- `repository/RoleRepository`
- `repository/PermissionRepository`
- `repository/RefreshTokenRepository`
- `repository/LoginAuditRepository`
- `PasswordHasher`
- `TokenSigner` / `TokenClaims` / `TokenPair`
- `JWKSProvider`
- `EventPublisher`

Repository interfaces are grouped in `port/repository/` to follow Interface Segregation — each interface in its own file.

These interfaces let the usecase layer stay independent from the concrete storage or security implementations.

### Application / Use Cases

The `internal/iam/usecase` package owns the business workflow for authentication:

- register a user (with password policy validation)
- authenticate a user (returns access + refresh token pair)
- refresh token (rotation with old token revocation)
- logout (revoke refresh tokens)
- token introspection
- validate JWT claims against the current user state
- check permissions
- CRUD for users, roles, permissions
- role assignment / revocation per user
- permission assignment / revocation per role
- login audit logging
- event publishing (UserCreated, UserUpdated, UserDeleted, RoleAssigned, RoleRevoked)

`AuthUsecase` depends only on ports and domain types. It does not know whether the database is memory, SQL, or remote, and it does not know whether JWTs are implemented with Echo, `net/http`, or anything else.

### Adapters

The `internal/iam/adapter` package contains infrastructure implementations:

- `adapter/security/argon2.go` implements `PasswordHasher` using Argon2id
- `adapter/security/jwt.go` implements `TokenSigner` using RS256 with 15-minute access tokens and 30-day refresh tokens
- `adapter/repository/memory/` implements all repository interfaces in-memory
- `adapter/repository/postgres/` implements all repository interfaces backed by PostgreSQL with:
  - connection pooling via `pgx/v5`
  - embedded SQL migrations via `golang-migrate`
  - seed data for default permissions and system roles
  - SHA256 token hashing for refresh tokens

These are replaceable details. If the hashing or token strategy changes, the application layer should not need to change.

### HTTP Transport

The `internal/iam/transport/http` package is the inbound HTTP adapter, split into sub-packages:

- `handler/` converts HTTP requests into usecase inputs and responses, one file per domain entity
- `middleware/` handles CORS, authentication, RBAC checks, branch-scoped access checks, and rate limiting
- `router.go` wires routes and middleware using Echo
- `/.well-known/jwks.json` publishes the public key set for external JWT verification

This layer should translate HTTP concerns only. It should not implement business rules beyond request/response shaping.

### Persistence

Persistence adapters live under `internal/iam/adapter/repository`:

- `memory` is the in-memory implementation for development and testing
- `postgres` is the PostgreSQL implementation for production

The bootstrap layer chooses which adapters to wire based on the presence of a `DATABASE_URL` environment variable.

## Request Flow

### Register

1. `POST /api/v1/auth/register` reaches the Echo router.
2. `handler/auth_handler.go` validates the request body shape.
3. `usecase.AuthUsecase.Register` validates the password policy, checks duplicates, hashes the password, persists the user, fetches permissions, and signs an access token.
4. The handler returns the access token and user as JSON.

### Login

1. `POST /api/v1/auth/login` reaches the Echo router.
2. `handler/auth_handler.go` validates the request body shape.
3. `usecase.AuthUsecase.Login` loads the user, verifies the password, checks user status, fetches permissions, signs an access token, creates a refresh token, logs the login audit, and stores the refresh token.
4. The handler returns the access token, refresh token, and user as JSON.

### Refresh

1. `POST /api/v1/auth/refresh` reaches the Echo router.
2. `handler/auth_handler.go` validates the request body shape.
3. `usecase.AuthUsecase.RefreshToken` verifies the refresh token, checks revocation and expiry, revokes the old token, and issues a new token pair.
4. The handler returns the new token pair and user as JSON.

### Protected Access

1. `middleware.AuthMiddleware` reads and validates the bearer token.
2. Valid claims are stored in Echo context.
3. `middleware.RBACMiddleware` checks the claims against the required permission.
4. The handler reads claims from context and returns the response.

## Dependency Direction

The intended dependency direction is:

- `transport/http` depends on `usecase` and `port`
- `usecase` depends on `domain` and `port`
- `adapter` depends on `port` and sometimes `domain`
- `domain` depends on nothing in the service

Nothing in `domain` or `usecase` should depend on Echo, JWT, Argon2id, or storage details.

## Composition Root

`internal/iam/bootstrap/bootstrap.go` is responsible for wiring the concrete adapters into the application:

- create repositories (memory or PostgreSQL based on config)
- create security adapters (Argon2id hasher, RS256 JWT signer)
- create the event publisher
- create the auth usecase
- create the HTTP handler and router
- start the Echo server
- run database migrations and seed data when using PostgreSQL

That is the only place where concrete implementations should be assembled.

`cmd/iam/main.go` should stay thin and only pick configuration, log startup, and call the bootstrap layer.

For local development, `cmd/iam/main.go` loads a `.env` file before reading process environment variables. Production can still rely on real environment injection.

## Extension Points

The current structure is set up to grow without forcing a rewrite:

- add SQL or Redis repositories by implementing the same ports in `internal/iam/port/repository/`
- add new HTTP endpoints by adding a handler file in `transport/http/handler/`
- add new use cases without changing the adapters
- replace Argon2id or JWT implementations without changing the application logic
- add event bus integration by replacing the `noopEventPublisher`

## Notes

- The `transport/http/handler/` package stores claims in Echo context using a package-local key so the HTTP layer can share auth state without leaking it into the application layer.
- The JWT signer uses RSA keys and exposes a JWKS document; if no private key is configured, a development key is generated at startup.
- Refresh tokens are stored as SHA256 hashes — the raw token is only returned at creation time.
- Rate limiting is applied per IP: 5 requests/minute on login, 20 requests/minute on refresh.
- PostgreSQL is the production database; in-memory is used when `DATABASE_URL` is not set.
