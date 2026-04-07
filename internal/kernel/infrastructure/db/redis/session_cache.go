package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionCache - Redis-based session/token management
type SessionCache struct {
	client *redis.Client
}

func NewSessionCache(client *redis.Client) *SessionCache {
	return &SessionCache{client: client}
}

// BlacklistToken - add a JWT to the blacklist (logout)
// TTL = remaining token lifetime
func (s *SessionCache) BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	return s.client.Set(ctx, "blacklist:"+tokenID, "1", ttl).Err()
}

// IsBlacklisted - check if a JWT has been revoked
func (s *SessionCache) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	exists, err := s.client.Exists(ctx, "blacklist:"+tokenID).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// SetSession - cache session data
func (s *SessionCache) SetSession(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.client.Set(ctx, "session:"+key, value, ttl).Err()
}

// GetSession - get cached session
func (s *SessionCache) GetSession(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, "session:"+key).Result()
}

// DeleteSession - remove session
func (s *SessionCache) DeleteSession(ctx context.Context, key string) error {
	return s.client.Del(ctx, "session:"+key).Err()
}
