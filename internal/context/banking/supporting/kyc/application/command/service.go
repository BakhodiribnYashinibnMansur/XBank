package command

import (
	"context"
	"time"

	kyc "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

type Service struct {
	repo kyc.Repository
}

func NewService(repo kyc.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Submit(ctx context.Context, userID string, docType kyc.DocType, docNumber, firstName, lastName, dob string) (_ *kyc.Verification, err error) {
	defer metrics.ObserveService("KYCService", "Submit", time.Now(), &err)

	v, err := kyc.NewVerification(userID, docType, docNumber, firstName, lastName, dob)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Service) GetStatus(ctx context.Context, userID string) (_ *kyc.Verification, err error) {
	defer metrics.ObserveService("KYCService", "GetStatus", time.Now(), &err)
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) Approve(ctx context.Context, verificationID, reviewerID string) (err error) {
	defer metrics.ObserveService("KYCService", "Approve", time.Now(), &err)

	v, err := s.repo.GetByID(ctx, verificationID)
	if err != nil {
		return err
	}
	v.Approve(reviewerID)
	return s.repo.Update(ctx, v)
}

func (s *Service) Reject(ctx context.Context, verificationID, reviewerID, reason string) (err error) {
	defer metrics.ObserveService("KYCService", "Reject", time.Now(), &err)

	v, err := s.repo.GetByID(ctx, verificationID)
	if err != nil {
		return err
	}
	v.Reject(reviewerID, reason)
	return s.repo.Update(ctx, v)
}

func (s *Service) ListPending(ctx context.Context, limit, offset int) (_ []*kyc.Verification, _ int64, err error) {
	defer metrics.ObserveService("KYCService", "ListPending", time.Now(), &err)

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
