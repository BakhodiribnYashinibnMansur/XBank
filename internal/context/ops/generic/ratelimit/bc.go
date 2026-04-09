package ratelimit

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Rate Limit BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Rate Limit BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)
	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
