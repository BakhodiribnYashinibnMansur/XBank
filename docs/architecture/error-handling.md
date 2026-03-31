# Error Handling

## Overview

XBank uses a centralized, structured error package (`pkg/apperror`) across all layers.
Every error carries a unique **error code**, HTTP status, and human-readable message.

## Error Response Format

All API errors return the same JSON structure:

```json
{
  "status": "error",
  "error": {
    "code": "AUTH_INVALID_CREDENTIALS",
    "message": "Invalid email or password"
  },
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "correlation_id": "corr-uuid",
    "timestamp": "2026-03-31T10:00:00Z"
  }
}
```

| Field | Description |
|---|---|
| `status` | Always `"error"` for error responses |
| `error.code` | Machine-readable error code (e.g. `USER_NOT_FOUND`) |
| `error.message` | Human-readable message (safe to display to end-user) |
| `meta.request_id` | Unique request ID for log tracing |
| `meta.correlation_id` | Cross-service correlation ID |
| `meta.timestamp` | UTC timestamp of the error |

## Package Structure

```
pkg/apperror/
├── error.go      # AppError struct, constructors, helpers
├── codes.go      # All error code constants (per layer, per domain)
└── response.go   # Fiber ErrorHandler, JSON response envelope
```

## AppError Struct

```go
type AppError struct {
    Code       string // Machine-readable: "USER_NOT_FOUND"
    Message    string // Human-readable: "User not found"
    HTTPStatus int    // HTTP status: 404
    Err        error  // Wrapped original error (for logging, never exposed)
}
```

### Key Methods

| Method | Description |
|---|---|
| `error.Wrap(err)` | Returns a copy with a wrapped cause (for logging) |
| `error.WithMessage(msg)` | Returns a copy with a custom message |
| `error.Error()` | `[CODE] message: cause` (for logs) |
| `error.Unwrap()` | Returns the wrapped cause (for `errors.Is/As`) |
| `error.Is(target)` | Compares by error code |

## Error Code Convention

```
{DOMAIN}_{DESCRIPTION}

Examples:
  USER_NOT_FOUND              — Domain: User
  AUTH_INVALID_CREDENTIALS    — Domain: Auth
  ACCOUNT_INSUFFICIENT_FUNDS  — Domain: Account
  TRANSFER_DUPLICATE          — Domain: Transfer
  CARD_PIN_LOCKED             — Domain: Card
  INFRA_DB_TIMEOUT            — Layer: Infrastructure
  HTTP_INVALID_JSON           — Layer: HTTP
```

## Error Codes by Layer

### Common

| Code | HTTP | Message |
|---|---|---|
| `INTERNAL_ERROR` | 500 | Internal server error |
| `BAD_REQUEST` | 400 | Invalid request |
| `VALIDATION_ERROR` | 400 | Validation failed |
| `NOT_FOUND` | 404 | Resource not found |
| `FORBIDDEN` | 403 | Access denied |
| `RATE_LIMITED` | 429 | Too many requests, try again later |

### Domain — User

| Code | HTTP | Message |
|---|---|---|
| `USER_INVALID_EMAIL` | 400 | Invalid email format |
| `USER_INVALID_PASSWORD` | 400 | Password must be at least 8 characters |
| `USER_INVALID_NAME` | 400 | Name cannot be empty |
| `USER_NOT_FOUND` | 404 | User not found |
| `USER_EMAIL_EXISTS` | 409 | This email is already registered |

### Domain — Auth & Session

| Code | HTTP | Message |
|---|---|---|
| `AUTH_INVALID_CREDENTIALS` | 401 | Invalid email or password |
| `AUTH_ACCOUNT_LOCKED` | 403 | Account is locked due to multiple failed attempts |
| `AUTH_2FA_REQUIRED` | 428 | Two-factor authentication required |
| `AUTH_INVALID_OTP` | 400 | Invalid or expired OTP code |
| `SESSION_NOT_FOUND` | 401 | Session not found |
| `SESSION_EXPIRED` | 401 | Session has expired |
| `SESSION_INVALID_TOKEN` | 401 | Invalid token |

### Domain — Account

| Code | HTTP | Message |
|---|---|---|
| `ACCOUNT_NOT_FOUND` | 404 | Account not found |
| `ACCOUNT_FROZEN` | 403 | Account is frozen |
| `ACCOUNT_INSUFFICIENT_FUNDS` | 400 | Insufficient balance |
| `ACCOUNT_DAILY_LIMIT` | 400 | Daily transaction limit exceeded |
| `ACCOUNT_MONTHLY_LIMIT` | 400 | Monthly transaction limit exceeded |

### Domain — Transfer

| Code | HTTP | Message |
|---|---|---|
| `TRANSFER_NOT_FOUND` | 404 | Transfer not found |
| `TRANSFER_FAILED` | 500 | Transfer processing failed |
| `TRANSFER_SAGA_COMPENSATION` | 500 | Transfer failed, compensating |
| `TRANSFER_DUPLICATE` | 409 | Duplicate transfer (idempotency key exists) |
| `TRANSFER_SAME_ACCOUNT` | 400 | Cannot transfer to the same account |
| `TRANSFER_INVALID_AMOUNT` | 400 | Transfer amount must be greater than zero |
| `TRANSFER_TIMEOUT` | 504 | Transfer processing timed out |

### Domain — Card

| Code | HTTP | Message |
|---|---|---|
| `CARD_NOT_FOUND` | 404 | Card not found |
| `CARD_BLOCKED` | 403 | Card is blocked |
| `CARD_EXPIRED` | 400 | Card has expired |
| `CARD_INVALID_PAN` | 400 | Invalid card number (Luhn check failed) |
| `CARD_INVALID_PIN` | 400 | Invalid PIN |
| `CARD_INVALID_CVV` | 400 | Invalid CVV |
| `CARD_PIN_LOCKED` | 403 | Card locked: too many wrong PIN attempts |
| `CARD_LIMIT_EXCEEDED` | 400 | Card transaction limit exceeded |

### Domain — KYC & AML

| Code | HTTP | Message |
|---|---|---|
| `KYC_REQUIRED` | 403 | KYC verification required |
| `KYC_PENDING` | 403 | KYC verification is pending |
| `KYC_REJECTED` | 403 | KYC verification was rejected |
| `AML_BLOCKED` | 403 | Transaction blocked by AML screening |
| `AML_FLAGGED` | 428 | Transaction flagged for AML review |

### Domain — Fraud

| Code | HTTP | Message |
|---|---|---|
| `FRAUD_DETECTED` | 403 | Suspicious activity detected |
| `FRAUD_HIGH_RISK` | 403 | Transaction blocked: high risk score |
| `FRAUD_DEVICE_UNTRUSTED` | 403 | Unrecognized device, verification required |

### Domain — Beneficiary

| Code | HTTP | Message |
|---|---|---|
| `BENEFICIARY_NOT_FOUND` | 404 | Beneficiary not found |
| `BENEFICIARY_EXISTS` | 409 | Beneficiary already exists |

### Application

| Code | HTTP | Message |
|---|---|---|
| `APP_IDEMPOTENCY_CONFLICT` | 409 | Request already processed (idempotency) |
| `APP_CONCURRENCY_CONFLICT` | 409 | Resource was modified by another request |

### Infrastructure

| Code | HTTP | Message |
|---|---|---|
| `INFRA_DB_ERROR` | 500 | Database operation failed |
| `INFRA_DB_TIMEOUT` | 503 | Database connection timed out |
| `INFRA_DB_DEADLOCK` | 500 | Database deadlock detected, please retry |
| `INFRA_REDIS_ERROR` | 500 | Redis operation failed |
| `INFRA_REDIS_TIMEOUT` | 503 | Redis connection timed out |
| `INFRA_KAFKA_ERROR` | 500 | Kafka operation failed |
| `INFRA_KAFKA_TIMEOUT` | 503 | Kafka connection timed out |
| `INFRA_VAULT_ERROR` | 500 | Vault operation failed |
| `INFRA_VAULT_TIMEOUT` | 503 | Vault connection timed out |
| `INFRA_ENCRYPTION_ERROR` | 500 | Encryption/decryption failed |
| `INFRA_E2EE_DECRYPT_FAILED` | 400 | E2EE decryption failed |

### HTTP

| Code | HTTP | Message |
|---|---|---|
| `HTTP_INVALID_JSON` | 400 | Invalid JSON format |
| `HTTP_MISSING_FIELD` | 400 | Required field is missing |
| `HTTP_INVALID_PARAM` | 400 | Invalid URL parameter |
| `HTTP_UNAUTHORIZED` | 401 | Authentication required |
| `HTTP_TOKEN_EXPIRED` | 401 | Access token has expired |
| `HTTP_TOKEN_INVALID` | 401 | Invalid access token |
| `HTTP_CSRF_MISMATCH` | 403 | CSRF token mismatch |
| `HTTP_CORS_NOT_ALLOWED` | 403 | Origin not allowed |
| `HTTP_REQUEST_TOO_LARGE` | 413 | Request body too large |

## Usage Examples

### Domain Layer

```go
// internal/domain/user/entity.go
import "github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"

func NewUser(email, hashedPassword, firstName, lastName string) (*User, error) {
    if email == "" {
        return nil, apperror.ErrInvalidEmail
    }
    if hashedPassword == "" {
        return nil, apperror.ErrInvalidPassword
    }
    if firstName == "" {
        return nil, apperror.ErrInvalidName
    }
    // ...
}
```

### Application Layer

```go
// internal/application/auth/service.go
import "github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
    u, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return nil, apperror.ErrInvalidCredentials // don't leak "user not found"
    }

    if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
        return nil, apperror.ErrInvalidCredentials
    }
    // ...
}
```

### Infrastructure Layer

```go
// internal/infrastructure/postgres/user_repository.go
import "github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
    row := r.db.QueryRow(ctx, "SELECT ... WHERE id = $1", id)
    if err := row.Scan(...); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, apperror.ErrUserNotFound
        }
        return nil, apperror.ErrDatabase.Wrap(err)  // wrap for logging
    }
    return u, nil
}
```

### HTTP Handler Layer

```go
// internal/interfaces/http/handler/user.go
import "github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"

func (h *UserHandler) Register(c *fiber.Ctx) error {
    var req dto.RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        return apperror.ErrInvalidJSON
    }

    if req.Email == "" || req.Password == "" || req.FirstName == "" {
        return apperror.ErrMissingField.WithMessage("email, password, and first_name are required")
    }

    u, err := h.service.Register(c.Context(), req.Email, req.Password, req.FirstName, req.LastName)
    if err != nil {
        return err  // AppError goes straight to ErrorHandler
    }

    return c.Status(http.StatusCreated).JSON(dto.UserResponse{...})
}
```

### Fiber App Setup

```go
// cmd/api/main.go
import "github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"

app := fiber.New(fiber.Config{
    ErrorHandler: apperror.ErrorHandler,
})
```

## Error Flow Architecture

```
Client Request
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│  HTTP Handler                                                │
│    ├── BodyParser fails    → return apperror.ErrInvalidJSON  │
│    ├── Validation fails    → return apperror.ErrMissingField │
│    └── Call application service                              │
│         │                                                    │
│         ▼                                                    │
│  Application Layer                                           │
│    ├── Business rule fails → return apperror.ErrXxx          │
│    └── Call repository                                       │
│         │                                                    │
│         ▼                                                    │
│  Infrastructure Layer                                        │
│    ├── DB no rows          → return apperror.ErrUserNotFound │
│    ├── DB timeout          → return apperror.ErrDBTimeout    │
│    └── DB error            → return apperror.ErrDatabase.    │
│                                          Wrap(originalErr)   │
└─────────────────────────────────────────────────────────────┘
    │
    │  error bubbles up (no catch/rethrow needed)
    ▼
┌─────────────────────────────────────────────────────────────┐
│  Fiber ErrorHandler (apperror.ErrorHandler)                  │
│                                                              │
│    if *AppError → structured JSON with code + message        │
│    if *fiber.Error → standard HTTP error                     │
│    else → 500 + generic message (never leak internals)       │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
{
  "status": "error",
  "error": { "code": "USER_EMAIL_EXISTS", "message": "..." },
  "meta":  { "request_id": "...", "timestamp": "..." }
}
```

## Rules

1. **Never expose internal errors** — `ErrDatabase.Wrap(err)` keeps `err` for logs but only returns `"Database operation failed"` to the client.
2. **Never leak user existence** — Use `ErrInvalidCredentials` for both "user not found" and "wrong password".
3. **Always use error codes** — Clients should match on `error.code`, not `error.message`. Messages may change, codes are stable.
4. **Infrastructure errors wrap** — `apperror.ErrDatabase.Wrap(pgErr)` preserves the original error for debugging.
5. **No switch statements in handlers** — Errors bubble up directly to `ErrorHandler`. Handlers just `return err`.
6. **Domain errors are sentinel** — `apperror.ErrUserNotFound` is a package-level var, comparable with `errors.Is()`.
