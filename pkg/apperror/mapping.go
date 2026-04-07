package apperror

import "errors"

// domainCodeToAppError maps DomainError.Code() strings to their AppError equivalents.
var domainCodeToAppError = map[string]*AppError{
	// Common
	"MISSING_FIELD":    ErrMissingField,
	"VALIDATION_ERROR": ErrValidation,

	// User (2xxx)
	"INVALID_EMAIL":    ErrInvalidEmail,
	"INVALID_PASSWORD": ErrInvalidPassword,
	"INVALID_NAME":     ErrInvalidName,
	"USER_NOT_FOUND":   ErrUserNotFound,
	"EMAIL_EXISTS":     ErrEmailExists,

	// Session / Auth (3xxx)
	"SESSION_NOT_FOUND": ErrSessionNotFound,
	"SESSION_EXPIRED":   ErrSessionExpired,
	"INVALID_TOKEN":     ErrInvalidToken,

	// Account (4xxx)
	"ACCOUNT_NOT_FOUND": ErrAccountNotFound,
	"ACCOUNT_FROZEN":    ErrAccountFrozen,
	"ACCOUNT_CLOSED":    ErrAccountClosed,
	"BALANCE_NOT_ZERO":  ErrBalanceNotZero,
	"INSUFFICIENT_FUNDS": ErrInsufficientFunds,

	// Transfer (5xxx)
	"TRANSFER_NOT_FOUND": ErrTransferNotFound,
	"SAME_ACCOUNT":       ErrSameAccount,
	"INVALID_AMOUNT":     ErrInvalidAmount,
	"TRANSFER_FAILED":    ErrTransferFailed,
	"CURRENCY_MISMATCH":  ErrCurrencyMismatch,

	// Card (6xxx)
	"CARD_NOT_FOUND":        ErrCardNotFound,
	"CARD_BLOCKED":          ErrCardBlocked,
	"CARD_EXPIRED":          ErrCardExpired,
	"INVALID_PIN":           ErrInvalidPIN,
	"PIN_ATTEMPTS_EXCEEDED": ErrPINAttemptsExceeded,
	"CARD_VALIDATION":       ErrValidation,

	// KYC / AML / Fraud (7xxx)
	"KYC_REQUIRED":   ErrKYCRequired,
	"KYC_PENDING":    ErrKYCPending,
	"AML_BLOCKED":    ErrAMLBlocked,
	"FRAUD_DETECTED": ErrFraudDetected,

	// Beneficiary (8xxx)
	"BENEFICIARY_NOT_FOUND": ErrBeneficiaryNotFound,
	"BENEFICIARY_EXISTS":    ErrBeneficiaryExists,

	// Contact (86xx)
	"CONTACT_NOT_FOUND": ErrContactNotFound,
	"CONTACT_EXISTS":    ErrContactExists,
	"CONTACT_SELF":      ErrContactSelf,

	// Exchange (45xx)
	"RATE_NOT_FOUND": ErrRateNotFound,
}

// domainCoder is satisfied by domain.DomainError without importing it.
type domainCoder interface {
	Code() string
}

// MapDomainError converts a DomainError to the corresponding AppError.
// If the error is not a DomainError or the code is unknown, returns ErrInternal.
func MapDomainError(err error) *AppError {
	var dc domainCoder
	if errors.As(err, &dc) {
		if appErr, ok := domainCodeToAppError[dc.Code()]; ok {
			return appErr.Wrap(err)
		}
	}
	return ErrInternal.Wrap(err)
}

// IsDomainError checks whether the error is a DomainError (has a Code() method).
func IsDomainError(err error) bool {
	var dc domainCoder
	return errors.As(err, &dc)
}
