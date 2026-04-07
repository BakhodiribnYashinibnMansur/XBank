package ledger

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Ledger BC components.
// Ledger is an internal-only BC (no HTTP handler).
type BoundedContext struct {
	Repo domain.Repository
}

// NewBoundedContext creates the Ledger BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	return &BoundedContext{
		Repo: postgres.NewWriteRepo(pool),
	}
}
