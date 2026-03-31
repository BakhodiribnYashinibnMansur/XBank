package user

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

// Domain errors
var (
	ErrInvalidEmail    = apperror.ErrInvalidEmail
	ErrInvalidPassword = apperror.ErrInvalidPassword
	ErrInvalidName     = apperror.ErrInvalidName
	ErrUserNotFound    = apperror.ErrUserNotFound
	ErrEmailExists     = apperror.ErrEmailExists
)

// User - core entity (business object)
type User struct {
	ID        string
	Email     string
	Password  string // hashed
	FirstName string
	LastName  string
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
}
