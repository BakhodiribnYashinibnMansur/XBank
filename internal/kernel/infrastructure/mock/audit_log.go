package mock

import (
	"context"
	"sync"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

type AuditLog struct {
	mu      sync.RWMutex
	Entries []domain.AuditEntry
}

func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

func (a *AuditLog) Log(ctx context.Context, entry domain.AuditEntry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Entries = append(a.Entries, entry)
	return nil
}
