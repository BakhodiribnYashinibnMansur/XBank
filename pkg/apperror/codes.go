package apperror

// ============================================================
// Error code ranges:
//   1xxx — Common
//   2xxx — User
//   3xxx — Session / Auth
//   4xxx — Account
//   5xxx — Transfer
//   6xxx — Card
//   7xxx — KYC / AML / Fraud
//   8xxx — Beneficiary
//   9xxx — Infrastructure
// ============================================================

// --- Common (1xxx) ---
var (
	ErrInternal       = Internal(1000, "Internal server error")
	ErrBadRequest     = BadRequest(1001, "Invalid request")
	ErrValidation     = BadRequest(1002, "Validation failed")
	ErrNotFound       = NotFound(1003, "Resource not found")
	ErrForbidden      = Forbidden(1004, "Access denied")
	ErrTooManyRequest = TooManyRequests(1005, "Too many requests, try again later")
	ErrInvalidJSON    = BadRequest(1006, "Invalid JSON format")
	ErrMissingField   = BadRequest(1007, "Required field is missing")
	ErrInvalidParam   = BadRequest(1008, "Invalid parameter")
)

// --- User (2xxx) ---
var (
	ErrInvalidEmail    = BadRequest(2001, "Invalid email format")
	ErrInvalidPassword = BadRequest(2002, "Password must be at least 8 characters")
	ErrInvalidName     = BadRequest(2003, "Name cannot be empty")
	ErrUserNotFound    = NotFound(2004, "User not found")
	ErrEmailExists     = Conflict(2005, "This email is already registered")
)

// --- Session / Auth (3xxx) ---
var (
	ErrSessionNotFound    = Unauthorized(3001, "Session not found")
	ErrSessionExpired     = Unauthorized(3002, "Session has expired")
	ErrInvalidToken       = Unauthorized(3003, "Invalid token")
	ErrInvalidCredentials = Unauthorized(3004, "Invalid email or password")
	ErrUnauthorized       = Unauthorized(3005, "Authentication required")
	ErrTokenExpired       = Unauthorized(3006, "Access token has expired")
	ErrTokenInvalid       = Unauthorized(3007, "Invalid access token")
)

// --- Account (4xxx) ---
var (
	ErrAccountNotFound   = NotFound(4001, "Account not found")
	ErrAccountFrozen     = Forbidden(4002, "Account is frozen")
	ErrAccountClosed     = Forbidden(4003, "Account is closed")
	ErrBalanceNotZero    = BadRequest(4004, "Account balance must be 0 to close")
	ErrInsufficientFunds = BadRequest(4005, "Insufficient balance")
)

// --- Transfer (5xxx) ---
var (
	ErrTransferNotFound = NotFound(5001, "Transfer not found")
	ErrTransferFailed   = Internal(5002, "Transfer processing failed")
	ErrSameAccount      = BadRequest(5003, "Cannot transfer to the same account")
	ErrInvalidAmount    = BadRequest(5004, "Amount must be greater than zero")
	ErrCurrencyMismatch = BadRequest(5005, "Currencies do not match")
)

// --- Card (6xxx) ---
var (
	ErrCardNotFound        = NotFound(6001, "Card not found")
	ErrCardBlocked         = Forbidden(6002, "Card is blocked")
	ErrCardExpired         = BadRequest(6003, "Card has expired")
	ErrInvalidPIN          = BadRequest(6004, "Invalid PIN")
	ErrPINAttemptsExceeded = Forbidden(6005, "Card locked: too many wrong PIN attempts")
)

// --- KYC / AML / Fraud (7xxx) ---
var (
	ErrKYCRequired   = Forbidden(7001, "KYC verification required")
	ErrKYCPending    = Forbidden(7002, "KYC verification is pending")
	ErrAMLBlocked    = Forbidden(7003, "Transaction blocked by AML screening")
	ErrFraudDetected = Forbidden(7004, "Suspicious activity detected")
)

// --- Beneficiary (8xxx) ---
var (
	ErrBeneficiaryNotFound = NotFound(8001, "Beneficiary not found")
	ErrBeneficiaryExists   = Conflict(8002, "Beneficiary already exists")
)

// --- Application (85xx) ---
var (
	ErrConcurrencyConflict = Conflict(8501, "Resource was modified by another request, please retry")
)

// --- Infrastructure (9xxx) ---
var (
	ErrDatabase  = Internal(9001, "Database operation failed")
	ErrDBTimeout = ServiceUnavailable(9002, "Database connection timed out")
)
