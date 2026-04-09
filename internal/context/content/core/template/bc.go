package template

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/template/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/template/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/template/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Template BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Template BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)
	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
