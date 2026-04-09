package currency

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/currency/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/currency/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/currency/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Currency BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Currency BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)
	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
