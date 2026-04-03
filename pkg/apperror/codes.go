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
	ErrInternal = Internal(1000, "Internal server error").
			WithType("INTERNAL_ERROR").
			WithSeverity(SeverityHigh).
			WithCategory(CategorySystem)

	ErrBadRequest = BadRequest(1001, "Invalid request").
			WithType("BAD_REQUEST").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)

	ErrValidation = BadRequest(1002, "Validation failed").
			WithType("VALIDATION_ERROR").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)

	ErrNotFound = NotFound(1003, "Resource not found").
			WithType("NOT_FOUND").
			WithSeverity(SeverityLow).
			WithCategory(CategoryData)

	ErrForbidden = Forbidden(1004, "Access denied").
			WithType("FORBIDDEN").
			WithSeverity(SeverityMedium).
			WithCategory(CategorySecurity)

	ErrTooManyRequest = TooManyRequests(1005, "Too many requests, try again later").
				WithType("TOO_MANY_REQUESTS").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)

	ErrInvalidJSON = BadRequest(1006, "Invalid JSON format").
			WithType("INVALID_JSON").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)

	ErrMissingField = BadRequest(1007, "Required field is missing").
			WithType("MISSING_FIELD").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)

	ErrInvalidParam = BadRequest(1008, "Invalid parameter").
			WithType("INVALID_PARAM").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)
)

// --- User (2xxx) ---
var (
	ErrInvalidEmail = BadRequest(2001, "Invalid email format").
			WithType("INVALID_EMAIL").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)

	ErrInvalidPassword = BadRequest(2002, "Password must be at least 8 characters").
				WithType("INVALID_PASSWORD").
				WithSeverity(SeverityLow).
				WithCategory(CategoryValidation)

	ErrInvalidName = BadRequest(2003, "Name cannot be empty").
			WithType("INVALID_NAME").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)

	ErrUserNotFound = NotFound(2004, "User not found").
			WithType("USER_NOT_FOUND").
			WithSeverity(SeverityLow).
			WithCategory(CategoryData)

	ErrEmailExists = Conflict(2005, "This email is already registered").
			WithType("EMAIL_EXISTS").
			WithSeverity(SeverityLow).
			WithCategory(CategoryBusiness)
)

// --- Session / Auth (3xxx) ---
var (
	ErrSessionNotFound = Unauthorized(3001, "Session not found").
				WithType("SESSION_NOT_FOUND").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)

	ErrSessionExpired = Unauthorized(3002, "Session has expired").
				WithType("SESSION_EXPIRED").
				WithSeverity(SeverityLow).
				WithCategory(CategorySecurity)

	ErrInvalidToken = Unauthorized(3003, "Invalid token").
			WithType("INVALID_TOKEN").
			WithSeverity(SeverityMedium).
			WithCategory(CategorySecurity)

	ErrInvalidCredentials = Unauthorized(3004, "Invalid email or password").
				WithType("INVALID_CREDENTIALS").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)

	ErrUnauthorized = Unauthorized(3005, "Authentication required").
			WithType("UNAUTHORIZED").
			WithSeverity(SeverityMedium).
			WithCategory(CategorySecurity)

	ErrTokenExpired = Unauthorized(3006, "Access token has expired").
			WithType("TOKEN_EXPIRED").
			WithSeverity(SeverityLow).
			WithCategory(CategorySecurity)

	ErrTokenInvalid = Unauthorized(3007, "Invalid access token").
			WithType("TOKEN_INVALID").
			WithSeverity(SeverityMedium).
			WithCategory(CategorySecurity)
)

// --- Account (4xxx) ---
var (
	ErrAccountNotFound = NotFound(4001, "Account not found").
				WithType("ACCOUNT_NOT_FOUND").
				WithSeverity(SeverityLow).
				WithCategory(CategoryData)

	ErrAccountFrozen = Forbidden(4002, "Account is frozen").
			WithType("ACCOUNT_FROZEN").
			WithSeverity(SeverityMedium).
			WithCategory(CategoryBusiness)

	ErrAccountClosed = Forbidden(4003, "Account is closed").
			WithType("ACCOUNT_CLOSED").
			WithSeverity(SeverityMedium).
			WithCategory(CategoryBusiness)

	ErrBalanceNotZero = BadRequest(4004, "Account balance must be 0 to close").
				WithType("BALANCE_NOT_ZERO").
				WithSeverity(SeverityLow).
				WithCategory(CategoryBusiness)

	ErrInsufficientFunds = BadRequest(4005, "Insufficient balance").
				WithType("INSUFFICIENT_FUNDS").
				WithSeverity(SeverityMedium).
				WithCategory(CategoryBusiness)
)

// --- Transfer (5xxx) ---
var (
	ErrTransferNotFound = NotFound(5001, "Transfer not found").
				WithType("TRANSFER_NOT_FOUND").
				WithSeverity(SeverityLow).
				WithCategory(CategoryData)

	ErrTransferFailed = Internal(5002, "Transfer processing failed").
				WithType("TRANSFER_FAILED").
				WithSeverity(SeverityHigh).
				WithCategory(CategorySystem)

	ErrSameAccount = BadRequest(5003, "Cannot transfer to the same account").
			WithType("SAME_ACCOUNT").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)

	ErrInvalidAmount = BadRequest(5004, "Amount must be greater than zero").
			WithType("INVALID_AMOUNT").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)

	ErrCurrencyMismatch = BadRequest(5005, "Currencies do not match").
				WithType("CURRENCY_MISMATCH").
				WithSeverity(SeverityLow).
				WithCategory(CategoryBusiness)
)

// --- Card (6xxx) ---
var (
	ErrCardNotFound = NotFound(6001, "Card not found").
			WithType("CARD_NOT_FOUND").
			WithSeverity(SeverityLow).
			WithCategory(CategoryData)

	ErrCardBlocked = Forbidden(6002, "Card is blocked").
			WithType("CARD_BLOCKED").
			WithSeverity(SeverityMedium).
			WithCategory(CategoryBusiness)

	ErrCardExpired = BadRequest(6003, "Card has expired").
			WithType("CARD_EXPIRED").
			WithSeverity(SeverityLow).
			WithCategory(CategoryBusiness)

	ErrInvalidPIN = BadRequest(6004, "Invalid PIN").
			WithType("INVALID_PIN").
			WithSeverity(SeverityMedium).
			WithCategory(CategorySecurity)

	ErrPINAttemptsExceeded = Forbidden(6005, "Card locked: too many wrong PIN attempts").
				WithType("PIN_ATTEMPTS_EXCEEDED").
				WithSeverity(SeverityHigh).
				WithCategory(CategorySecurity)
)

// --- KYC / AML / Fraud (7xxx) ---
var (
	ErrKYCRequired = Forbidden(7001, "KYC verification required").
			WithType("KYC_REQUIRED").
			WithSeverity(SeverityMedium).
			WithCategory(CategoryBusiness)

	ErrKYCPending = Forbidden(7002, "KYC verification is pending").
			WithType("KYC_PENDING").
			WithSeverity(SeverityLow).
			WithCategory(CategoryBusiness)

	ErrAMLBlocked = Forbidden(7003, "Transaction blocked by AML screening").
			WithType("AML_BLOCKED").
			WithSeverity(SeverityHigh).
			WithCategory(CategorySecurity)

	ErrFraudDetected = Forbidden(7004, "Suspicious activity detected").
			WithType("FRAUD_DETECTED").
			WithSeverity(SeverityCritical).
			WithCategory(CategorySecurity)
)

// --- Beneficiary (8xxx) ---
var (
	ErrBeneficiaryNotFound = NotFound(8001, "Beneficiary not found").
				WithType("BENEFICIARY_NOT_FOUND").
				WithSeverity(SeverityLow).
				WithCategory(CategoryData)

	ErrBeneficiaryExists = Conflict(8002, "Beneficiary already exists").
				WithType("BENEFICIARY_EXISTS").
				WithSeverity(SeverityLow).
				WithCategory(CategoryBusiness)
)

// --- Contact (86xx) ---
var (
	ErrContactNotFound = NotFound(8601, "Contact not found").
				WithType("CONTACT_NOT_FOUND").
				WithSeverity(SeverityLow).
				WithCategory(CategoryData)

	ErrContactExists = Conflict(8602, "Contact already exists").
			WithType("CONTACT_EXISTS").
			WithSeverity(SeverityLow).
			WithCategory(CategoryBusiness)

	ErrContactSelf = BadRequest(8603, "Cannot add yourself as a contact").
			WithType("CONTACT_SELF").
			WithSeverity(SeverityLow).
			WithCategory(CategoryValidation)
)

// --- Application (85xx) ---
var (
	ErrConcurrencyConflict = Conflict(8501, "Resource was modified by another request, please retry").
				WithType("CONCURRENCY_CONFLICT").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySystem)
)

// --- TOTP / 2FA (35xx) ---
var (
	ErrTOTPNotEnabled = BadRequest(3500, "TOTP is not enabled for this account").
				WithType("TOTP_NOT_ENABLED").
				WithSeverity(SeverityLow).
				WithCategory(CategoryBusiness)

	ErrTOTPAlreadyEnabled = Conflict(3501, "TOTP is already enabled").
				WithType("TOTP_ALREADY_ENABLED").
				WithSeverity(SeverityLow).
				WithCategory(CategoryBusiness)

	ErrTOTPInvalidCode = Unauthorized(3502, "Invalid TOTP code").
				WithType("TOTP_INVALID_CODE").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)

	ErrTOTPRequired = Unauthorized(3503, "TOTP verification required").
			WithType("TOTP_REQUIRED").
			WithSeverity(SeverityMedium).
			WithCategory(CategorySecurity)
)

// --- Challenge / Step-Up Auth (39xx) ---
var (
	ErrChallengeRequired = Forbidden(3900, "Step-up authentication required").
				WithType("CHALLENGE_REQUIRED").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)

	ErrChallengeNotFound = NotFound(3901, "Challenge not found").
				WithType("CHALLENGE_NOT_FOUND").
				WithSeverity(SeverityLow).
				WithCategory(CategoryData)

	ErrChallengeExpired = Unauthorized(3902, "Challenge has expired").
				WithType("CHALLENGE_EXPIRED").
				WithSeverity(SeverityLow).
				WithCategory(CategorySecurity)

	ErrChallengeAlreadyUsed = BadRequest(3903, "Challenge has already been used or failed").
				WithType("CHALLENGE_ALREADY_USED").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)

	ErrChallengeTokenInvalid = Unauthorized(3904, "Invalid or expired challenge token").
				WithType("CHALLENGE_TOKEN_INVALID").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)

	ErrChallengeTokenMissing = Unauthorized(3905, "X-Challenge-Token header is required").
				WithType("CHALLENGE_TOKEN_MISSING").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)
)

// --- HMAC / Request Signing (38xx) ---
var (
	ErrHMACMissing = Unauthorized(3800, "Request signature required: X-Signature and X-Signature-Timestamp headers are missing").
			WithType("HMAC_MISSING").
			WithSeverity(SeverityHigh).
			WithCategory(CategorySecurity)

	ErrHMACTimestampInvalid = BadRequest(3801, "Invalid X-Signature-Timestamp: must be a Unix timestamp").
				WithType("HMAC_TIMESTAMP_INVALID").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)

	ErrHMACTimestampExpired = Unauthorized(3802, "Request signature expired: timestamp is too old or too far in the future").
				WithType("HMAC_TIMESTAMP_EXPIRED").
				WithSeverity(SeverityMedium).
				WithCategory(CategorySecurity)

	ErrHMACSignatureInvalid = Unauthorized(3803, "Request signature verification failed").
				WithType("HMAC_SIGNATURE_INVALID").
				WithSeverity(SeverityHigh).
				WithCategory(CategorySecurity)
)

// --- Exchange (45xx) ---
var (
	ErrRateNotFound = NotFound(4501, "Exchange rate not found").
			WithType("RATE_NOT_FOUND").
			WithSeverity(SeverityLow).
			WithCategory(CategoryData)
)

// --- Infrastructure (9xxx) ---
var (
	ErrDatabase = Internal(9001, "Database operation failed").
			WithType("DATABASE_ERROR").
			WithSeverity(SeverityHigh).
			WithCategory(CategorySystem)

	ErrDBTimeout = ServiceUnavailable(9002, "Database connection timed out").
			WithType("DB_TIMEOUT").
			WithSeverity(SeverityCritical).
			WithCategory(CategorySystem)
)
