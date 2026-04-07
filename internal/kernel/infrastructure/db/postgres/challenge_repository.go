package postgres

import (
	"context"
	"time"

	challenge "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChallengeRepository struct {
	pool *pgxpool.Pool
}

func NewChallengeRepository(pool *pgxpool.Pool) *ChallengeRepository {
	return &ChallengeRepository{pool: pool}
}

func (r *ChallengeRepository) Create(ctx context.Context, c *challenge.Challenge) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`INSERT INTO challenges (id, user_id, method, status, action, metadata, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.UserID, c.Method, c.Status, c.Action, c.Metadata, c.ExpiresAt, c.CreatedAt,
	)
	metrics.ObserveQuery("ChallengeRepo.Create", start, err)
	return err
}

func (r *ChallengeRepository) GetByID(ctx context.Context, id string) (*challenge.Challenge, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	c := &challenge.Challenge{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, method, status, token, action, metadata, expires_at, created_at, verified_at
		 FROM challenges WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.UserID, &c.Method, &c.Status, &c.Token, &c.Action, &c.Metadata, &c.ExpiresAt, &c.CreatedAt, &c.VerifiedAt)

	metrics.ObserveQuery("ChallengeRepo.GetByID", start, err)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ChallengeRepository) GetByToken(ctx context.Context, token string) (*challenge.Challenge, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	c := &challenge.Challenge{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, method, status, token, action, metadata, expires_at, created_at, verified_at
		 FROM challenges WHERE token = $1`,
		token,
	).Scan(&c.ID, &c.UserID, &c.Method, &c.Status, &c.Token, &c.Action, &c.Metadata, &c.ExpiresAt, &c.CreatedAt, &c.VerifiedAt)

	metrics.ObserveQuery("ChallengeRepo.GetByToken", start, err)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ChallengeRepository) Update(ctx context.Context, c *challenge.Challenge) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`UPDATE challenges SET status = $1, token = $2, expires_at = $3, verified_at = $4
		 WHERE id = $5`,
		c.Status, c.Token, c.ExpiresAt, c.VerifiedAt, c.ID,
	)
	metrics.ObserveQuery("ChallengeRepo.Update", start, err)
	return err
}

func (r *ChallengeRepository) CountPendingByUser(ctx context.Context, userID string) (int, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	var count int
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM challenges WHERE user_id = $1 AND status = 'PENDING' AND expires_at > NOW()`,
		userID,
	).Scan(&count)
	metrics.ObserveQuery("ChallengeRepo.CountPendingByUser", start, err)
	return count, err
}
