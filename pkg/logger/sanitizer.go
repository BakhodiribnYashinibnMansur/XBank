package logger

import "strings"

// sensitive field names that must never appear in logs
var sensitiveFields = map[string]bool{
	"password":      true,
	"pin":           true,
	"cvv":           true,
	"pan":           true,
	"card_number":   true,
	"refresh_token": true,
	"access_token":  true,
	"token":         true,
	"secret":        true,
	"authorization": true,
	"cookie":        true,
}

// SanitizeField - check if a field name is sensitive
func IsSensitiveField(field string) bool {
	return sensitiveFields[strings.ToLower(field)]
}

// Redact - mask a sensitive value
func Redact(value string) string {
	if len(value) <= 4 {
		return "[REDACTED]"
	}
	return value[:2] + "***" + value[len(value)-2:]
}

// RedactPAN - mask card number: **** **** **** 1234
func RedactPAN(pan string) string {
	if len(pan) < 4 {
		return "[REDACTED]"
	}
	return "**** **** **** " + pan[len(pan)-4:]
}

// RedactBearer - mask Authorization header: Bearer [REDACTED]
func RedactBearer(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return "Bearer [REDACTED]"
	}
	return "[REDACTED]"
}
