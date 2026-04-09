package account

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/application/command"
	accountDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/interfaces/http"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Account BC components.
type BoundedContext struct {
	Handler      *httpHandler.Handler
	Service      *command.Service
	Repo         accountDomain.Repository // internal use within Account BC
	TransferPort ports.AccountTransferPort // cross-BC port for Transfer BC
	Reader       ports.AccountReader       // cross-BC port for Reconciliation BC
}

// NewBoundedContext creates the Account BC with all dependencies.
func NewBoundedContext(
	pool *pgxpool.Pool,
	txManager domain.TxManager,
	publisher domain.EventPublisher,
	topics config.KafkaTopicsConfig,
	auditLog domain.AuditLog,
) *BoundedContext {
	repo := postgres.NewWriteRepo(pool)
	eventRepo := postgres.NewEventRepo(pool)
	svc := command.NewService(repo, eventRepo, txManager, publisher, topics, auditLog)

	return &BoundedContext{
		Handler:      httpHandler.NewHandler(svc),
		Service:      svc,
		Repo:         repo,
		TransferPort: postgres.NewTransferPortAdapter(repo),
		Reader:       postgres.NewReaderPortAdapter(pool),
	}
}
