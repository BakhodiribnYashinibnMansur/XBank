package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const revokedPrefix = "revoked:"

// RevocationStore manages a Redis-backed JWT session denylist.
// When a user logs out or a session is revoked, the session ID is added
// to the denylist with TTL matching the token's remaining lifetime.
// On Redis failure, the store fails open (allows access) to prevent
// total lockout during Redis outages.
type RevocationStore struct {
	client *goredis.Client
}

// NewRevocationStore creates a token revocation store.
func NewRevocationStore(client *goredis.Client) *RevocationStore {
	return &RevocationStore{client: client}
}

// Revoke adds a session ID to the denylist.
// TTL should match the remaining lifetime of the access token.
func (s *RevocationStore) Revoke(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := revokedPrefix + sessionID
	err := s.client.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		logger.Log.Error("revocation store: revoke failed",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return fmt.Errorf("revocation store: revoke: %w", err)
	}
	return nil
}

// RevokeMany adds multiple session IDs to the denylist in a single pipeline.
func (s *RevocationStore) RevokeMany(ctx context.Context, sessionIDs []string, ttl time.Duration) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, id := range sessionIDs {
		pipe.Set(ctx, revokedPrefix+id, "1", ttl)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Log.Error("revocation store: revoke many failed",
			zap.Int("count", len(sessionIDs)),
			zap.Error(err),
		)
		return fmt.Errorf("revocation store: revoke many: %w", err)
	}
	return nil
}

// IsRevoked checks if a session ID is in the denylist.
// Fails open on Redis error (returns false) to prevent total lockout.
func (s *RevocationStore) IsRevoked(ctx context.Context, sessionID string) bool {
	key := revokedPrefix + sessionID
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		// Fail open: if Redis is down, allow access rather than block everyone
		logger.Log.Warn("revocation store: check failed, failing open",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return false
	}
	return exists > 0
}
