package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// Domain errors
var (
	ErrInvalidEmail    = domain.NewDomainError("INVALID_EMAIL", "invalid email format")
	ErrInvalidPassword = domain.NewDomainError("INVALID_PASSWORD", "password must be at least 8 characters")
	ErrInvalidName     = domain.NewDomainError("INVALID_NAME", "name cannot be empty")
	ErrUserNotFound    = domain.NewDomainError("USER_NOT_FOUND", "user not found")
	ErrEmailExists     = domain.NewDomainError("EMAIL_EXISTS", "this email is already registered")
)

// Role - user role for RBAC
type Role string

const (
	RoleCustomer Role = "CUSTOMER"
	RoleTeller   Role = "TELLER"
	RoleAdmin    Role = "ADMIN"
)

// User - core entity (business object)
type User struct {
	ID        string
	Email     string
	Password  string // hashed
	FirstName string
	LastName  string
	Role      Role

	// TOTP (2FA)
	TOTPSecret     string     // base32-encoded shared secret
	TOTPEnabled    bool       // whether 2FA is active
	TOTPVerifiedAt *time.Time // when user first verified TOTP setup

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser - create a new user (with business validation)
// This is a factory function - the only correct way to create a user
func NewUser(email, hashedPassword, firstName, lastName string) (*User, error) {
	if email == "" {
		return nil, ErrInvalidEmail
	}
	if hashedPassword == "" {
		return nil, ErrInvalidPassword
	}
	if firstName == "" {
		return nil, ErrInvalidName
	}

	now := time.Now()
	return &User{
		Email:     email,
		Password:  hashedPassword,
		FirstName: firstName,
		LastName:  lastName,
		Role:      RoleCustomer, // default role
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Repository - the domain layer ONLY defines the interface.
// It does not know (and must not know) which database is used.
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdatePassword(ctx context.Context, userID, hashedPassword string) error
	UpdateTOTP(ctx context.Context, userID, totpSecret string, enabled bool, verifiedAt *time.Time) error
	Anonymize(ctx context.Context, userID string) error // GDPR right to erasure
}
