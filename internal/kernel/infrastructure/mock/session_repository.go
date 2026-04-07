package mock

import (
	"context"
	"sync"

	session "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session/domain"
	"github.com/google/uuid"
)

type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*session.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{sessions: make(map[string]*session.Session)}
}

func (r *SessionRepository) Create(ctx context.Context, s *session.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s.ID = uuid.New().String()
	r.sessions[s.ID] = s
	return nil
}

func (r *SessionRepository) GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, s := range r.sessions {
		if s.RefreshToken == refreshTokenHash {
			return s, nil
		}
	}
	return nil, session.ErrSessionNotFound
}

func (r *SessionRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, id)
	return nil
}

func (r *SessionRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, s := range r.sessions {
		if s.UserID == userID {
			delete(r.sessions, id)
		}
	}
	return nil
}
