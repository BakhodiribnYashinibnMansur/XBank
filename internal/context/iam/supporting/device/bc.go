package device

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Device BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Device BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)
	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
