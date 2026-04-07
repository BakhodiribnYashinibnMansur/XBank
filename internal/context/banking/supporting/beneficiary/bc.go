package beneficiary

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/beneficiary/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/beneficiary/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/beneficiary/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Beneficiary BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Beneficiary BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)
	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
