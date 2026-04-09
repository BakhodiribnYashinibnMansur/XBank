package audit

import "time"

// EventType identifies the kind of security event.
type EventType string

const (
	EventLoginSuccess      EventType = "login_success"
	EventLoginFailed       EventType = "login_failed"
	EventLogout            EventType = "logout"
	EventTokenRevoked      EventType = "token_revoked"
	EventPermissionChanged EventType = "permission_changed"
	EventAPIKeyCreated     EventType = "apikey_created"
	EventAPIKeyRevoked     EventType = "apikey_revoked"
	EventSuspiciousAccess  EventType = "suspicious_access"
	EventDeviceEnrolled    EventType = "device_enrolled"
	EventPasswordChanged   EventType = "password_changed"
)

// Event represents a security audit event.
type Event struct {
	ID        string            `json:"id"`
	Type      EventType         `json:"type"`
	UserID    string            `json:"user_id"`
	IPAddress string            `json:"ip_address"`
	UserAgent string            `json:"user_agent"`
	Metadata  map[string]string `json:"metadata"`
	Timestamp time.Time         `json:"timestamp"`
}
