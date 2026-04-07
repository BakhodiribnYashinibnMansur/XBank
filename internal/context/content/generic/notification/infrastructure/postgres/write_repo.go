package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WriteRepo struct{ pool *pgxpool.Pool }

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo { return &WriteRepo{pool: pool} }

func (r *WriteRepo) Save(ctx context.Context, n *domain.Notification) error {
	dataJSON, _ := json.Marshal(n.Data)
	return r.pool.QueryRow(ctx,
		`INSERT INTO notifications (user_id, title, message, type, data, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		n.UserID, n.Title, n.Message, n.Type, dataJSON, n.CreatedAt, n.UpdatedAt,
	).Scan(&n.ID)
}

func (r *WriteRepo) Update(ctx context.Context, n *domain.Notification) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET read_at = $1, updated_at = $2 WHERE id = $3`,
		n.ReadAt, n.UpdatedAt, n.ID)
	return err
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notifications WHERE id = $1`, id)
	return err
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	n := &domain.Notification{}
	var dataJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, message, type, data, read_at, created_at, updated_at
		 FROM notifications WHERE id = $1`, id,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &dataJSON, &n.ReadAt, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("notification find: %w", err)
	}
	if dataJSON != nil {
		json.Unmarshal(dataJSON, &n.Data)
	}
	return n, nil
}
