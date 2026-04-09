package kyc

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/interfaces/http"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all KYC BC components.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the KYC BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool, publisher domain.EventPublisher, topics config.KafkaTopicsConfig) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	svc := command.NewService(repo, publisher, topics)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
