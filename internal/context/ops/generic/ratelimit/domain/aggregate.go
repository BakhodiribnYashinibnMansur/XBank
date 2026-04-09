package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

var (
	ErrRateLimitNotFound = domain.NewDomainError("RATE_LIMIT_NOT_FOUND", "rate limit rule not found")
	ErrRateLimitExists   = domain.NewDomainError("RATE_LIMIT_EXISTS", "rate limit rule with this key already exists")
)

// RateLimitRule defines a configurable rate limiting rule.
type RateLimitRule struct {
	ID            string
	Key           string // unique key (e.g. endpoint path)
	MaxRequests   int
	WindowSeconds int
	Description   string
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewRateLimitRule creates a new rate limit rule.
func NewRateLimitRule(key string, maxRequests, windowSeconds int, description string, enabled bool) (*RateLimitRule, error) {
	if key == "" {
		return nil, domain.NewDomainError("MISSING_FIELD", "key is required")
	}
	if maxRequests <= 0 {
		return nil, domain.NewDomainError("INVALID_FIELD", "max_requests must be positive")
	}
	if windowSeconds <= 0 {
		return nil, domain.NewDomainError("INVALID_FIELD", "window_seconds must be positive")
	}

	now := time.Now()
	return &RateLimitRule{
		Key:           key,
		MaxRequests:   maxRequests,
		WindowSeconds: windowSeconds,
		Description:   description,
		Enabled:       enabled,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Update modifies the rate limit rule fields.
func (r *RateLimitRule) Update(maxRequests, windowSeconds int, description string, enabled bool) {
	r.MaxRequests = maxRequests
	r.WindowSeconds = windowSeconds
	r.Description = description
	r.Enabled = enabled
	r.UpdatedAt = time.Now()
}

// Repository defines the persistence contract for rate limit rules.
type Repository interface {
	Save(ctx context.Context, rule *RateLimitRule) error
	FindByID(ctx context.Context, id string) (*RateLimitRule, error)
	FindByKey(ctx context.Context, key string) (*RateLimitRule, error)
	FindAll(ctx context.Context) ([]*RateLimitRule, error)
	Update(ctx context.Context, rule *RateLimitRule) error
	Delete(ctx context.Context, id string) error
}
