package redact

import "strings"

// sensitiveFields contains field names that must never appear as plaintext in logs.
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
	"otp":           true,
	"private_key":   true,
}

// IsSensitive returns true if the field name is in the sensitive list.
func IsSensitive(field string) bool {
	return sensitiveFields[strings.ToLower(field)]
}

// Value masks a sensitive value, preserving only the first and last two characters.
// Short values (<=4 chars) are fully redacted.
func Value(v string) string {
	if len(v) <= 4 {
		return "[REDACTED]"
	}
	return v[:2] + "***" + v[len(v)-2:]
}

// PAN masks a card number in the format **** **** **** 1234.
func PAN(pan string) string {
	if len(pan) < 4 {
		return "[REDACTED]"
	}
	return "**** **** **** " + pan[len(pan)-4:]
}

// Bearer masks an Authorization header value.
func Bearer(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return "Bearer [REDACTED]"
	}
	return "[REDACTED]"
}

// E2EE returns the standard placeholder for E2EE-encrypted fields in logs.
func E2EE() string {
	return "[E2EE_REDACTED]"
}

// Email masks an email address: ab***@***.com
func Email(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 1 {
		return "[REDACTED]"
	}
	local := email[:at]
	domain := email[at+1:]

	maskedLocal := local[:1] + "***"
	dot := strings.LastIndex(domain, ".")
	if dot < 1 {
		return maskedLocal + "@***"
	}
	return maskedLocal + "@***" + domain[dot:]
}

// Phone masks a phone number, keeping country code and last 2 digits: +998***12
func Phone(phone string) string {
	if len(phone) < 6 {
		return "[REDACTED]"
	}
	return phone[:4] + "***" + phone[len(phone)-2:]
}

// Map redacts all sensitive fields in a map by key name.
func Map(data map[string]string) map[string]string {
	result := make(map[string]string, len(data))
	for k, v := range data {
		if IsSensitive(k) {
			result[k] = "[REDACTED]"
		} else {
			result[k] = v
		}
	}
	return result
}
