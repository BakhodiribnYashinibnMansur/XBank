package dashboard

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/dashboard/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/dashboard/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/dashboard/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Dashboard BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Dashboard BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)
	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
