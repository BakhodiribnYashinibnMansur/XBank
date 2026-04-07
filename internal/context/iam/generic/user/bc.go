package user

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/application/command"
	userDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
	Repo    userDomain.Repository // exposed for cross-BC use (session, challenge)
}

func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
		Repo:    repo,
	}
}
