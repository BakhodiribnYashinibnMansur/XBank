package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/contact/domain"
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

func (r *WriteRepo) Add(ctx context.Context, c *domain.Contact) error {
	start := time.Now()
	query := `
		INSERT INTO user_contacts (owner_id, contact_id, custom_name, is_blocked, created_at)
		VALUES ($1, $2, $3, FALSE, now())
		ON CONFLICT (owner_id, contact_id) DO NOTHING
	`
	_, err := sharedpg.ExtractDBTX(ctx, r.pool).Exec(ctx, query, c.OwnerID, c.ContactID, c.CustomName)
	if err != nil {
		metrics.ObserveQuery("ContactRepo.Add", start, err)
		return fmt.Errorf("contact_repo: add: %w", err)
	}
	metrics.ObserveQuery("ContactRepo.Add", start, nil)
	return nil
}

func (r *WriteRepo) GetByID(ctx context.Context, id string) (*domain.Contact, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	c := &domain.Contact{}
	err := db.QueryRow(ctx,
		`SELECT id, owner_id, contact_id, custom_name, is_blocked, created_at
		 FROM user_contacts WHERE id = $1`, id,
	).Scan(&c.ID, &c.OwnerID, &c.ContactID, &c.CustomName, &c.IsBlocked, &c.CreatedAt)
	metrics.ObserveQuery("ContactRepo.GetByID", start, err)
	if err != nil {
		return nil, domain.ErrContactNotFound
	}
	return c, nil
}

func (r *WriteRepo) ListByOwnerID(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Contact, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, owner_id, contact_id, custom_name, is_blocked, created_at
		 FROM user_contacts WHERE owner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		ownerID, limit, offset,
	)
	if err != nil {
		metrics.ObserveQuery("ContactRepo.ListByOwnerID", start, err)
		return nil, fmt.Errorf("contact_repo: list: %w", err)
	}
	defer rows.Close()

	var items []*domain.Contact
	for rows.Next() {
		c := &domain.Contact{}
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.ContactID, &c.CustomName, &c.IsBlocked, &c.CreatedAt); err != nil {
			metrics.ObserveQuery("ContactRepo.ListByOwnerID", start, err)
			return nil, fmt.Errorf("contact_repo: list scan: %w", err)
		}
		items = append(items, c)
	}
	metrics.ObserveQuery("ContactRepo.ListByOwnerID", start, nil)
	return items, nil
}

func (r *WriteRepo) CountByOwnerID(ctx context.Context, ownerID string) (int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM user_contacts WHERE owner_id = $1`, ownerID).Scan(&count)
	metrics.ObserveQuery("ContactRepo.CountByOwnerID", start, err)
	return count, err
}

func (r *WriteRepo) Delete(ctx context.Context, ownerID, contactID string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM user_contacts WHERE owner_id = $1 AND contact_id = $2`, ownerID, contactID)
	metrics.ObserveQuery("ContactRepo.Delete", start, err)
	return err
}

func (r *WriteRepo) IsContact(ctx context.Context, ownerID, contactID string) (bool, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var exists bool
	err := db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_contacts WHERE owner_id = $1 AND contact_id = $2)`,
		ownerID, contactID,
	).Scan(&exists)
	metrics.ObserveQuery("ContactRepo.IsContact", start, err)
	return exists, err
}
