package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/contact"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

type Service struct {
	repo contact.Repository
}

func NewService(repo contact.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, ownerID, contactID, customName string) (_ *contact.Contact, err error) {
	defer metrics.ObserveService("ContactService", "Add", time.Now(), &err)

	c, err := contact.NewContact(ownerID, contactID, customName)
	if err != nil {
		return nil, err
	}

	exists, _ := s.repo.IsContact(ctx, ownerID, contactID)
	if exists {
		return nil, contact.ErrContactExists
	}

	if err := s.repo.Add(ctx, c); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, ownerID string, limit, offset int) (_ []*contact.Contact, _ int64, err error) {
	defer metrics.ObserveService("ContactService", "List", time.Now(), &err)

	contacts, err := s.repo.ListByOwnerID(ctx, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, 0, err
	}
	return contacts, total, nil
}

func (s *Service) Delete(ctx context.Context, ownerID, contactID string) (err error) {
	defer metrics.ObserveService("ContactService", "Delete", time.Now(), &err)
	return s.repo.Delete(ctx, ownerID, contactID)
}
