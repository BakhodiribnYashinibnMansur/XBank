package mock

import (
	"context"
	"sync"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

type AuditLog struct {
	mu      sync.RWMutex
	Entries []shared.AuditEntry
}

func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

func (a *AuditLog) Log(ctx context.Context, entry shared.AuditEntry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Entries = append(a.Entries, entry)
	return nil
}
