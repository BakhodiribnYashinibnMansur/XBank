package command

import (
	"context"
	"time"

	integration "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/integration/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

type Service struct {
	repo integration.Repository
}

func NewService(repo integration.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, baseURL, apiKey string, status integration.Status, webhookURL string) (_ *integration.Integration, err error) {
	defer metrics.ObserveService("IntegrationService", "Create", time.Now(), &err)

	existing, _ := s.repo.FindByName(ctx, name)
	if existing != nil {
		return nil, integration.ErrIntegrationExists
	}

	i, err := integration.NewIntegration(name, baseURL, apiKey, status, webhookURL)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (_ *integration.Integration, err error) {
	defer metrics.ObserveService("IntegrationService", "GetByID", time.Now(), &err)
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListAll(ctx context.Context) (_ []*integration.Integration, err error) {
	defer metrics.ObserveService("IntegrationService", "ListAll", time.Now(), &err)
	return s.repo.ListAll(ctx)
}

func (s *Service) Update(ctx context.Context, id, baseURL, apiKey string, status integration.Status, webhookURL string) (_ *integration.Integration, err error) {
	defer metrics.ObserveService("IntegrationService", "Update", time.Now(), &err)

	i, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	i.Update(baseURL, apiKey, status, webhookURL)
	if err := s.repo.Update(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *Service) Delete(ctx context.Context, id string) (err error) {
	defer metrics.ObserveService("IntegrationService", "Delete", time.Now(), &err)
	return s.repo.Delete(ctx, id)
}
