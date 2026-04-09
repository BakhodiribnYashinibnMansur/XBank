package authz

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
