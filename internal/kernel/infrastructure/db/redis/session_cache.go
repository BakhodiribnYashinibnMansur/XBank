package redis

import (
	"context"
	"fmt"
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

// AddUserSession tracks an active session for a user.
// Uses a sorted set with session creation time as score.
func (s *SessionCache) AddUserSession(ctx context.Context, userID, sessionID string, ttl time.Duration) error {
	key := fmt.Sprintf("user_sessions:%s", userID)
	now := float64(time.Now().Unix())
	if err := s.client.ZAdd(ctx, key, redis.Z{Score: now, Member: sessionID}).Err(); err != nil {
		return err
	}
	return s.client.Expire(ctx, key, ttl).Err()
}

// CountUserSessions returns the number of active sessions for a user.
func (s *SessionCache) CountUserSessions(ctx context.Context, userID string) (int64, error) {
	key := fmt.Sprintf("user_sessions:%s", userID)
	return s.client.ZCard(ctx, key).Result()
}

// RemoveUserSession removes a specific session from the user's active set.
func (s *SessionCache) RemoveUserSession(ctx context.Context, userID, sessionID string) error {
	key := fmt.Sprintf("user_sessions:%s", userID)
	return s.client.ZRem(ctx, key, sessionID).Err()
}

// RemoveOldestUserSession removes the oldest session if the user exceeds maxSessions.
// Returns the removed session ID, or empty string if no eviction was needed.
func (s *SessionCache) RemoveOldestUserSession(ctx context.Context, userID string, maxSessions int64) (string, error) {
	key := fmt.Sprintf("user_sessions:%s", userID)

	count, err := s.client.ZCard(ctx, key).Result()
	if err != nil {
		return "", err
	}

	if count < maxSessions {
		return "", nil
	}

	// Get the oldest session (lowest score = earliest creation time)
	members, err := s.client.ZRange(ctx, key, 0, 0).Result()
	if err != nil || len(members) == 0 {
		return "", err
	}

	oldest := members[0]
	if err := s.client.ZRem(ctx, key, oldest).Err(); err != nil {
		return "", err
	}

	return oldest, nil
}
