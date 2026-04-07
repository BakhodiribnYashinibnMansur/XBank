package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/beneficiary"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
)

type Service struct {
	repo beneficiary.Repository
}

func NewService(repo beneficiary.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, userID, name, accountNumber, bankName, bankCode, currency string, benType beneficiary.Type) (_ *beneficiary.Beneficiary, err error) {
	defer metrics.ObserveService("BeneficiaryService", "Add", time.Now(), &err)

	exists, err := s.repo.ExistsByUserAndAccount(ctx, userID, accountNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, beneficiary.ErrBeneficiaryExists
	}

	b, err := beneficiary.NewBeneficiary(userID, name, accountNumber, bankName, bankCode, currency, benType)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (_ *beneficiary.Beneficiary, err error) {
	defer metrics.ObserveService("BeneficiaryService", "GetByID", time.Now(), &err)
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByUserID(ctx context.Context, userID string, limit, offset int) (_ []*beneficiary.Beneficiary, _ int64, err error) {
	defer metrics.ObserveService("BeneficiaryService", "ListByUserID", time.Now(), &err)

	items, err := s.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) Delete(ctx context.Context, id string) (err error) {
	defer metrics.ObserveService("BeneficiaryService", "Delete", time.Now(), &err)
	return s.repo.Delete(ctx, id)
}
