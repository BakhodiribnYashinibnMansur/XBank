package contact

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

var (
	ErrContactNotFound = shared.NewDomainError("CONTACT_NOT_FOUND", "contact not found")
	ErrContactExists   = shared.NewDomainError("CONTACT_EXISTS", "contact already exists")
	ErrContactSelf     = shared.NewDomainError("CONTACT_SELF", "cannot add yourself as a contact")
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
		return nil, shared.NewDomainError("MISSING_FIELD", "owner_id is required")
	}
	if contactID == "" {
		return nil, shared.NewDomainError("MISSING_FIELD", "contact_id is required")
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
