package tbh

import "time"

// Handshake represents a short-lived token for device enrollment or key exchange.
type Handshake struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Purpose   string    `json:"purpose"`
	Payload   string    `json:"payload,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}
