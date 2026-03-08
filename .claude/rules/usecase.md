---
paths:
  - "internal/usecase/**/*.go"
---
- Interface dan implementation dalam satu file
- Business errors sebagai package-level `var` sentinel errors (e.g., `ErrUserNotFound`)
- Gunakan `errors.Is()` untuk check error, bukan string comparison
- PostgreSQL unique constraint (pgconn.PgError code 23505) di-map ke sentinel error
- Jangan akses HTTP layer (gin.Context) dari usecase — terima parameter primitif/model saja
