package command

import (
	"context"
	"time"

	ratelimit "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

type Service struct {
	repo ratelimit.Repository
}

func NewService(repo ratelimit.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, key string, maxRequests, windowSeconds int, description string, enabled bool) (_ *ratelimit.RateLimitRule, err error) {
	defer metrics.ObserveService("RateLimitService", "Create", time.Now(), &err)

	existing, _ := s.repo.FindByKey(ctx, key)
	if existing != nil {
		return nil, ratelimit.ErrRateLimitExists
	}

	rule, err := ratelimit.NewRateLimitRule(key, maxRequests, windowSeconds, description, enabled)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (_ *ratelimit.RateLimitRule, err error) {
	defer metrics.ObserveService("RateLimitService", "GetByID", time.Now(), &err)
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListAll(ctx context.Context) (_ []*ratelimit.RateLimitRule, err error) {
	defer metrics.ObserveService("RateLimitService", "ListAll", time.Now(), &err)
	return s.repo.FindAll(ctx)
}

func (s *Service) Update(ctx context.Context, id string, maxRequests, windowSeconds int, description string, enabled bool) (_ *ratelimit.RateLimitRule, err error) {
	defer metrics.ObserveService("RateLimitService", "Update", time.Now(), &err)

	rule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rule.Update(maxRequests, windowSeconds, description, enabled)
	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Service) Delete(ctx context.Context, id string) (err error) {
	defer metrics.ObserveService("RateLimitService", "Delete", time.Now(), &err)
	return s.repo.Delete(ctx, id)
}
