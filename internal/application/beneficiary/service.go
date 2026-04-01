package beneficiary

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/beneficiary"
)

type Service struct {
	repo beneficiary.Repository
}

func NewService(repo beneficiary.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, userID, name, accountNumber, bankName, bankCode, currency string, benType beneficiary.Type) (*beneficiary.Beneficiary, error) {
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

func (s *Service) GetByID(ctx context.Context, id string) (*beneficiary.Beneficiary, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*beneficiary.Beneficiary, int64, error) {
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

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
