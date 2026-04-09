package mock

import (
	"context"
	"sync"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	user "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	"github.com/google/uuid"
)

// MockUserAuthReader implements ports.UserAuthReader for testing.
// It stores user.User internally but returns ports.UserAuthView.
type MockUserAuthReader struct {
	mu    sync.RWMutex
	users map[string]*user.User
}

func NewMockUserAuthReader() *MockUserAuthReader {
	return &MockUserAuthReader{users: make(map[string]*user.User)}
}

// Create adds a user (test helper, not part of the port interface).
func (r *MockUserAuthReader) Create(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	r.users[u.ID] = u
	return nil
}

// CreateRaw adds a user without importing user/domain (for cross-BC test isolation).
func (r *MockUserAuthReader) CreateRaw(email, hashedPassword, firstName, lastName string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := uuid.New().String()
	r.users[id] = &user.User{
		ID:        id,
		Email:     email,
		Password:  hashedPassword,
		FirstName: firstName,
		LastName:  lastName,
		Role:      user.RoleCustomer,
	}
	return id
}

// GetInternalByEmail returns the raw user entity (test helper for TOTP verification).
func (r *MockUserAuthReader) GetInternalByEmail(ctx context.Context, email string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound
}

// GetInternalByID returns the raw user entity (test helper).
func (r *MockUserAuthReader) GetInternalByID(ctx context.Context, id string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (r *MockUserAuthReader) GetByID(ctx context.Context, id string) (*ports.UserAuthView, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return toAuthView(u), nil
}

func (r *MockUserAuthReader) GetByEmail(ctx context.Context, email string) (*ports.UserAuthView, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return toAuthView(u), nil
		}
	}
	return nil, user.ErrUserNotFound
}

func (r *MockUserAuthReader) UpdateTOTP(ctx context.Context, userID, totpSecret string, enabled bool, verifiedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[userID]; ok {
		u.TOTPSecret = totpSecret
		u.TOTPEnabled = enabled
		u.TOTPVerifiedAt = verifiedAt
	}
	return nil
}

func toAuthView(u *user.User) *ports.UserAuthView {
	return &ports.UserAuthView{
		ID:          u.ID,
		Email:       u.Email,
		Password:    u.Password,
		Role:        string(u.Role),
		TOTPSecret:  u.TOTPSecret,
		TOTPEnabled: u.TOTPEnabled,
	}
}
