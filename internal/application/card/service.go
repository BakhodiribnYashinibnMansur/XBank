package card

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/crypto"
)

type Service struct {
	repo      card.Repository
	encryptor *crypto.AESEncryptor // nil = no encryption (dev mode)
}

func NewService(repo card.Repository, encryptor *crypto.AESEncryptor) *Service {
	return &Service{repo: repo, encryptor: encryptor}
}

// IssueCard - create a new card, encrypt PAN before saving
func (s *Service) IssueCard(ctx context.Context, accountID string, cardType card.Type) (*card.Card, error) {
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

func (s *Service) Activate(ctx context.Context, cardID, pin string) (*card.Card, error) {
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

func (s *Service) VerifyPIN(ctx context.Context, cardID, pin string) error {
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

func (s *Service) ChangePIN(ctx context.Context, cardID, oldPIN, newPIN string) error {
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

func (s *Service) Block(ctx context.Context, cardID string) error {
	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}
	if err := c.Block(); err != nil {
		return err
	}
	return s.repo.Update(ctx, c)
}

func (s *Service) Unblock(ctx context.Context, cardID string) error {
	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}
	if err := c.Unblock(); err != nil {
		return err
	}
	return s.repo.Update(ctx, c)
}

func (s *Service) GetByID(ctx context.Context, cardID string) (*card.Card, error) {
	return s.repo.GetByID(ctx, cardID)
}

func (s *Service) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*card.Card, int64, error) {
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
