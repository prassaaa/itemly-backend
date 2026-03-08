---
paths:
  - "internal/model/**/*.go"
  - "internal/repository/**/*.go"
---
- Model pakai UUID primary key dengan `BeforeCreate` hook untuk auto-generate
- Semua model pakai `gorm.DeletedAt` untuk soft delete
- Password field wajib pakai `json:"-"` tag
- Repository: interface didefinisikan duluan, lalu struct + constructor di file yang sama
- Constructor return interface type, bukan concrete struct
