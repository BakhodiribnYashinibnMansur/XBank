package beneficiary

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

var (
	ErrBeneficiaryNotFound = apperror.ErrBeneficiaryNotFound
	ErrBeneficiaryExists   = apperror.ErrBeneficiaryExists
)

type Type string

const (
	TypeInternal      Type = "INTERNAL"      // XBank ichida
	TypeExternal      Type = "EXTERNAL"      // boshqa bank (mahalliy)
	TypeInternational Type = "INTERNATIONAL" // xalqaro
)

// Beneficiary - trusted transfer recipient
type Beneficiary struct {
	ID            string
	UserID        string // who owns this beneficiary
	Name          string // recipient name
	AccountNumber string // IBAN or local account
	BankName      string
	BankCode      string // BIC/SWIFT for international
	Currency      string
	BenType       Type
	Verified      bool
	CreatedAt     time.Time
}

// NewBeneficiary - create a new beneficiary
func NewBeneficiary(userID, name, accountNumber, bankName, bankCode, currency string, benType Type) (*Beneficiary, error) {
	if userID == "" {
		return nil, apperror.ErrMissingField.WithMessage("user_id is required")
	}
	if name == "" {
		return nil, apperror.ErrMissingField.WithMessage("name is required")
	}
	if accountNumber == "" {
		return nil, apperror.ErrMissingField.WithMessage("account_number is required")
	}

	return &Beneficiary{
		UserID:        userID,
		Name:          name,
		AccountNumber: accountNumber,
		BankName:      bankName,
		BankCode:      bankCode,
		Currency:      currency,
		BenType:       benType,
		Verified:      false,
		CreatedAt:     time.Now(),
	}, nil
}

type Repository interface {
	Create(ctx context.Context, b *Beneficiary) error
	GetByID(ctx context.Context, id string) (*Beneficiary, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*Beneficiary, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
	Delete(ctx context.Context, id string) error
	ExistsByUserAndAccount(ctx context.Context, userID, accountNumber string) (bool, error)
}
