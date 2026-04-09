package command

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	healthcheck "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

// Service provides health check operations.
type Service struct {
	repo     healthcheck.Repository
	checkers []healthcheck.HealthChecker
}

// NewService creates a new healthcheck service.
func NewService(repo healthcheck.Repository, checkers ...healthcheck.HealthChecker) *Service {
	return &Service{
		repo:     repo,
		checkers: checkers,
	}
}

// RunCheck executes all health checkers and persists the result.
func (s *Service) RunCheck(ctx context.Context) (_ *healthcheck.SystemHealth, err error) {
	defer metrics.ObserveService("HealthcheckService", "RunCheck", time.Now(), &err)

	var checks []healthcheck.ComponentCheck
	for _, checker := range s.checkers {
		checks = append(checks, checker.Check(ctx))
	}

	health := healthcheck.NewSystemHealth(checks)

	componentsJSON, err := json.Marshal(health.Components)
	if err != nil {
		return nil, fmt.Errorf("healthcheck: marshal components: %w", err)
	}

	record := &healthcheck.HealthRecord{
		Status:     health.Status,
		Components: string(componentsJSON),
		CheckedAt:  health.CheckedAt,
	}

	if err := s.repo.Save(ctx, record); err != nil {
		return nil, fmt.Errorf("healthcheck: save record: %w", err)
	}

	return health, nil
}

// GetLatest returns the most recent health check result.
func (s *Service) GetLatest(ctx context.Context) (_ *healthcheck.HealthRecord, err error) {
	defer metrics.ObserveService("HealthcheckService", "GetLatest", time.Now(), &err)
	return s.repo.GetLatest(ctx)
}

// ListHistory returns historical health check records with pagination.
func (s *Service) ListHistory(ctx context.Context, limit, offset int) (_ []*healthcheck.HealthRecord, _ int64, err error) {
	defer metrics.ObserveService("HealthcheckService", "ListHistory", time.Now(), &err)

	records, err := s.repo.ListHistory(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountHistory(ctx)
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
