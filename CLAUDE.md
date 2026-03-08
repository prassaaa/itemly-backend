# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
go build ./...                  # Build all packages
go vet ./...                    # Static analysis
go run cmd/api/main.go          # Run the server (requires .env, PostgreSQL, Redis)
go mod tidy                     # Sync dependencies
```

## Test Commands

```bash
make test                       # Run all tests
make test-verbose               # Verbose output
make test-cover                 # Coverage report (terminal)
make test-cover-html            # Coverage report (browser)
make k6-auth                    # k6 load test: auth flow
make k6-ratelimit               # k6 load test: rate limiter
make k6                         # Run all k6 tests
make k6-cleanup                 # Delete loadtest users from database
```

### Test structure

- Tests use `testify` (assert/require) and hand-written mocks colocated in `_test.go` files
- Redis tests use `miniredis` (in-memory, no Redis server needed)
- Mock pattern: struct with function pointer fields, nil = panic on unexpected call
- `internal/testutil/jwt.go` provides shared helpers: `TestJWTService()`, `TestAccessToken()`, `TestRefreshToken()`
- `*jwtutil.JWTService` is concrete (not interface) — tests use real instance with test secret
- k6 scripts in `k6/` require running server + Redis
- `cmd/cleanup/main.go` deletes users with `%@loadtest.com` emails

## Environment Setup

Copy `.env.example` to `.env` and configure PostgreSQL + Redis credentials. The app reads config via Viper from `.env` in the project root. Database migrations run automatically on startup via GORM AutoMigrate. Permissions are seeded via `database.SeedPermissions()` on every startup (idempotent).

## Architecture

Clean Architecture with manual dependency injection wired in `cmd/api/main.go`:

```
model → repository (interface) → usecase (interface) → handler → router
```

**Dependency flow:** `main.go` creates concrete implementations and injects them upward:
`Config → DB/Redis → JWTService → TokenBlacklist → RateLimiter → Repos → Usecases → Handlers → Router`

### Layer conventions

- **model** (`internal/model/`): GORM entities with UUID primary keys (auto-generated via `BeforeCreate` hook) and soft delete. Password fields use `json:"-"`.
- **repository** (`internal/repository/`): Interface + GORM implementation in the same file. Interface defined first, then struct + constructor.
- **usecase** (`internal/usecase/`): Interface + implementation in the same file. Business errors defined as package-level `var` sentinel errors. Two permission implementations: in-memory (`permission_usecase.go`) and Redis-backed (`permission_usecase_redis.go`).
- **handler** (`internal/delivery/http/handler/`): Gin handlers. Binds request DTOs with `ShouldBindJSON`, maps usecase errors to HTTP status codes.
- **dto** (`internal/delivery/http/dto/`): Request DTOs (with `binding` validation tags) and Response DTOs are separate from models. Use `ToXxxResponse()` converter functions.
- **middleware** (`internal/delivery/http/middleware/`): JWT auth extracts `userID` (uuid.UUID), `username` (string), `role` (string), `jti`, `tokenExpiresAt` into Gin context. Permission-based auth via `PermissionAuth()`. Redis-backed rate limiter via `RateLimiter` interface.
- **router** (`internal/delivery/http/router.go`): All routes under `/api/v1`. Public (with rate limit), JWT-protected, and permission-protected groups.
- **pkg** (`pkg/`): Reusable utilities. `pkg/jwt` (package name `jwtutil`) has JWT service + `TokenBlacklist` interface with Redis implementation. `pkg/hash` wraps bcrypt. `pkg/validator` has custom password strength validator.
- **testutil** (`internal/testutil/`): Shared test helpers for JWT token generation with test secrets.

### Auth system

- JWT access + refresh token pair
- Token blacklist via Redis (key: `blacklist:<jti>`, TTL = remaining token lifetime)
- Refresh token rotation: old refresh token blacklisted on use
- Custom `password` validator registered on Gin's binding engine

### Roles & Permissions

Three roles: `admin`, `manager`, `staff`. Default is `staff`. Permission checking via `middleware.PermissionAuth(permUsecase, "permission:name")`. Permissions cached in Redis hashes (key: `perm:<role>`, fields: permission names). 15 role-permission mappings seeded on startup.

## Conventions

- Module path: `github.com/prassaaa/itemly-backend`
- All API routes are versioned under `/api/v1`
- Interfaces and implementations coexist in the same file per layer
- Constructor functions follow `NewXxx` pattern returning the interface type
- Config uses `mapstructure` tags matching `.env` variable names
- Sentinel errors for business logic (e.g., `ErrEmailAlreadyExists`, `ErrInvalidCredentials`)
- PostgreSQL unique constraint violations (pgconn.PgError code 23505) mapped to sentinel errors
