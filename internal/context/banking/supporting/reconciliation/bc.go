package reconciliation

import (
	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	ledger "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/reconciliation/application/command"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/reconciliation/interfaces/http"
)

// BoundedContext wires all Reconciliation BC components.
// Reconciliation has no own infrastructure — it queries account and ledger repos directly.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Reconciliation BC with external repo dependencies.
func NewBoundedContext(accountRepo account.Repository, ledgerRepo ledger.Repository) *BoundedContext {
	svc := command.NewService(accountRepo, ledgerRepo)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
