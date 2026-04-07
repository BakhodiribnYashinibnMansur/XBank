package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReadRepo struct{ pool *pgxpool.Pool }

func NewReadRepo(pool *pgxpool.Pool) *ReadRepo { return &ReadRepo{pool: pool} }

func (r *ReadRepo) FindByID(ctx context.Context, id string) (*domain.AnnouncementView, error) {
	v := &domain.AnnouncementView{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, title_uz, title_ru, title_en, body_uz, body_ru, body_en, priority, status, start_date, end_date,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM announcements WHERE id = $1`, id,
	).Scan(&v.ID, &v.TitleUz, &v.TitleRu, &v.TitleEn, &v.BodyUz, &v.BodyRu, &v.BodyEn,
		&v.Priority, &v.Status, &v.StartDate, &v.EndDate, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("announcement read: %w", err)
	}
	return v, nil
}

func (r *ReadRepo) List(ctx context.Context, filter domain.AnnouncementFilter) ([]*domain.AnnouncementView, int64, error) {
	countQ := `SELECT COUNT(*) FROM announcements WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if filter.Status != "" {
		countQ += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, filter.Status)
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := `SELECT id, title_uz, title_ru, title_en, body_uz, body_ru, body_en, priority, status, start_date, end_date,
	                 to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	          FROM announcements WHERE 1=1`
	lArgs := []interface{}{}
	lIdx := 1

	if filter.Status != "" {
		listQ += fmt.Sprintf(` AND status = $%d`, lIdx)
		lArgs = append(lArgs, filter.Status)
		lIdx++
	}

	listQ += fmt.Sprintf(` ORDER BY priority DESC, created_at DESC LIMIT $%d OFFSET $%d`, lIdx, lIdx+1)
	lArgs = append(lArgs, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQ, lArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*domain.AnnouncementView
	for rows.Next() {
		v := &domain.AnnouncementView{}
		if err := rows.Scan(&v.ID, &v.TitleUz, &v.TitleRu, &v.TitleEn, &v.BodyUz, &v.BodyRu, &v.BodyEn,
			&v.Priority, &v.Status, &v.StartDate, &v.EndDate, &v.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, nil
}

func (r *ReadRepo) ListActive(ctx context.Context, now time.Time) ([]*domain.AnnouncementView, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title_uz, title_ru, title_en, body_uz, body_ru, body_en, priority, status, start_date, end_date,
		        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		 FROM announcements
		 WHERE status = 'PUBLISHED'
		   AND (start_date IS NULL OR start_date <= $1)
		   AND (end_date IS NULL OR end_date >= $1)
		 ORDER BY priority DESC`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.AnnouncementView
	for rows.Next() {
		v := &domain.AnnouncementView{}
		if err := rows.Scan(&v.ID, &v.TitleUz, &v.TitleRu, &v.TitleEn, &v.BodyUz, &v.BodyRu, &v.BodyEn,
			&v.Priority, &v.Status, &v.StartDate, &v.EndDate, &v.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, nil
}
