package translation

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/interfaces/http"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BoundedContext struct {
	Handler *httpHandler.Handler
}

func NewBoundedContext(pool *pgxpool.Pool, eventBus appKernel.EventBus) *BoundedContext {
	writeRepo := postgres.NewWriteRepo(pool)
	readRepo := postgres.NewReadRepo(pool)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(
			command.NewCreateHandler(writeRepo),
			command.NewUpdateHandler(writeRepo, eventBus),
			command.NewDeleteHandler(writeRepo),
			query.NewGetHandler(readRepo),
			query.NewListHandler(readRepo),
		),
	}
}
