package mock

import (
	"context"
	"sync"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/user"
	"github.com/google/uuid"
)

// UserRepository - in-memory user repository for testing
type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*user.User // id -> user
}

func NewUserRepository() *UserRepository {
	return &UserRepository{users: make(map[string]*user.User)}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u.ID = uuid.New().String()
	r.users[u.ID] = u
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[userID]; ok {
		u.Password = hashedPassword
	}
	return nil
}

func (r *UserRepository) UpdateTOTP(ctx context.Context, userID, totpSecret string, enabled bool, verifiedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[userID]; ok {
		u.TOTPSecret = totpSecret
		u.TOTPEnabled = enabled
		u.TOTPVerifiedAt = verifiedAt
	}
	return nil
}

func (r *UserRepository) Anonymize(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[userID]; ok {
		u.Email = "deleted_" + userID
		u.FirstName = "[DELETED]"
		u.LastName = "[DELETED]"
		u.Password = ""
	}
	return nil
}
