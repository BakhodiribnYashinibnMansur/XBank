package card

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
)

type Service struct {
	repo card.Repository
}

func NewService(repo card.Repository) *Service {
	return &Service{repo: repo}
}

// IssueCard - create a new card for an account
func (s *Service) IssueCard(ctx context.Context, accountID string, cardType card.Type) (*card.Card, error) {
	c, err := card.NewCard(accountID, cardType)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Activate - activate a card by setting PIN
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

// VerifyPIN - verify card PIN
func (s *Service) VerifyPIN(ctx context.Context, cardID, pin string) error {
	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}

	if err := c.VerifyPIN(pin); err != nil {
		s.repo.Update(ctx, c) // save updated attempts/status
		return err
	}

	s.repo.Update(ctx, c) // reset attempts
	return nil
}

// ChangePIN - change card PIN
func (s *Service) ChangePIN(ctx context.Context, cardID, oldPIN, newPIN string) error {
	c, err := s.repo.GetByID(ctx, cardID)
	if err != nil {
		return err
	}

	if err := c.ChangePIN(oldPIN, newPIN); err != nil {
		s.repo.Update(ctx, c) // save updated attempts
		return err
	}

	return s.repo.Update(ctx, c)
}

// Block - block a card
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

// Unblock - unblock a card
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

// GetByID - get card by ID
func (s *Service) GetByID(ctx context.Context, cardID string) (*card.Card, error) {
	return s.repo.GetByID(ctx, cardID)
}

// ListByAccountID - list cards for an account
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
