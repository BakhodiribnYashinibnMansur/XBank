package command

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/systemerror/domain"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ResolveHandler struct {
	pool     *pgxpool.Pool
	eventBus appKernel.EventBus
}

func NewResolveHandler(pool *pgxpool.Pool, bus appKernel.EventBus) *ResolveHandler {
	return &ResolveHandler{pool: pool, eventBus: bus}
}

func (h *ResolveHandler) Handle(ctx context.Context, id, resolvedBy string) error {
	// Check current status
	var resolution string
	err := h.pool.QueryRow(ctx, `SELECT resolution FROM system_errors WHERE id = $1`, id).Scan(&resolution)
	if err != nil {
		return domain.ErrErrorNotFound
	}
	if resolution == string(domain.StatusResolved) {
		return domain.ErrAlreadyResolved
	}

	_, err = h.pool.Exec(ctx,
		`UPDATE system_errors SET resolution = $1, resolved_by = $2, resolved_at = NOW(), updated_at = NOW() WHERE id = $3`,
		domain.StatusResolved, resolvedBy, id)
	if err != nil {
		return fmt.Errorf("resolve error: %w", err)
	}

	h.eventBus.Publish(ctx, domain.NewErrorResolved(id, resolvedBy))
	return nil
}
