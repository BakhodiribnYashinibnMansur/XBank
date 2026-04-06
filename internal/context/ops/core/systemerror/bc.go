package systemerror

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/systemerror/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/systemerror/application/query"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/systemerror/interfaces/http"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BoundedContext struct {
	Handler       *httpHandler.Handler
	RecordHandler *command.RecordHandler // exposed for middleware integration
}

func NewBoundedContext(pool *pgxpool.Pool, eventBus appKernel.EventBus) *BoundedContext {
	recordHandler := command.NewRecordHandler(pool, eventBus)
	resolveHandler := command.NewResolveHandler(pool, eventBus)
	listHandler := query.NewListHandler(pool)

	return &BoundedContext{
		Handler:       httpHandler.NewHandler(resolveHandler, listHandler),
		RecordHandler: recordHandler,
	}
}
