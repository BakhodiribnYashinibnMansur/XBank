package csrf

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"
)

// Service provides CSRF double-submit cookie token generation and validation.
// Tokens are HMAC-SHA256(secret, sessionID + timestamp), base64url encoded.
type Service struct {
	secret   []byte
	tokenTTL time.Duration
}

// NewService creates a CSRF service with the given secret and token TTL.
// Secret must be at least 32 bytes.
func NewService(secret []byte, tokenTTL time.Duration) (*Service, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("csrf: secret must be at least 32 bytes, got %d", len(secret))
	}
	return &Service{
		secret:   secret,
		tokenTTL: tokenTTL,
	}, nil
}

// GenerateToken creates a CSRF token bound to the given session ID.
func (s *Service) GenerateToken(sessionID string) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("csrf: session ID cannot be empty")
	}

	now := time.Now().Unix()
	payload := buildCSRFPayload(sessionID, now)

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	sig := mac.Sum(nil)

	// token = base64url(timestamp_bytes + signature)
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(now))

	token := make([]byte, 8+len(sig))
	copy(token, ts)
	copy(token[8:], sig)

	return base64.URLEncoding.EncodeToString(token), nil
}

// ValidateToken verifies a CSRF token against the session ID.
func (s *Service) ValidateToken(sessionID, token string) error {
	if sessionID == "" || token == "" {
		return fmt.Errorf("csrf: session ID and token are required")
	}

	raw, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return fmt.Errorf("csrf: invalid token encoding: %w", err)
	}

	if len(raw) < 8+sha256.Size {
		return fmt.Errorf("csrf: token too short")
	}

	// Extract timestamp
	ts := int64(binary.BigEndian.Uint64(raw[:8]))
	sig := raw[8:]

	// Check expiry
	elapsed := time.Since(time.Unix(ts, 0))
	if elapsed > s.tokenTTL || elapsed < 0 {
		return fmt.Errorf("csrf: token expired")
	}

	// Recompute and compare
	payload := buildCSRFPayload(sessionID, ts)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	expected := mac.Sum(nil)

	if subtle.ConstantTimeCompare(sig, expected) != 1 {
		return fmt.Errorf("csrf: token mismatch")
	}

	return nil
}

func buildCSRFPayload(sessionID string, timestamp int64) []byte {
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(timestamp))
	payload := make([]byte, len(ts)+len(sessionID))
	copy(payload, ts)
	copy(payload[len(ts):], sessionID)
	return payload
}
