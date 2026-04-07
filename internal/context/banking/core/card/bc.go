package card

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/interfaces/http"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Card BC components.
type BoundedContext struct {
	Handler         *httpHandler.Handler
	ExtendedHandler *httpHandler.ExtendedHandler
	Service         *command.Service
	TokenService    *command.TokenService
	HoldService     *command.HoldService
}

// NewBoundedContext creates the Card BC with all dependencies.
func NewBoundedContext(
	pool *pgxpool.Pool,
	encryptor *crypto.AESEncryptor,
	tokenizer *crypto.Tokenizer,
) *BoundedContext {
	writeRepo := postgres.NewWriteRepo(pool)
	tokenRepo := postgres.NewTokenRepo(pool)
	holdRepo := postgres.NewHoldRepo(pool)

	svc := command.NewService(writeRepo, encryptor)
	tokenSvc := command.NewTokenService(tokenRepo, writeRepo, tokenizer)
	holdSvc := command.NewHoldService(holdRepo, writeRepo)

	return &BoundedContext{
		Handler:         httpHandler.NewHandler(svc),
		ExtendedHandler: httpHandler.NewExtendedHandler(svc, tokenSvc, holdSvc),
		Service:         svc,
		TokenService:    tokenSvc,
		HoldService:     holdSvc,
	}
}
