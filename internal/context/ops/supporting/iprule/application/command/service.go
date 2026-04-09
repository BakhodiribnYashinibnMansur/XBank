package command

import (
	"context"
	"time"

	iprule "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

type Service struct {
	repo iprule.Repository
}

func NewService(repo iprule.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, ipAddress string, ruleType iprule.RuleType, reason, createdBy string, expiresAt *time.Time) (_ *iprule.IPRule, err error) {
	defer metrics.ObserveService("IPRuleService", "Create", time.Now(), &err)

	existing, _ := s.repo.FindByIP(ctx, ipAddress)
	if existing != nil {
		return nil, iprule.ErrIPRuleExists
	}

	rule, err := iprule.NewIPRule(ipAddress, ruleType, reason, createdBy, expiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (_ *iprule.IPRule, err error) {
	defer metrics.ObserveService("IPRuleService", "GetByID", time.Now(), &err)
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListAll(ctx context.Context) (_ []*iprule.IPRule, err error) {
	defer metrics.ObserveService("IPRuleService", "ListAll", time.Now(), &err)
	return s.repo.ListAll(ctx)
}

func (s *Service) Delete(ctx context.Context, id string) (err error) {
	defer metrics.ObserveService("IPRuleService", "Delete", time.Now(), &err)
	return s.repo.Delete(ctx, id)
}
