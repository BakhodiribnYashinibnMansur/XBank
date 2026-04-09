package file

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all File Management BC components.
type BoundedContext struct {
	Handler       *httpHandler.Handler
	UploadHandler *command.UploadHandler
}

// NewBoundedContext creates the File BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	writeRepo := postgres.NewWriteRepo(pool)
	readRepo := postgres.NewReadRepo(pool)

	uploadHandler := command.NewUploadHandler(writeRepo)
	deleteHandler := command.NewDeleteHandler(writeRepo)
	getHandler := query.NewGetHandler(readRepo)
	listHandler := query.NewListHandler(readRepo)

	return &BoundedContext{
		Handler:       httpHandler.NewHandler(uploadHandler, deleteHandler, getHandler, listHandler),
		UploadHandler: uploadHandler,
	}
}
