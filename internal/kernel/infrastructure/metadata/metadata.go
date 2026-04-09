package metadata

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Entry represents a single EAV (Entity-Attribute-Value) metadata record.
type Entry struct {
	EntityType string `json:"entity_type"` // e.g. "account", "user", "card"
	EntityID   string `json:"entity_id"`
	Key        string `json:"key"`
	Value      string `json:"value"`
}

// Repository manages entity metadata stored as EAV in PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a metadata repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Set upserts a metadata entry (insert or update on conflict).
func (r *Repository) Set(ctx context.Context, entityType, entityID, key, value string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO entity_metadata (entity_type, entity_id, key, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (entity_type, entity_id, key)
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, entityType, entityID, key, value)
	return err
}

// Get retrieves a single metadata value.
func (r *Repository) Get(ctx context.Context, entityType, entityID, key string) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx, `
		SELECT value FROM entity_metadata
		WHERE entity_type = $1 AND entity_id = $2 AND key = $3
	`, entityType, entityID, key).Scan(&value)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return value, err
}

// GetAll retrieves all metadata entries for a given entity.
func (r *Repository) GetAll(ctx context.Context, entityType, entityID string) ([]Entry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT entity_type, entity_id, key, value FROM entity_metadata
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY key
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.EntityType, &e.EntityID, &e.Key, &e.Value); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetAsMap returns all metadata for an entity as a key-value map.
func (r *Repository) GetAsMap(ctx context.Context, entityType, entityID string) (map[string]string, error) {
	entries, err := r.GetAll(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		m[e.Key] = e.Value
	}
	return m, nil
}

// SetJSON marshals a value to JSON and stores it as metadata.
func (r *Repository) SetJSON(ctx context.Context, entityType, entityID, key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.Set(ctx, entityType, entityID, key, string(data))
}

// Delete removes a metadata entry.
func (r *Repository) Delete(ctx context.Context, entityType, entityID, key string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM entity_metadata
		WHERE entity_type = $1 AND entity_id = $2 AND key = $3
	`, entityType, entityID, key)
	return err
}

// DeleteAll removes all metadata for a given entity.
func (r *Repository) DeleteAll(ctx context.Context, entityType, entityID string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM entity_metadata
		WHERE entity_type = $1 AND entity_id = $2
	`, entityType, entityID)
	return err
}
