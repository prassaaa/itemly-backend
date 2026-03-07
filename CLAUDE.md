# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
go build ./...                  # Build all packages
go vet ./...                    # Static analysis
go run cmd/api/main.go          # Run the server (requires .env and PostgreSQL)
go mod tidy                     # Sync dependencies
```

No test framework is set up yet. When adding tests, use standard `go test ./...`.

## Environment Setup

Copy `.env.example` to `.env` and configure PostgreSQL credentials. The app reads config via Viper from `.env` in the project root. Database migrations run automatically on startup via GORM AutoMigrate.

## Architecture

Clean Architecture with manual dependency injection wired in `cmd/api/main.go`:

```
model → repository (interface) → usecase (interface) → handler → router
```

**Dependency flow:** `main.go` creates concrete implementations and injects them upward:
`Config → DB → JWTService → UserRepo → AuthUsecase → AuthHandler → Router`

### Layer conventions

- **model** (`internal/model/`): GORM entities with UUID primary keys (auto-generated via `BeforeCreate` hook) and soft delete. Password fields use `json:"-"`.
- **repository** (`internal/repository/`): Interface + GORM implementation in the same file. Interface defined first, then struct + constructor.
- **usecase** (`internal/usecase/`): Interface + implementation in the same file. Business errors defined as package-level `var` sentinel errors.
- **handler** (`internal/delivery/http/handler/`): Gin handlers. Binds request DTOs with `ShouldBindJSON`, maps usecase errors to HTTP status codes.
- **dto** (`internal/delivery/http/dto/`): Request DTOs (with `binding` validation tags) and Response DTOs are separate from models. Use `ToXxxResponse()` converter functions.
- **middleware** (`internal/delivery/http/middleware/`): JWT auth extracts `userID` (uuid.UUID), `username` (string), `role` (string) into Gin context.
- **router** (`internal/delivery/http/router.go`): All routes under `/api/v1`. Public, JWT-protected, and role-protected groups.
- **pkg** (`pkg/`): Reusable utilities. Note: `pkg/jwt` uses package name `jwtutil` (not `jwt`) to avoid conflict with the jwt library.

### Roles

Three roles: `admin`, `manager`, `staff`. Default is `staff`. Role checking via `middleware.RoleAuth(model.RoleAdmin, ...)`.

## Conventions

- Module path: `github.com/prassaaa/itemly-backend`
- All API routes are versioned under `/api/v1`
- Interfaces and implementations coexist in the same file per layer
- Constructor functions follow `NewXxx` pattern returning the interface type
- Config uses `mapstructure` tags matching `.env` variable names
