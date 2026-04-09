package command

import (
	"context"
	"time"

	metric "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/metric/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

type Service struct {
	repo metric.Repository
}

func NewService(repo metric.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Collect(ctx context.Context, name string, value float64, labels map[string]string) (_ *metric.AppMetric, err error) {
	defer metrics.ObserveService("MetricService", "Collect", time.Now(), &err)

	m, err := metric.NewAppMetric(name, value, labels)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) FindByName(ctx context.Context, name string) (_ []*metric.AppMetric, err error) {
	defer metrics.ObserveService("MetricService", "FindByName", time.Now(), &err)
	return s.repo.FindByName(ctx, name)
}

func (s *Service) ListRecent(ctx context.Context, limit int) (_ []*metric.AppMetric, err error) {
	defer metrics.ObserveService("MetricService", "ListRecent", time.Now(), &err)
	if limit <= 0 {
		limit = 100
	}
	return s.repo.ListRecent(ctx, limit)
}
