package postgres

import (
	"context"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WriteRepo struct {
	pool *pgxpool.Pool
}

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

func (r *WriteRepo) Create(ctx context.Context, c *domain.Challenge) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`INSERT INTO challenges (id, user_id, method, status, action, metadata, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.UserID, c.Method, c.Status, c.Action, c.Metadata, c.ExpiresAt, c.CreatedAt,
	)
	metrics.ObserveQuery("ChallengeRepo.Create", start, err)
	return err
}

func (r *WriteRepo) GetByID(ctx context.Context, id string) (*domain.Challenge, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	c := &domain.Challenge{}
	var token *string
	err := db.QueryRow(ctx,
		`SELECT id, user_id, method, status, token, action, metadata, expires_at, created_at, verified_at
		 FROM challenges WHERE id = $1`, id,
	).Scan(&c.ID, &c.UserID, &c.Method, &c.Status, &token, &c.Action, &c.Metadata, &c.ExpiresAt, &c.CreatedAt, &c.VerifiedAt)
	metrics.ObserveQuery("ChallengeRepo.GetByID", start, err)
	if err != nil {
		return nil, err
	}
	if token != nil {
		c.Token = *token
	}
	return c, nil
}

func (r *WriteRepo) GetByToken(ctx context.Context, tok string) (*domain.Challenge, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	c := &domain.Challenge{}
	var scannedToken *string
	err := db.QueryRow(ctx,
		`SELECT id, user_id, method, status, token, action, metadata, expires_at, created_at, verified_at
		 FROM challenges WHERE token = $1`, tok,
	).Scan(&c.ID, &c.UserID, &c.Method, &c.Status, &scannedToken, &c.Action, &c.Metadata, &c.ExpiresAt, &c.CreatedAt, &c.VerifiedAt)
	metrics.ObserveQuery("ChallengeRepo.GetByToken", start, err)
	if err != nil {
		return nil, err
	}
	if scannedToken != nil {
		c.Token = *scannedToken
	}
	return c, nil
}

func (r *WriteRepo) Update(ctx context.Context, c *domain.Challenge) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE challenges SET status = $1, token = $2, expires_at = $3, verified_at = $4
		 WHERE id = $5`,
		c.Status, c.Token, c.ExpiresAt, c.VerifiedAt, c.ID,
	)
	metrics.ObserveQuery("ChallengeRepo.Update", start, err)
	return err
}

func (r *WriteRepo) CountPendingByUser(ctx context.Context, userID string) (int, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var count int
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM challenges WHERE user_id = $1 AND status = 'PENDING' AND expires_at > NOW()`,
		userID,
	).Scan(&count)
	metrics.ObserveQuery("ChallengeRepo.CountPendingByUser", start, err)
	return count, err
}
