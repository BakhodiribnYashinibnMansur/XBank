package kyc

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/kyc"
)

type Service struct {
	repo kyc.Repository
}

func NewService(repo kyc.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Submit(ctx context.Context, userID string, docType kyc.DocType, docNumber, firstName, lastName, dob string) (*kyc.Verification, error) {
	v, err := kyc.NewVerification(userID, docType, docNumber, firstName, lastName, dob)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Service) GetStatus(ctx context.Context, userID string) (*kyc.Verification, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) Approve(ctx context.Context, verificationID, reviewerID string) error {
	v, err := s.repo.GetByID(ctx, verificationID)
	if err != nil {
		return err
	}
	v.Approve(reviewerID)
	return s.repo.Update(ctx, v)
}

func (s *Service) Reject(ctx context.Context, verificationID, reviewerID, reason string) error {
	v, err := s.repo.GetByID(ctx, verificationID)
	if err != nil {
		return err
	}
	v.Reject(reviewerID, reason)
	return s.repo.Update(ctx, v)
}

func (s *Service) ListPending(ctx context.Context, limit, offset int) ([]*kyc.Verification, int64, error) {
	items, err := s.repo.ListPending(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountPending(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
