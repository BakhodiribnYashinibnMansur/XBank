package card

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/crypto"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
)

type Service struct {
	repo      card.Repository
	encryptor *crypto.AESEncryptor // nil = no encryption (dev mode)
}

func NewService(repo card.Repository, encryptor *crypto.AESEncryptor) *Service {
	return &Service{repo: repo, encryptor: encryptor}
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
	return c, nil
}

func (s *Service) Activate(ctx context.Context, cardID, pin string) (_ *card.Card, err error) {
	defer metrics.ObserveService("CardService", "Activate", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if err := c.Activate(pin); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) VerifyPIN(ctx context.Context, cardID, pin string) (err error) {
	defer metrics.ObserveService("CardService", "VerifyPIN", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}
	if err := c.VerifyPIN(pin); err != nil {
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
	if err := c.ChangePIN(oldPIN, newPIN); err != nil {
		s.repo.Update(ctx, c)
		return err
	}
	return s.repo.Update(ctx, c)
}

func (s *Service) Block(ctx context.Context, cardID string) (err error) {
	defer metrics.ObserveService("CardService", "Block", time.Now(), &err)

	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}
	if err := c.Block(); err != nil {
		return err
	}
	return s.repo.Update(ctx, c)
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
