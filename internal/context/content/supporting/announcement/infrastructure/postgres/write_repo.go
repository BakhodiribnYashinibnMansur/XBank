package postgres

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WriteRepo struct{ pool *pgxpool.Pool }

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo { return &WriteRepo{pool: pool} }

func (r *WriteRepo) Save(ctx context.Context, a *domain.Announcement) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO announcements (title_uz, title_ru, title_en, body_uz, body_ru, body_en, priority, status, start_date, end_date, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
		a.TitleUz, a.TitleRu, a.TitleEn, a.BodyUz, a.BodyRu, a.BodyEn,
		a.Priority, a.Status, a.StartDate, a.EndDate, a.CreatedAt, a.UpdatedAt,
	).Scan(&a.ID)
}

func (r *WriteRepo) Update(ctx context.Context, a *domain.Announcement) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE announcements SET title_uz=$1, title_ru=$2, title_en=$3, body_uz=$4, body_ru=$5, body_en=$6,
		 priority=$7, status=$8, start_date=$9, end_date=$10, updated_at=$11 WHERE id=$12`,
		a.TitleUz, a.TitleRu, a.TitleEn, a.BodyUz, a.BodyRu, a.BodyEn,
		a.Priority, a.Status, a.StartDate, a.EndDate, a.UpdatedAt, a.ID)
	return err
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM announcements WHERE id = $1`, id)
	return err
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.Announcement, error) {
	a := &domain.Announcement{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, title_uz, title_ru, title_en, body_uz, body_ru, body_en, priority, status, start_date, end_date, created_at, updated_at
		 FROM announcements WHERE id = $1`, id,
	).Scan(&a.ID, &a.TitleUz, &a.TitleRu, &a.TitleEn, &a.BodyUz, &a.BodyRu, &a.BodyEn,
		&a.Priority, &a.Status, &a.StartDate, &a.EndDate, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("announcement find: %w", err)
	}
	return a, nil
}
