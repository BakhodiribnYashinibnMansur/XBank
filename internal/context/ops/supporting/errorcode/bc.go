package errorcode

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BoundedContext struct {
	Handler *httpHandler.Handler
}

func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	writeRepo := postgres.NewWriteRepo(pool)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(
			command.NewCreateHandler(writeRepo),
			command.NewUpdateHandler(writeRepo),
			command.NewDeleteHandler(writeRepo),
			query.NewListHandler(pool),
			query.NewLookupHandler(pool),
		),
	}
}
