package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"time"
)

// HMACSigner - HMAC-SHA256 request signing and verification.
//
// Used to protect sensitive endpoints (transfers, cards) from
// request body tampering (man-in-the-middle attacks).
//
// Client signs: HMAC-SHA256(secret, timestamp + "." + body)
// Server verifies: recompute HMAC and constant-time compare.
type HMACSigner struct {
	secret       []byte
	maxClockSkew time.Duration // max allowed time difference
}

// NewHMACSigner - create signer from hex-encoded secret key.
// Key must be at least 32 bytes (64 hex chars).
// maxClockSkew limits how old a signed request can be (prevents replay).
func NewHMACSigner(hexSecret string, maxClockSkew time.Duration) (*HMACSigner, error) {
	secret, err := hex.DecodeString(hexSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid hex secret: %w", err)
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("secret must be at least 32 bytes (64 hex chars), got %d bytes", len(secret))
	}

	return &HMACSigner{
		secret:       secret,
		maxClockSkew: maxClockSkew,
	}, nil
}

// Sign - compute HMAC-SHA256 over timestamp + "." + body.
// Returns hex-encoded signature.
func (s *HMACSigner) Sign(timestamp int64, body []byte) string {
	payload := buildPayload(timestamp, body)

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify - verify signature against timestamp + body.
//
// Returns nil if valid, error describing the failure otherwise.
// Uses constant-time comparison to prevent timing attacks.
func (s *HMACSigner) Verify(signature string, timestamp int64, body []byte) error {
	// 1. Check timestamp freshness (anti-replay)
	now := time.Now().Unix()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}
	maxSkewSeconds := int64(math.Ceil(s.maxClockSkew.Seconds()))
	if diff > maxSkewSeconds {
		return fmt.Errorf("request timestamp expired: %d seconds skew (max %v)", diff, s.maxClockSkew)
	}

	// 2. Decode provided signature
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}

	// 3. Recompute expected HMAC
	payload := buildPayload(timestamp, body)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	expected := mac.Sum(nil)

	// 4. Constant-time comparison (prevents timing attacks)
	if subtle.ConstantTimeCompare(sigBytes, expected) != 1 {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// buildPayload - constructs "timestamp.body" for signing.
func buildPayload(timestamp int64, body []byte) []byte {
	ts := strconv.FormatInt(timestamp, 10)
	payload := make([]byte, len(ts)+1+len(body))
	copy(payload, ts)
	payload[len(ts)] = '.'
	copy(payload[len(ts)+1:], body)
	return payload
}

// GenerateHMACSecret - generate a random 32-byte HMAC secret (hex encoded).
func GenerateHMACSecret() (string, error) {
	return GenerateKey() // reuse AES key generation (same: 32 random bytes)
}
