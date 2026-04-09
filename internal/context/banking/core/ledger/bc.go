package ledger

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Ledger BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Repo    domain.Repository
	Reader  ports.LedgerReader // cross-BC port for Reconciliation BC
}

// NewBoundedContext creates the Ledger BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	listHandler := query.NewListHandler(repo)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(listHandler),
		Repo:    repo,
		Reader:  postgres.NewReaderPortAdapter(pool),
	}
}
