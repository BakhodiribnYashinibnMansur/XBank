package iprule

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all IP Rule BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the IP Rule BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)
	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
