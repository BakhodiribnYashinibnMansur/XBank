package cache

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/repository"
	localCache "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/cache"
)

// CachedEvaluator wraps WriteRepository with an LRU cache for flag lookups.
// Cache is invalidated via Pub/Sub on featureflag.updated events.
type CachedEvaluator struct {
	repo  repository.WriteRepository
	cache *localCache.LRU[string, *entity.FeatureFlag]
	ttl   time.Duration
}

// NewCachedEvaluator creates a cached evaluator with the given TTL.
func NewCachedEvaluator(repo repository.WriteRepository, capacity int, ttl time.Duration) *CachedEvaluator {
	return &CachedEvaluator{
		repo:  repo,
		cache: localCache.NewLRU[string, *entity.FeatureFlag](capacity),
		ttl:   ttl,
	}
}

// IsEnabled checks if a boolean flag is enabled for the given user.
func (e *CachedEvaluator) IsEnabled(ctx context.Context, key string, userID string, attributes map[string]string) bool {
	flag, err := e.getFlag(ctx, key)
	if err != nil || flag == nil {
		return false
	}

	if !flag.Active {
		return false
	}

	// Check rule groups first
	for _, rg := range flag.RuleGroups {
		if rg.Matches(attributes) {
			return true
		}
	}

	return flag.IsEnabledForUser(userID)
}

// GetValue returns the resolved value for a flag.
func (e *CachedEvaluator) GetValue(ctx context.Context, key string, userID string, attributes map[string]string) (string, error) {
	flag, err := e.getFlag(ctx, key)
	if err != nil {
		return "", err
	}
	if flag == nil {
		return "", entity.ErrFlagNotFound
	}

	if !flag.Active {
		return flag.DefaultValue, nil
	}

	// Check rule groups (sorted by priority)
	for _, rg := range flag.RuleGroups {
		if rg.Matches(attributes) {
			return rg.Value, nil
		}
	}

	return flag.DefaultValue, nil
}

// Invalidate removes a flag from cache (called on update events).
func (e *CachedEvaluator) Invalidate(key string) {
	e.cache.Remove(key)
}

// InvalidateAll clears the entire cache.
func (e *CachedEvaluator) InvalidateAll() {
	e.cache.Purge()
}

func (e *CachedEvaluator) getFlag(ctx context.Context, key string) (*entity.FeatureFlag, error) {
	if cached, ok := e.cache.Get(key); ok {
		return cached, nil
	}

	flag, err := e.repo.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	e.cache.Set(key, flag, e.ttl)
	return flag, nil
}
