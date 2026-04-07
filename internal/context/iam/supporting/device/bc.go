package device

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Device BC components.
type BoundedContext struct {
	Repo *postgres.WriteRepo
}

// NewBoundedContext creates the Device BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	return &BoundedContext{
		Repo: postgres.NewWriteRepo(pool),
	}
}
