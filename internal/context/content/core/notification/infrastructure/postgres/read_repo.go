package postgres

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReadRepo struct{ pool *pgxpool.Pool }

func NewReadRepo(pool *pgxpool.Pool) *ReadRepo { return &ReadRepo{pool: pool} }

func (r *ReadRepo) FindByID(ctx context.Context, id string) (*repository.NotificationView, error) {
	v := &repository.NotificationView{}
	var read bool
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, message, type, (read_at IS NOT NULL) as read,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM notifications WHERE id = $1`, id,
	).Scan(&v.ID, &v.UserID, &v.Title, &v.Message, &v.Type, &read, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("notification read: %w", err)
	}
	v.Read = read
	return v, nil
}

func (r *ReadRepo) List(ctx context.Context, filter repository.NotificationFilter) ([]*repository.NotificationView, int64, error) {
	countQ := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	args := []interface{}{filter.UserID}
	idx := 2

	if filter.Type != "" {
		countQ += fmt.Sprintf(` AND type = $%d`, idx)
		args = append(args, filter.Type)
		idx++
	}
	if filter.Unread != nil && *filter.Unread {
		countQ += ` AND read_at IS NULL`
	}

	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := `SELECT id, user_id, title, message, type, (read_at IS NOT NULL) as read,
	                 to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	          FROM notifications WHERE user_id = $1`
	lArgs := []interface{}{filter.UserID}
	lIdx := 2

	if filter.Type != "" {
		listQ += fmt.Sprintf(` AND type = $%d`, lIdx)
		lArgs = append(lArgs, filter.Type)
		lIdx++
	}
	if filter.Unread != nil && *filter.Unread {
		listQ += ` AND read_at IS NULL`
	}

	listQ += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, lIdx, lIdx+1)
	lArgs = append(lArgs, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQ, lArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*repository.NotificationView
	for rows.Next() {
		v := &repository.NotificationView{}
		if err := rows.Scan(&v.ID, &v.UserID, &v.Title, &v.Message, &v.Type, &v.Read, &v.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, nil
}
