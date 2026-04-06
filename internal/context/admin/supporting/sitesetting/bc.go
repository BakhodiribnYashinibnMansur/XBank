// Package sitesetting provides the Site Setting Bounded Context.
// Global key-value configuration managed by admins (maintenance mode, etc.).
package sitesetting

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/interfaces/http"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Site Setting components.
type BoundedContext struct {
	Handler *httpHandler.Handler
}

// NewBoundedContext creates the Site Setting BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool, eventBus appKernel.EventBus) *BoundedContext {
	writeRepo := postgres.NewWriteRepo(pool)
	readRepo := postgres.NewReadRepo(pool)

	upsertHandler := command.NewUpsertHandler(writeRepo, eventBus)
	deleteHandler := command.NewDeleteHandler(writeRepo)
	getHandler := query.NewGetHandler(readRepo)
	listHandler := query.NewListHandler(readRepo)

	handler := httpHandler.NewHandler(upsertHandler, deleteHandler, getHandler, listHandler)

	return &BoundedContext{Handler: handler}
}
