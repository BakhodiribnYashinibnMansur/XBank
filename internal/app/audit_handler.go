package app

import (
	"net/http"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/mongodb"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/httpx"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type AuditHandler struct {
	reader *mongodb.AuditReader
}

func NewAuditHandler(reader *mongodb.AuditReader) *AuditHandler {
	return &AuditHandler{reader: reader}
}

// List godoc
// @Summary      List audit log entries (admin only)
// @Tags         Admin - Audit
// @Produce      json
// @Param        aggregate_type query string false "Filter by type (Account, Transfer, Card)"
// @Param        aggregate_id   query string false "Filter by entity ID"
// @Param        action         query string false "Filter by action"
// @Param        actor_id       query string false "Filter by actor user ID"
// @Param        from           query string false "From timestamp (RFC3339)"
// @Param        to             query string false "To timestamp (RFC3339)"
// @Param        page           query int    false "Page number" default(1)
// @Param        limit          query int    false "Items per page" default(50)
// @Success      200 {object} httpx.PaginatedResponse
// @Security     BearerAuth
// @Router       /admin/audit [get]
func (h *AuditHandler) List(c *fiber.Ctx) error {
	if h.reader == nil {
		return apperror.ErrInternal.WithMessage("Audit logging not configured")
	}

	pg := httpx.ParsePagination(c)
	if pg.Limit > 100 {
		pg.Limit = 100
	}

	filter := mongodb.AuditFilter{
		AggregateType: c.Query("aggregate_type"),
		AggregateID:   c.Query("aggregate_id"),
		Action:        c.Query("action"),
		ActorID:       c.Query("actor_id"),
		Limit:         int64(pg.Limit),
		Offset:        int64(pg.Offset()),
	}

	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.From = &t
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.To = &t
		}
	}

	entries, total, err := h.reader.List(c.Context(), filter)
	if err != nil {
		return apperror.ErrInternal.Wrap(err)
	}

	var data []fiber.Map
	for _, e := range entries {
		data = append(data, fiber.Map{
			"aggregate_type": e.AggregateType,
			"aggregate_id":   e.AggregateID,
			"action":         e.Action,
			"actor_id":       e.ActorID,
			"attributes":     e.Attributes,
			"ip_address":     e.IPAddress,
			"user_agent":     e.UserAgent,
			"timestamp":      e.Timestamp,
		})
	}
	if data == nil {
		data = []fiber.Map{}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data:       data,
		Pagination: httpx.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}
