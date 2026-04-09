package dataexport

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Data Export BC components.
type BoundedContext struct {
	Handler        *httpHandler.Handler
	RequestHandler *command.RequestHandler
	ProcessHandler *command.ProcessHandler
}

// NewBoundedContext creates the Data Export BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	writeRepo := postgres.NewWriteRepo(pool)
	readRepo := postgres.NewReadRepo(pool)

	requestHandler := command.NewRequestHandler(writeRepo)
	processHandler := command.NewProcessHandler(writeRepo)
	getHandler := query.NewGetHandler(readRepo)
	listHandler := query.NewListHandler(readRepo)

	return &BoundedContext{
		Handler:        httpHandler.NewHandler(requestHandler, processHandler, getHandler, listHandler),
		RequestHandler: requestHandler,
		ProcessHandler: processHandler,
	}
}
