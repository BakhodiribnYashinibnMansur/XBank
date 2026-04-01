package contact

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

var (
	ErrContactNotFound = apperror.ErrContactNotFound
	ErrContactExists   = apperror.ErrContactExists
	ErrContactSelf     = apperror.ErrContactSelf
)

// Contact represents a user's saved contact (another XBank user).
type Contact struct {
	ID         string
	OwnerID    string // who owns this contact
	ContactID  string // the referenced user
	CustomName string // owner's custom label
	IsBlocked  bool
	CreatedAt  time.Time
}

// NewContact creates a new Contact with validation.
func NewContact(ownerID, contactID, customName string) (*Contact, error) {
	if ownerID == "" {
		return nil, apperror.ErrMissingField.WithMessage("owner_id is required")
	}
	if contactID == "" {
		return nil, apperror.ErrMissingField.WithMessage("contact_id is required")
	}
	if ownerID == contactID {
		return nil, ErrContactSelf
	}

	return &Contact{
		OwnerID:    ownerID,
		ContactID:  contactID,
		CustomName: customName,
		IsBlocked:  false,
		CreatedAt:  time.Now(),
	}, nil
}

type Repository interface {
	Add(ctx context.Context, contact *Contact) error
	GetByID(ctx context.Context, id string) (*Contact, error)
	ListByOwnerID(ctx context.Context, ownerID string, limit, offset int) ([]*Contact, error)
	CountByOwnerID(ctx context.Context, ownerID string) (int64, error)
	Delete(ctx context.Context, ownerID, contactID string) error
	IsContact(ctx context.Context, ownerID, contactID string) (bool, error)
}
