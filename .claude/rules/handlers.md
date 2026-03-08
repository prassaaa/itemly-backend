---
paths:
  - "internal/delivery/http/handler/**/*.go"
---
- Selalu bind request DTO dengan `ShouldBindJSON`
- Map usecase sentinel errors ke HTTP status code yang sesuai
- Jangan return model struct langsung, gunakan `ToXxxResponse()` converter
- Format validation errors lewat `formatValidationErrors()`
- Error 500 selalu return generic message "internal server error", jangan expose detail
