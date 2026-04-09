package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^\+[0-9]{8,15}$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	currencyRegex = regexp.MustCompile(`^[A-Z]{3}$`)
)

// Error holds a single field-level validation error.
type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Errors is a collection of field validation errors.
type Errors []*Error

func (e Errors) Error() string {
	msgs := make([]string, len(e))
	for i, v := range e {
		msgs[i] = v.Error()
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are any validation errors.
func (e Errors) HasErrors() bool { return len(e) > 0 }

// Required checks that the value is not empty after trimming whitespace.
func Required(field, value string) *Error {
	if strings.TrimSpace(value) == "" {
		return &Error{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}
	return nil
}

// MinLength checks minimum string length.
func MinLength(field, value string, min int) *Error {
	if len(value) < min {
		return &Error{Field: field, Message: fmt.Sprintf("%s must be at least %d characters", field, min)}
	}
	return nil
}

// MaxLength checks maximum string length.
func MaxLength(field, value string, max int) *Error {
	if len(value) > max {
		return &Error{Field: field, Message: fmt.Sprintf("%s must be at most %d characters", field, max)}
	}
	return nil
}

// Email checks if the value is a valid email format.
func Email(field, value string) *Error {
	if !emailRegex.MatchString(value) {
		return &Error{Field: field, Message: "invalid email format"}
	}
	return nil
}

// Phone checks if the value is a valid international phone number.
func Phone(field, value string) *Error {
	if !phoneRegex.MatchString(value) {
		return &Error{Field: field, Message: "phone must start with + and contain 8-15 digits"}
	}
	return nil
}

// UUID checks if the value is a valid UUID v4 format.
func UUID(field, value string) *Error {
	if !uuidRegex.MatchString(strings.ToLower(value)) {
		return &Error{Field: field, Message: "invalid UUID format"}
	}
	return nil
}

// Currency checks if the value is a valid ISO 4217 currency code.
func Currency(field, value string) *Error {
	if !currencyRegex.MatchString(value) {
		return &Error{Field: field, Message: "currency must be a 3-letter uppercase code"}
	}
	return nil
}

// Password checks password strength: minimum 8 characters.
func Password(field, value string) *Error {
	if len(value) < 8 {
		return &Error{Field: field, Message: "password must be at least 8 characters"}
	}
	return nil
}

// StrongPassword checks password complexity:
// minimum 8 characters, at least 1 uppercase, 1 lowercase, 1 digit, 1 special character.
func StrongPassword(field, value string) *Error {
	if len(value) < 8 {
		return &Error{Field: field, Message: "password must be at least 8 characters"}
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return &Error{Field: field, Message: "password must contain at least one uppercase letter"}
	}
	if !hasLower {
		return &Error{Field: field, Message: "password must contain at least one lowercase letter"}
	}
	if !hasDigit {
		return &Error{Field: field, Message: "password must contain at least one digit"}
	}
	if !hasSpecial {
		return &Error{Field: field, Message: "password must contain at least one special character"}
	}
	return nil
}

// PositiveInt64 checks that a number is positive.
func PositiveInt64(field string, value int64) *Error {
	if value <= 0 {
		return &Error{Field: field, Message: fmt.Sprintf("%s must be positive", field)}
	}
	return nil
}

// InSlice checks that a value is one of the allowed values.
func InSlice[T comparable](field string, value T, allowed []T) *Error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return &Error{Field: field, Message: fmt.Sprintf("%s has an invalid value", field)}
}

// Collect gathers non-nil errors into a slice.
func Collect(errs ...*Error) Errors {
	var result Errors
	for _, e := range errs {
		if e != nil {
			result = append(result, e)
		}
	}
	return result
}
