package session

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session/interfaces/http"
	infraRedis "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/redis"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Session BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Session BC with all dependencies.
func NewBoundedContext(
	pool *pgxpool.Pool,
	userAuth ports.UserAuthReader,
	jwtService *infraAuth.JWTService,
	totpService *infraAuth.TOTPService,
	sessionCache *infraRedis.SessionCache,
	loginLimiter *infraRedis.LoginLimiter,
) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(userAuth, repo, jwtService, totpService, sessionCache, loginLimiter)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
