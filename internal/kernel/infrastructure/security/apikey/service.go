package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	rawPrefix  = "xbk_live_"
	prefixLen  = 8 // characters after rawPrefix used for lookup
	cacheKey   = "apikey:valid:"
	cacheTTL   = 5 * time.Minute
)

// Service manages API key lifecycle: generation, validation, revocation.
type Service struct {
	store Store
	cache *goredis.Client // optional: nil to disable caching
}

// NewService creates an API key service with a persistence store and optional Redis cache.
func NewService(store Store, cache *goredis.Client) *Service {
	return &Service{store: store, cache: cache}
}

// Generate creates a new API key. The raw key is returned only once and never stored.
func (s *Service) Generate(ctx context.Context, userID string, scopes []string, ttl *time.Duration) (string, *Key, error) {
	if userID == "" {
		return "", nil, fmt.Errorf("apikey: user_id is required")
	}

	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", nil, fmt.Errorf("apikey: generate random: %w", err)
	}

	rawHex := hex.EncodeToString(rawBytes)
	rawKey := rawPrefix + rawHex
	prefix := rawHex[:prefixLen]
	hash := hashKey(rawKey)

	var expiresAt *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresAt = &t
	}

	key := &Key{
		Prefix:    prefix,
		Hash:      hash,
		UserID:    userID,
		Scopes:    scopes,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.store.Save(ctx, key); err != nil {
		return "", nil, fmt.Errorf("apikey: save: %w", err)
	}

	return rawKey, key, nil
}

// Validate checks a raw API key against the store.
func (s *Service) Validate(ctx context.Context, rawKey string) (*Key, error) {
	if len(rawKey) < len(rawPrefix)+prefixLen {
		return nil, fmt.Errorf("apikey: invalid key format")
	}

	prefix := rawKey[len(rawPrefix) : len(rawPrefix)+prefixLen]

	key, err := s.store.GetByPrefix(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("apikey: not found")
	}

	if key.Revoked {
		return nil, fmt.Errorf("apikey: key has been revoked")
	}
	if key.IsExpired() {
		return nil, fmt.Errorf("apikey: key has expired")
	}

	// Constant-time hash comparison
	expectedHash := hashKey(rawKey)
	if subtle.ConstantTimeCompare([]byte(key.Hash), []byte(expectedHash)) != 1 {
		return nil, fmt.Errorf("apikey: invalid key")
	}

	return key, nil
}

// Revoke marks an API key as revoked.
func (s *Service) Revoke(ctx context.Context, keyID string) error {
	return s.store.Revoke(ctx, keyID)
}

func hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
