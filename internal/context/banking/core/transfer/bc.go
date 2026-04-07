package transfer

import (
	accountDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/interfaces/http"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Transfer BC components.
type BoundedContext struct {
	Handler          *httpHandler.Handler
	ScheduledHandler *httpHandler.ScheduledHandler
	Service          *command.Service
	ScheduledService *command.ScheduledService
}

// NewBoundedContext creates the Transfer BC with all dependencies.
func NewBoundedContext(
	pool *pgxpool.Pool,
	accountRepo accountDomain.Repository,
	txManager domain.TxManager,
	publisher domain.EventPublisher,
	topics config.KafkaTopicsConfig,
) *BoundedContext {
	writeRepo := postgres.NewWriteRepo(pool)
	eventRepo := postgres.NewEventRepo(pool)
	scheduledRepo := postgres.NewScheduledRepo(pool)

	svc := command.NewService(writeRepo, eventRepo, accountRepo, txManager, publisher, topics)
	schedSvc := command.NewScheduledService(scheduledRepo, svc)

	return &BoundedContext{
		Handler:          httpHandler.NewHandler(svc),
		ScheduledHandler: httpHandler.NewScheduledHandler(schedSvc),
		Service:          svc,
		ScheduledService: schedSvc,
	}
}
