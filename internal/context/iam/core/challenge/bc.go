package challenge

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge/interfaces/http"
	infraRedis "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/redis"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Challenge BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Challenge BC with all dependencies.
func NewBoundedContext(
	pool *pgxpool.Pool,
	userAuth ports.UserAuthReader,
	cache *infraRedis.ChallengeCache,
) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo, userAuth, cache)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
