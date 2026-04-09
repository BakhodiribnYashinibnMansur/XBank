package command

import (
	"context"
	"time"

	card "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/crypto"
	cardpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/cards"
	commonpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/common"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	repo      card.Repository
	hasher    card.PINHasher
	encryptor *crypto.AESEncryptor // nil = no encryption (dev mode)
	publisher domain.EventPublisher
	topics    config.KafkaTopicsConfig
}

func NewService(
	repo card.Repository,
	hasher card.PINHasher,
	encryptor *crypto.AESEncryptor,
	publisher domain.EventPublisher,
	topics config.KafkaTopicsConfig,
) *Service {
	return &Service{
		repo:      repo,
		hasher:    hasher,
		encryptor: encryptor,
		publisher: publisher,
		topics:    topics,
	}
}

// IssueCard - create a new card, encrypt PAN before saving
func (s *Service) IssueCard(ctx context.Context, accountID string, cardType card.Type) (_ *card.Card, err error) {
	defer metrics.ObserveService("CardService", "IssueCard", time.Now(), &err)

	c, err := card.NewCard(accountID, cardType)
	if err != nil {
		return nil, err
	}

	// Encrypt PAN before persisting
	if s.encryptor != nil {
		encrypted, err := s.encryptor.Encrypt(c.PAN)
		if err != nil {
			return nil, err
		}
		c.PAN = encrypted // DB da encrypted saqlanadi
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}

	s.publishCardIssued(ctx, c, "")
	return c, nil
}

func (s *Service) Activate(ctx context.Context, cardID, pin string) (_ *card.Card, err error) {
	defer metrics.ObserveService("CardService", "Activate", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if err := c.Activate(pin, s.hasher); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}

	s.publishCardActivated(ctx, c, "")
	return c, nil
}

func (s *Service) VerifyPIN(ctx context.Context, cardID, pin string) (err error) {
	defer metrics.ObserveService("CardService", "VerifyPIN", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}
	if err := c.VerifyPIN(pin, s.hasher); err != nil {
		s.repo.Update(ctx, c)
		return err
	}
	s.repo.Update(ctx, c)
	return nil
}

func (s *Service) ChangePIN(ctx context.Context, cardID, oldPIN, newPIN string) (err error) {
	defer metrics.ObserveService("CardService", "ChangePIN", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}
	if err := c.ChangePIN(oldPIN, newPIN, s.hasher); err != nil {
		s.repo.Update(ctx, c)
		return err
	}
	return s.repo.Update(ctx, c)
}

func (s *Service) Block(ctx context.Context, cardID, reason string) (err error) {
	defer metrics.ObserveService("CardService", "Block", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}
	if err := c.Block(); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return err
	}

	s.publishCardBlocked(ctx, c, "", reason)
	return nil
}

func (s *Service) Unblock(ctx context.Context, cardID string) (err error) {
	defer metrics.ObserveService("CardService", "Unblock", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}
	if err := c.Unblock(); err != nil {
		return err
	}
	return s.repo.Update(ctx, c)
}

// Enroll3DS - enroll a card in 3D Secure
func (s *Service) Enroll3DS(ctx context.Context, cardID, version string) (_ *card.Card, err error) {
	defer metrics.ObserveService("CardService", "Enroll3DS", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if err := c.Enroll3DS(version); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// SetEMVAID - set EMV Application Identifier for a card
func (s *Service) SetEMVAID(ctx context.Context, cardID, aid string) (_ *card.Card, err error) {
	defer metrics.ObserveService("CardService", "SetEMVAID", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if err := c.SetEMVAID(aid); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) GetByID(ctx context.Context, cardID string) (_ *card.Card, err error) {
	defer metrics.ObserveService("CardService", "GetByID", time.Now(), &err)
	return s.repo.GetByID(ctx, cardID)
}

func (s *Service) ListByAccountID(ctx context.Context, accountID string, limit, offset int) (_ []*card.Card, _ int64, err error) {
	defer metrics.ObserveService("CardService", "ListByAccountID", time.Now(), &err)

	cards, err := s.repo.ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByAccountID(ctx, accountID)
	if err != nil {
		return nil, 0, err
	}
	return cards, total, nil
}

// --- Kafka publish helpers ---

func newCardMetadata(userID string) *commonpb.Metadata {
	return &commonpb.Metadata{
		EventId:   uuid.New().String(),
		UserId:    userID,
		Timestamp: timestamppb.Now(),
		Source:    "xbank-api",
	}
}

func (s *Service) publishCardIssued(ctx context.Context, c *card.Card, userID string) {
	if s.publisher == nil {
		return
	}
	msg := &cardpb.CardIssued{
		Metadata:  newCardMetadata(userID),
		CardId:    c.ID,
		AccountId: c.AccountID,
		UserId:    userID,
		CardType:  string(c.CardType),
		MaskedPan: c.MaskedPAN,
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.CardIssued, c.AccountID, data)
	}
}

func (s *Service) publishCardActivated(ctx context.Context, c *card.Card, userID string) {
	if s.publisher == nil {
		return
	}
	msg := &cardpb.CardActivated{
		Metadata: newCardMetadata(userID),
		CardId:   c.ID,
		UserId:   userID,
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.CardActivated, c.AccountID, data)
	}
}

func (s *Service) publishCardBlocked(ctx context.Context, c *card.Card, userID, reason string) {
	if s.publisher == nil {
		return
	}
	msg := &cardpb.CardBlocked{
		Metadata: newCardMetadata(userID),
		CardId:   c.ID,
		UserId:   userID,
		Reason:   reason,
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.CardBlocked, c.AccountID, data)
	}
}
