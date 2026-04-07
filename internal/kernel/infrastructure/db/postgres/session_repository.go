package postgres

import (
	"context"
	"time"

	session "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, s *session.Session) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	query := `
		INSERT INTO sessions (user_id, refresh_token, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := db.QueryRow(ctx, query,
		s.UserID, s.RefreshToken, s.UserAgent, s.IPAddress, s.ExpiresAt, s.CreatedAt,
	).Scan(&s.ID)
	metrics.ObserveQuery("SessionRepo.Create", start, err)
	return err
}

func (r *SessionRepository) GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*session.Session, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
		FROM sessions WHERE refresh_token = $1`

	s := &session.Session{}
	err := db.QueryRow(ctx, query, refreshTokenHash).Scan(
		&s.ID, &s.UserID, &s.RefreshToken, &s.UserAgent, &s.IPAddress, &s.ExpiresAt, &s.CreatedAt,
	)
	metrics.ObserveQuery("SessionRepo.GetByRefreshToken", start, err)
	if err != nil {
		return nil, session.ErrSessionNotFound
	}
	return s, nil
}

func (r *SessionRepository) DeleteByID(ctx context.Context, id string) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	metrics.ObserveQuery("SessionRepo.DeleteByID", start, err)
	return err
}

func (r *SessionRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	metrics.ObserveQuery("SessionRepo.DeleteAllByUserID", start, err)
	return err
}
