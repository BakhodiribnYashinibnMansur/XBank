package revocation

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const keyPrefix = "jwt:revoked:"

// Store provides JWT token revocation via Redis.
// IsRevoked fails open on Redis errors (availability > security for reads).
type Store struct {
	client *goredis.Client
}

// NewStore creates a revocation store backed by Redis.
func NewStore(client *goredis.Client) *Store {
	return &Store{client: client}
}

// Revoke marks a JWT (by JTI) as revoked for the given remaining TTL.
func (s *Store) Revoke(ctx context.Context, jti string, remainingTTL time.Duration) error {
	return s.client.Set(ctx, keyPrefix+jti, "1", remainingTTL).Err()
}

// RevokeAll marks multiple JTIs as revoked in a single pipeline.
func (s *Store) RevokeAll(ctx context.Context, jtis []string, ttl time.Duration) error {
	if len(jtis) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, jti := range jtis {
		pipe.Set(ctx, keyPrefix+jti, "1", ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// IsRevoked checks whether a JTI has been revoked.
// Returns false on Redis errors (fail-open).
func (s *Store) IsRevoked(ctx context.Context, jti string) bool {
	exists, err := s.client.Exists(ctx, keyPrefix+jti).Result()
	if err != nil {
		logger.Log.Warn("revocation: redis check failed, failing open",
			zap.String("jti", jti),
			zap.Error(err),
		)
		return false
	}
	return exists > 0
}
