package apperror

import "net/http"

// ============================================================
// COMMON — shared across all layers
// ============================================================

var (
	ErrInternal       = Internal("INTERNAL_ERROR", "Internal server error")
	ErrBadRequest     = BadRequest("BAD_REQUEST", "Invalid request")
	ErrValidation     = BadRequest("VALIDATION_ERROR", "Validation failed")
	ErrNotFound       = NotFound("NOT_FOUND", "Resource not found")
	ErrForbidden      = Forbidden("FORBIDDEN", "Access denied")
	ErrTooManyRequest = TooManyRequests("RATE_LIMITED", "Too many requests, try again later")
)

// ============================================================
// DOMAIN — business rule violations
// ============================================================

// --- User ---
var (
	ErrInvalidEmail    = BadRequest("USER_INVALID_EMAIL", "Invalid email format")
	ErrInvalidPassword = BadRequest("USER_INVALID_PASSWORD", "Password must be at least 8 characters")
	ErrInvalidName     = BadRequest("USER_INVALID_NAME", "Name cannot be empty")
	ErrUserNotFound    = NotFound("USER_NOT_FOUND", "User not found")
	ErrEmailExists     = Conflict("USER_EMAIL_EXISTS", "This email is already registered")
)

// --- Session ---
var (
	ErrSessionNotFound = Unauthorized("SESSION_NOT_FOUND", "Session not found")
	ErrSessionExpired  = Unauthorized("SESSION_EXPIRED", "Session has expired")
	ErrInvalidToken    = Unauthorized("SESSION_INVALID_TOKEN", "Invalid token")
)

// --- Auth ---
var (
	ErrInvalidCredentials = Unauthorized("AUTH_INVALID_CREDENTIALS", "Invalid email or password")
	ErrAccountLocked      = Forbidden("AUTH_ACCOUNT_LOCKED", "Account is locked due to multiple failed attempts")
	ErrTwoFactorRequired  = New("AUTH_2FA_REQUIRED", http.StatusPreconditionRequired, "Two-factor authentication required")
	ErrInvalidOTP         = BadRequest("AUTH_INVALID_OTP", "Invalid or expired OTP code")
)

// --- Account ---
var (
	ErrAccountNotFound      = NotFound("ACCOUNT_NOT_FOUND", "Account not found")
	ErrAccountFrozen        = Forbidden("ACCOUNT_FROZEN", "Account is frozen")
	ErrInsufficientFunds    = BadRequest("ACCOUNT_INSUFFICIENT_FUNDS", "Insufficient balance")
	ErrDailyLimitExceeded   = BadRequest("ACCOUNT_DAILY_LIMIT", "Daily transaction limit exceeded")
	ErrMonthlyLimitExceeded = BadRequest("ACCOUNT_MONTHLY_LIMIT", "Monthly transaction limit exceeded")
)

// --- Transfer ---
var (
	ErrTransferNotFound    = NotFound("TRANSFER_NOT_FOUND", "Transfer not found")
	ErrTransferFailed      = Internal("TRANSFER_FAILED", "Transfer processing failed")
	ErrSagaCompensation    = Internal("TRANSFER_SAGA_COMPENSATION", "Transfer failed, compensating")
	ErrDuplicateTransfer   = Conflict("TRANSFER_DUPLICATE", "Duplicate transfer (idempotency key exists)")
	ErrSameAccount         = BadRequest("TRANSFER_SAME_ACCOUNT", "Cannot transfer to the same account")
	ErrInvalidAmount       = BadRequest("TRANSFER_INVALID_AMOUNT", "Transfer amount must be greater than zero")
	ErrTransferTimeout     = New("TRANSFER_TIMEOUT", http.StatusGatewayTimeout, "Transfer processing timed out")
)

// --- Card ---
var (
	ErrCardNotFound       = NotFound("CARD_NOT_FOUND", "Card not found")
	ErrCardBlocked        = Forbidden("CARD_BLOCKED", "Card is blocked")
	ErrCardExpired        = BadRequest("CARD_EXPIRED", "Card has expired")
	ErrInvalidPAN         = BadRequest("CARD_INVALID_PAN", "Invalid card number (Luhn check failed)")
	ErrInvalidPIN         = BadRequest("CARD_INVALID_PIN", "Invalid PIN")
	ErrInvalidCVV         = BadRequest("CARD_INVALID_CVV", "Invalid CVV")
	ErrPINAttemptsExceeded = Forbidden("CARD_PIN_LOCKED", "Card locked: too many wrong PIN attempts")
	ErrCardLimitExceeded  = BadRequest("CARD_LIMIT_EXCEEDED", "Card transaction limit exceeded")
)

// --- KYC / AML ---
var (
	ErrKYCRequired      = Forbidden("KYC_REQUIRED", "KYC verification required")
	ErrKYCPending       = Forbidden("KYC_PENDING", "KYC verification is pending")
	ErrKYCRejected      = Forbidden("KYC_REJECTED", "KYC verification was rejected")
	ErrAMLBlocked       = Forbidden("AML_BLOCKED", "Transaction blocked by AML screening")
	ErrAMLFlagged       = New("AML_FLAGGED", http.StatusPreconditionRequired, "Transaction flagged for AML review")
)

// --- Fraud ---
var (
	ErrFraudDetected    = Forbidden("FRAUD_DETECTED", "Suspicious activity detected")
	ErrFraudHighRisk    = Forbidden("FRAUD_HIGH_RISK", "Transaction blocked: high risk score")
	ErrDeviceNotTrusted = Forbidden("FRAUD_DEVICE_UNTRUSTED", "Unrecognized device, verification required")
)

// --- Beneficiary ---
var (
	ErrBeneficiaryNotFound = NotFound("BENEFICIARY_NOT_FOUND", "Beneficiary not found")
	ErrBeneficiaryExists   = Conflict("BENEFICIARY_EXISTS", "Beneficiary already exists")
)

// ============================================================
// APPLICATION — use case / orchestration errors
// ============================================================

var (
	ErrIdempotencyConflict = Conflict("APP_IDEMPOTENCY_CONFLICT", "Request already processed (idempotency)")
	ErrConcurrencyConflict = Conflict("APP_CONCURRENCY_CONFLICT", "Resource was modified by another request")
)

// ============================================================
// INFRASTRUCTURE — external system errors
// ============================================================

var (
	ErrDatabase       = Internal("INFRA_DB_ERROR", "Database operation failed")
	ErrDBTimeout      = ServiceUnavailable("INFRA_DB_TIMEOUT", "Database connection timed out")
	ErrDBDeadlock     = Internal("INFRA_DB_DEADLOCK", "Database deadlock detected, please retry")
	ErrRedis          = Internal("INFRA_REDIS_ERROR", "Redis operation failed")
	ErrRedisTimeout   = ServiceUnavailable("INFRA_REDIS_TIMEOUT", "Redis connection timed out")
	ErrKafka          = Internal("INFRA_KAFKA_ERROR", "Kafka operation failed")
	ErrKafkaTimeout   = ServiceUnavailable("INFRA_KAFKA_TIMEOUT", "Kafka connection timed out")
	ErrVault          = Internal("INFRA_VAULT_ERROR", "Vault operation failed")
	ErrVaultTimeout   = ServiceUnavailable("INFRA_VAULT_TIMEOUT", "Vault connection timed out")
	ErrEncryption     = Internal("INFRA_ENCRYPTION_ERROR", "Encryption/decryption failed")
	ErrE2EEDecrypt    = BadRequest("INFRA_E2EE_DECRYPT_FAILED", "E2EE decryption failed (invalid key or corrupted data)")
)

// ============================================================
// HTTP — request/response level errors
// ============================================================

var (
	ErrInvalidJSON      = BadRequest("HTTP_INVALID_JSON", "Invalid JSON format")
	ErrMissingField     = BadRequest("HTTP_MISSING_FIELD", "Required field is missing")
	ErrInvalidParam     = BadRequest("HTTP_INVALID_PARAM", "Invalid URL parameter")
	ErrUnauthorized     = Unauthorized("HTTP_UNAUTHORIZED", "Authentication required")
	ErrTokenExpired     = Unauthorized("HTTP_TOKEN_EXPIRED", "Access token has expired")
	ErrTokenInvalid     = Unauthorized("HTTP_TOKEN_INVALID", "Invalid access token")
	ErrCSRFMismatch     = Forbidden("HTTP_CSRF_MISMATCH", "CSRF token mismatch")
	ErrCORSNotAllowed   = Forbidden("HTTP_CORS_NOT_ALLOWED", "Origin not allowed")
	ErrRequestTooLarge  = New("HTTP_REQUEST_TOO_LARGE", http.StatusRequestEntityTooLarge, "Request body too large")
)
