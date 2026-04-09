package command

import (
	"context"
	"time"

	dashboard "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/dashboard/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

// Service provides dashboard analytics operations.
type Service struct {
	repo dashboard.Repository
}

// NewService creates a new dashboard service.
func NewService(repo dashboard.Repository) *Service {
	return &Service{repo: repo}
}

// GetOverview returns aggregated system-wide statistics.
func (s *Service) GetOverview(ctx context.Context) (_ *dashboard.OverviewStats, err error) {
	defer metrics.ObserveService("DashboardService", "GetOverview", time.Now(), &err)
	return s.repo.GetOverview(ctx)
}

// GetPeriodStats returns statistics for a specific time period.
func (s *Service) GetPeriodStats(ctx context.Context, period string, start, end time.Time) (_ *dashboard.PeriodStats, err error) {
	defer metrics.ObserveService("DashboardService", "GetPeriodStats", time.Now(), &err)
	return s.repo.GetPeriodStats(ctx, period, start, end)
}
