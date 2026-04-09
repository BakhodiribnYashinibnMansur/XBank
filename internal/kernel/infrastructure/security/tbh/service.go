package tbh

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	keyPrefix   = "tbh:"
	tokenPrefix = "tbh_"
)

// Service manages token-based handshakes via Redis.
// Tokens are one-time-use: verification atomically deletes the token.
type Service struct {
	client *goredis.Client
	ttl    time.Duration
}

// NewService creates a TBH service with the given Redis client and token TTL.
func NewService(client *goredis.Client, ttl time.Duration) *Service {
	return &Service{client: client, ttl: ttl}
}

// Issue creates a new handshake token bound to a user and purpose.
func (s *Service) Issue(ctx context.Context, userID, purpose, payload string) (*Handshake, error) {
	if userID == "" || purpose == "" {
		return nil, fmt.Errorf("tbh: user_id and purpose are required")
	}

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("tbh: generate token: %w", err)
	}

	h := &Handshake{
		Token:     token,
		UserID:    userID,
		Purpose:   purpose,
		Payload:   payload,
		ExpiresAt: time.Now().Add(s.ttl),
	}

	data, err := json.Marshal(h)
	if err != nil {
		return nil, fmt.Errorf("tbh: marshal handshake: %w", err)
	}

	key := redisKey(token)
	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return nil, fmt.Errorf("tbh: store token: %w", err)
	}

	return h, nil
}

// Verify validates and consumes a handshake token (one-time use).
// The token is deleted atomically on successful verification.
func (s *Service) Verify(ctx context.Context, token string) (*Handshake, error) {
	if token == "" {
		return nil, fmt.Errorf("tbh: token is required")
	}

	key := redisKey(token)

	// GETDEL: atomic get-and-delete (Redis 6.2+)
	data, err := s.client.GetDel(ctx, key).Bytes()
	if err == goredis.Nil {
		return nil, fmt.Errorf("tbh: token not found or already consumed")
	}
	if err != nil {
		return nil, fmt.Errorf("tbh: verify token: %w", err)
	}

	var h Handshake
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("tbh: unmarshal handshake: %w", err)
	}

	if time.Now().After(h.ExpiresAt) {
		return nil, fmt.Errorf("tbh: token expired")
	}

	return &h, nil
}

// Revoke explicitly removes a handshake token.
func (s *Service) Revoke(ctx context.Context, token string) error {
	return s.client.Del(ctx, redisKey(token)).Err()
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return tokenPrefix + hex.EncodeToString(b), nil
}

func redisKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return keyPrefix + hex.EncodeToString(hash[:])
}
