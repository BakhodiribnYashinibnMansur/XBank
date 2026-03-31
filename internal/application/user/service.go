package user

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

// Service - manages all use cases related to User
// Depends on domain repository interface (not on PostgreSQL!)
type Service struct {
	repo user.Repository
}

// NewService - create a new service
func NewService(repo user.Repository) *Service {
	return &Service{repo: repo}
}

// Register - register a new user (use case)
func (s *Service) Register(ctx context.Context, email, password, firstName, lastName string) (*user.User, error) {
	// 1. Check that the email is not already taken
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, user.ErrEmailExists
	}

	// 2. Hash the password (never store it in plain text!)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Create domain entity (business validation happens here)
	u, err := user.NewUser(email, string(hashedPassword), firstName, lastName)
	if err != nil {
		return nil, err
	}

	// 4. Save to database
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

// GetByID - get a user by ID
func (s *Service) GetByID(ctx context.Context, id string) (*user.User, error) {
	return s.repo.GetByID(ctx, id)
}
