package postgres

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WriteRepo struct{ pool *pgxpool.Pool }

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo { return &WriteRepo{pool: pool} }

func (r *WriteRepo) Save(ctx context.Context, e *domain.ErrorCode) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO error_codes (code, message_en, message_uz, message_ru, category, severity, http_status, retryable, suggestion, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`,
		e.Code, e.MessageEn, e.MessageUz, e.MessageRu, e.Category, e.Severity,
		e.HTTPStatus, e.Retryable, e.Suggestion, e.CreatedAt, e.UpdatedAt,
	).Scan(&e.ID)
}

func (r *WriteRepo) Update(ctx context.Context, e *domain.ErrorCode) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE error_codes SET message_en=$1, message_uz=$2, message_ru=$3, http_status=$4, retryable=$5, suggestion=$6, updated_at=$7 WHERE id=$8`,
		e.MessageEn, e.MessageUz, e.MessageRu, e.HTTPStatus, e.Retryable, e.Suggestion, e.UpdatedAt, e.ID)
	return err
}

func (r *WriteRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM error_codes WHERE id = $1`, id)
	return err
}

func (r *WriteRepo) FindByID(ctx context.Context, id string) (*domain.ErrorCode, error) {
	e := &domain.ErrorCode{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, code, message_en, message_uz, message_ru, category, severity, http_status, retryable, suggestion, created_at, updated_at
		 FROM error_codes WHERE id = $1`, id,
	).Scan(&e.ID, &e.Code, &e.MessageEn, &e.MessageUz, &e.MessageRu, &e.Category, &e.Severity,
		&e.HTTPStatus, &e.Retryable, &e.Suggestion, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error_code find: %w", err)
	}
	return e, nil
}

func (r *WriteRepo) FindByCode(ctx context.Context, code string) (*domain.ErrorCode, error) {
	e := &domain.ErrorCode{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, code, message_en, message_uz, message_ru, category, severity, http_status, retryable, suggestion, created_at, updated_at
		 FROM error_codes WHERE code = $1`, code,
	).Scan(&e.ID, &e.Code, &e.MessageEn, &e.MessageUz, &e.MessageRu, &e.Category, &e.Severity,
		&e.HTTPStatus, &e.Retryable, &e.Suggestion, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error_code find by code: %w", err)
	}
	return e, nil
}
