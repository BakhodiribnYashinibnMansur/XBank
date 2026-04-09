package user

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/application/command"
	userDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BoundedContext struct {
	Handler  *httpHandler.Handler
	Service  *command.Service
	Repo     userDomain.Repository  // internal use within User BC
	AuthPort ports.UserAuthReader   // cross-BC port for Session/Challenge BCs
}

func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo)

	return &BoundedContext{
		Handler:  httpHandler.NewHandler(svc),
		Service:  svc,
		Repo:     repo,
		AuthPort: postgres.NewAuthPortAdapter(pool),
	}
}
