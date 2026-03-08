# Itemly Backend

Backend API untuk sistem manajemen inventaris, dibangun dengan **Go** menggunakan prinsip **Clean Architecture**.

## Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| Language | Go 1.26 |
| Router | Gin Gonic |
| ORM | GORM |
| Database | PostgreSQL |
| Cache & Rate Limit | Redis |
| Auth | JWT (access + refresh token) |
| Password | bcrypt |
| Config | Viper (`.env`) |
| API Docs | Swagger (swag) |

## Arsitektur

```
model → repository (interface) → usecase (interface) → handler → router
```

Dependency injection manual di `cmd/api/main.go`. Setiap layer punya interface dan implementasi dalam satu file.

```
.
├── cmd/
│   ├── api/            # Entry point server
│   └── cleanup/        # Cleanup loadtest data
├── internal/
│   ├── model/          # GORM entities (User, Permission, RolePermission)
│   ├── repository/     # Data access layer
│   ├── usecase/        # Business logic
│   ├── delivery/http/
│   │   ├── handler/    # Gin HTTP handlers
│   │   ├── middleware/  # JWT auth, RBAC, rate limit, security headers
│   │   ├── dto/        # Request/Response DTOs
│   │   └── router.go   # Route definitions
│   └── testutil/       # Shared test helpers
├── pkg/
│   ├── jwt/            # JWT service + Redis token blacklist
│   ├── hash/           # bcrypt wrapper
│   └── validator/      # Custom password validator
├── config/             # Viper config loader
├── database/           # PostgreSQL connection + seeder
├── k6/                 # Load test scripts
└── docs/               # Swagger generated docs
```

## API Endpoints

| Method | Path | Auth | Deskripsi |
|--------|------|------|-----------|
| `GET` | `/api/v1/health` | - | Health check |
| `POST` | `/api/v1/auth/register` | - | Register user baru |
| `POST` | `/api/v1/auth/login` | - | Login, returns token pair |
| `POST` | `/api/v1/auth/refresh` | - | Refresh token pair |
| `POST` | `/api/v1/auth/logout` | JWT | Revoke access token |
| `GET` | `/api/v1/profile` | JWT | Profile user saat ini |
| `GET` | `/api/v1/dashboard` | JWT + `dashboard:view` | Dashboard |
| `PUT` | `/api/v1/users/:id/role` | JWT + `users:manage` | Assign role ke user |

Endpoint auth (`register`, `login`, `refresh`) dilindungi rate limiter.

## Roles & Permissions

Tiga role: `admin`, `manager`, `staff` (default). Permission di-manage lewat tabel `role_permissions` dan di-cache ke Redis.

| Permission | admin | manager | staff |
|-----------|-------|---------|-------|
| `dashboard:view` | v | v | v |
| `users:view` | v | v | - |
| `users:manage` | v | - | - |
| `inventory:view` | v | v | v |
| `inventory:create` | v | v | - |
| `inventory:update` | v | v | - |
| `inventory:delete` | v | v | - |

## Getting Started

### Prerequisites

- Go 1.26+
- PostgreSQL
- Redis

### Setup

```bash
# Clone
git clone https://github.com/prassaaa/itemly-backend.git
cd itemly-backend

# Install dependencies
go mod tidy

# Config
cp .env.example .env
# Edit .env sesuai kredensial database kamu

# Jalankan server (migrasi & seed otomatis)
go run cmd/api/main.go
```

Server jalan di `http://localhost:8080`. Swagger docs di `http://localhost:8080/swagger/index.html` (development mode).

## Testing

### Unit & Integration Test

```bash
make test             # Jalankan semua test
make test-verbose     # Verbose output
make test-cover       # Coverage report (terminal)
make test-cover-html  # Coverage report (browser)
```

**Coverage:**

| Package | Coverage |
|---------|----------|
| `handler` | 89.2% |
| `middleware` | 69.8% |
| `usecase` | 69.0% |
| `pkg/jwt` | 80.0% |
| `pkg/hash` | 100% |
| `pkg/validator` | 100% |
| `model` | 100% |
| `dto` | 100% |

### k6 Load Test

Butuh server + Redis running dan [k6](https://k6.io/) terinstall.

```bash
make k6-auth          # Auth flow: register → login → profile → refresh → logout
make k6-ratelimit     # Rate limiter test: 20 rapid requests
make k6               # Jalankan semua k6 test
make k6-cleanup       # Hapus user loadtest dari database
```

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `APP_PORT` | `8080` | Port server |
| `APP_ENV` | `development` | `development` / `production` |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `itemly` | Nama database |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_SECRET` | - | Secret key untuk JWT |
| `JWT_EXPIRY_HOURS` | `24` | Masa berlaku access token (jam) |
| `JWT_REFRESH_EXPIRY_HOURS` | `168` | Masa berlaku refresh token (jam) |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | - | Redis password |
| `REDIS_DB` | `0` | Redis database number |
| `RATE_LIMIT_RPS` | `1` | Rate limit per detik |
| `RATE_LIMIT_BURST` | `5` | Rate limit burst |
| `MAX_BODY_SIZE` | `1048576` | Max request body (bytes) |
| `CORS_ALLOWED_ORIGINS` | `localhost:3000,localhost:5173` | Allowed CORS origins |

## Contributing

1. Fork repository ini
2. Buat branch fitur (`git checkout -b feature/nama-fitur`)
3. Commit perubahan (`git commit -m 'feat: tambah fitur baru'`)
4. Push ke branch (`git push origin feature/nama-fitur`)
5. Buat Pull Request

## License

Project ini dilisensikan di bawah [MIT License](LICENSE).

---
