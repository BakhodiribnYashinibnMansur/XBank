package usersetting

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all User Setting BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
}

// NewBoundedContext creates the User Setting BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	writeRepo := postgres.NewWriteRepo(pool)
	readRepo := postgres.NewReadRepo(pool)

	upsertHandler := command.NewUpsertHandler(writeRepo)
	deleteHandler := command.NewDeleteHandler(writeRepo)
	listHandler := query.NewListHandler(readRepo)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(upsertHandler, deleteHandler, listHandler),
	}
}
