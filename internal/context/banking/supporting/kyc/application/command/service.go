package command

import (
	"context"
	"time"

	kyc "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	commonpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/common"
	kycpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/kyc"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	repo      kyc.Repository
	publisher domain.EventPublisher
	topics    config.KafkaTopicsConfig
}

func NewService(repo kyc.Repository, publisher domain.EventPublisher, topics config.KafkaTopicsConfig) *Service {
	return &Service{repo: repo, publisher: publisher, topics: topics}
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

	s.publishKYCSubmitted(ctx, v)
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
	if err := s.repo.Update(ctx, v); err != nil {
		return err
	}

	s.publishKYCApproved(ctx, v)
	return nil
}

func (s *Service) Reject(ctx context.Context, verificationID, reviewerID, reason string) (err error) {
	defer metrics.ObserveService("KYCService", "Reject", time.Now(), &err)

	v, err := s.repo.GetByID(ctx, verificationID)
	if err != nil {
		return err
	}
	v.Reject(reviewerID, reason)
	if err := s.repo.Update(ctx, v); err != nil {
		return err
	}

	s.publishKYCRejected(ctx, v)
	return nil
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

// --- Kafka publish helpers ---

func newKYCMetadata(userID string) *commonpb.Metadata {
	return &commonpb.Metadata{
		EventId:   uuid.New().String(),
		UserId:    userID,
		Timestamp: timestamppb.Now(),
		Source:    "xbank-api",
	}
}

func (s *Service) publishKYCSubmitted(ctx context.Context, v *kyc.Verification) {
	if s.publisher == nil {
		return
	}
	msg := &kycpb.KYCSubmitted{
		Metadata:       newKYCMetadata(v.UserID),
		VerificationId: v.ID,
		UserId:         v.UserID,
		DocumentType:   string(v.DocumentType),
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.KYCSubmitted, v.UserID, data)
	}
}

func (s *Service) publishKYCApproved(ctx context.Context, v *kyc.Verification) {
	if s.publisher == nil {
		return
	}
	msg := &kycpb.KYCApproved{
		Metadata:       newKYCMetadata(v.UserID),
		VerificationId: v.ID,
		UserId:         v.UserID,
		Level:          string(v.Status),
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.KYCApproved, v.UserID, data)
	}
}

func (s *Service) publishKYCRejected(ctx context.Context, v *kyc.Verification) {
	if s.publisher == nil {
		return
	}
	msg := &kycpb.KYCRejected{
		Metadata:       newKYCMetadata(v.UserID),
		VerificationId: v.ID,
		UserId:         v.UserID,
		Reason:         v.RejectedReason,
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.KYCRejected, v.UserID, data)
	}
}
