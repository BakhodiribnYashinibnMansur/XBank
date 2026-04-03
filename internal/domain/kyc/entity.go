package kyc

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

var (
	ErrKYCRequired = shared.NewDomainError("KYC_REQUIRED", "KYC verification required")
	ErrKYCPending  = shared.NewDomainError("KYC_PENDING", "KYC verification is pending")
)

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusApproved Status = "APPROVED"
	StatusRejected Status = "REJECTED"
)

type DocType string

const (
	DocPassport      DocType = "PASSPORT"
	DocIDCard        DocType = "ID_CARD"
	DocDriverLicense DocType = "DRIVER_LICENSE"
)

// Verification - KYC verification record
type Verification struct {
	ID             string
	UserID         string
	DocumentType   DocType
	DocumentNumber string // encrypted
	FirstName      string
	LastName       string
	DateOfBirth    string
	Status         Status
	RejectedReason string
	ReviewedBy     string // admin user ID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewVerification(userID string, docType DocType, docNumber, firstName, lastName, dob string) (*Verification, error) {
	if userID == "" || docNumber == "" || firstName == "" {
		return nil, shared.NewDomainError("MISSING_FIELD", "user_id, document_number and first_name are required")
	}

	now := time.Now()
	return &Verification{
		UserID:         userID,
		DocumentType:   docType,
		DocumentNumber: docNumber,
		FirstName:      firstName,
		LastName:       lastName,
		DateOfBirth:    dob,
		Status:         StatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (v *Verification) Approve(reviewerID string) {
	v.Status = StatusApproved
	v.ReviewedBy = reviewerID
	v.UpdatedAt = time.Now()
}

func (v *Verification) Reject(reviewerID, reason string) {
	v.Status = StatusRejected
	v.ReviewedBy = reviewerID
	v.RejectedReason = reason
	v.UpdatedAt = time.Now()
}

type Repository interface {
	Create(ctx context.Context, v *Verification) error
	GetByID(ctx context.Context, id string) (*Verification, error)
	GetByUserID(ctx context.Context, userID string) (*Verification, error)
	Update(ctx context.Context, v *Verification) error
	ListPending(ctx context.Context, limit, offset int) ([]*Verification, error)
	CountPending(ctx context.Context) (int64, error)
}
