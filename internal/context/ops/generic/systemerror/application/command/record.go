package command

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/systemerror/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/systemerror/domain"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecordHandler struct {
	pool     *pgxpool.Pool
	eventBus appKernel.EventBus
}

func NewRecordHandler(pool *pgxpool.Pool, bus appKernel.EventBus) *RecordHandler {
	return &RecordHandler{pool: pool, eventBus: bus}
}

func (h *RecordHandler) Handle(ctx context.Context, req application.RecordErrorRequest) (string, error) {
	e, err := domain.NewSystemError(req.Code, req.Message, req.Severity, req.Category)
	if err != nil {
		return "", fmt.Errorf("record error: %w", err)
	}
	e.WithContext(req.RequestID, req.UserID, req.IPAddress, req.Path, req.Method, req.StackTrace, req.Metadata)

	metaJSON, _ := json.Marshal(e.Metadata)
	err = h.pool.QueryRow(ctx,
		`INSERT INTO system_errors (code, message, severity, category, stack_trace, request_id, user_id, ip_address, path, method, metadata, resolution, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id`,
		e.Code, e.Message, e.Severity, e.Category, e.StackTrace, e.RequestID, e.UserID,
		e.IPAddress, e.Path, e.Method, metaJSON, e.Resolution, e.CreatedAt, e.UpdatedAt,
	).Scan(&e.ID)
	if err != nil {
		return "", fmt.Errorf("record error: save: %w", err)
	}

	h.eventBus.Publish(ctx, domain.NewErrorRecorded(e.ID, e.Code, e.Severity))
	return e.ID, nil
}
