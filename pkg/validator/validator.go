// Package validator provides struct validation with custom rules.
package validator

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+[0-9]{8,15}$`)
)

// ValidationError holds field-level validation errors.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of field errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	msgs := make([]string, len(e))
	for i, v := range e {
		msgs[i] = v.Error()
	}
	return strings.Join(msgs, "; ")
}

// ValidateEmail checks if the email format is valid.
func ValidateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}
	if !emailRegex.MatchString(email) {
		return &ValidationError{Field: "email", Message: "invalid email format"}
	}
	return nil
}

// ValidatePhone checks if the phone format is valid (international format).
func ValidatePhone(phone string) error {
	if phone == "" {
		return &ValidationError{Field: "phone", Message: "phone is required"}
	}
	if !phoneRegex.MatchString(phone) {
		return &ValidationError{Field: "phone", Message: "phone must start with + and contain 8-15 digits"}
	}
	return nil
}

// ValidatePassword checks password strength requirements.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	return nil
}

// ValidateRequired checks that a field is not empty.
func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}
	return nil
}
