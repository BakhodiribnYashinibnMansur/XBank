package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/user"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	query := `
		INSERT INTO users (email, password, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := db.QueryRow(ctx, query,
		u.Email, u.Password, u.FirstName, u.LastName, u.Role, u.CreatedAt, u.UpdatedAt,
	).Scan(&u.ID)
	metrics.ObserveQuery("UserRepo.Create", start, err)
	if err != nil {
		return fmt.Errorf("user_repo: create: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	u := &user.User{}
	err := db.QueryRow(ctx,
		`SELECT id, email, password, first_name, last_name, role, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	metrics.ObserveQuery("UserRepo.GetByID", start, err)
	if err != nil {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	u := &user.User{}
	err := db.QueryRow(ctx,
		`SELECT id, email, password, first_name, last_name, role, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	metrics.ObserveQuery("UserRepo.GetByEmail", start, err)
	if err != nil {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	metrics.ObserveQuery("UserRepo.ExistsByEmail", start, err)
	return exists, err
}
