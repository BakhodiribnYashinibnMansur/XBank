package domain

import "fmt"

// DomainError represents a domain-level business rule violation.
// The code field acts as a machine-readable discriminator (e.g. "USER_NOT_FOUND")
// that the presentation layer maps to HTTP status codes and AppError instances.
// Use errors.Is for comparison — it matches on code, not message.
type DomainError struct {
	code    string
	message string
}

// NewDomainError creates a new DomainError with the given code and message.
func NewDomainError(code, message string) *DomainError {
	return &DomainError{code: code, message: message}
}

// Error returns the formatted error string.
func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// Code returns the machine-readable error code.
func (e *DomainError) Code() string {
	return e.code
}

// Message returns the human-readable error message.
func (e *DomainError) Message() string {
	return e.message
}

// Is checks if the target error is a DomainError with the same code,
// enabling errors.Is semantics for sentinel domain errors.
func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.code == t.code
}
