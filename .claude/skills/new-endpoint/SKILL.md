---
name: new-endpoint
description: Scaffold endpoint baru mengikuti clean architecture
---
Buat API endpoint baru untuk $ARGUMENTS mengikuti clean architecture project ini:

1. **Model** — Tambah struct di `internal/model/` dengan UUID PK, GORM tags, `BeforeCreate` hook, dan soft delete
2. **Repository** — Buat interface + GORM implementation di `internal/repository/` (satu file)
3. **Usecase** — Buat interface + implementation di `internal/usecase/` (satu file) dengan sentinel errors
4. **DTO** — Tambah request DTO (dengan `binding` validation tags) dan response DTO di `internal/delivery/http/dto/`
5. **Handler** — Buat Gin handler di `internal/delivery/http/handler/` yang bind DTO dan map errors ke HTTP status
6. **Router** — Register route di `internal/delivery/http/router.go` di group yang sesuai (public/protected/permission)
7. **Wiring** — Tambah dependency injection di `cmd/api/main.go`
8. **Test** — Buat `_test.go` untuk usecase dan handler dengan hand-written mocks

Ikuti pattern yang sudah ada di auth/admin endpoints sebagai referensi.
