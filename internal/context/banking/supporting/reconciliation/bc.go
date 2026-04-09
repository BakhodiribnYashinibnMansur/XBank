package reconciliation

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/reconciliation/application/command"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/reconciliation/interfaces/http"
)

// BoundedContext wires all Reconciliation BC components.
// Reconciliation has no own infrastructure — it queries account and ledger via ports.
type BoundedContext struct {
	Handler *httpHandler.Handler
	Service *command.Service
}

// NewBoundedContext creates the Reconciliation BC with port dependencies.
func NewBoundedContext(accountReader ports.AccountReader, ledgerReader ports.LedgerReader) *BoundedContext {
	svc := command.NewService(accountReader, ledgerReader)

	return &BoundedContext{
		Handler: httpHandler.NewHandler(svc),
		Service: svc,
	}
}
