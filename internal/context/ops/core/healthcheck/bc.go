package healthcheck

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Healthcheck BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Healthcheck BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool, checkers ...domain.HealthChecker) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo, checkers...)
	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
