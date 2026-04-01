package postgres

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	db := ExtractDBTX(ctx, r.pool)
	query := `
		INSERT INTO users (email, password, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	return db.QueryRow(ctx, query,
		u.Email, u.Password, u.FirstName, u.LastName, u.Role, u.CreatedAt, u.UpdatedAt,
	).Scan(&u.ID)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	db := ExtractDBTX(ctx, r.pool)
	u := &user.User{}
	err := db.QueryRow(ctx,
		`SELECT id, email, password, first_name, last_name, role, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	db := ExtractDBTX(ctx, r.pool)
	u := &user.User{}
	err := db.QueryRow(ctx,
		`SELECT id, email, password, first_name, last_name, role, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	db := ExtractDBTX(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}
