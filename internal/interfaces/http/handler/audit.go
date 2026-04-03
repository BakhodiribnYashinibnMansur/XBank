package handler

import (
	"net/http"
	"time"

	infraMongo "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/mongodb"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type AuditHandler struct {
	reader *infraMongo.AuditReader
}

func NewAuditHandler(reader *infraMongo.AuditReader) *AuditHandler {
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
// @Success      200 {object} dto.PaginatedResponse
// @Security     BearerAuth
// @Router       /admin/audit [get]
func (h *AuditHandler) List(c *fiber.Ctx) error {
	if h.reader == nil {
		return apperror.ErrInternal.WithMessage("Audit logging not configured")
	}

	pg := dto.ParsePagination(c)
	if pg.Limit > 100 {
		pg.Limit = 100
	}

	filter := infraMongo.AuditFilter{
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

	return apperror.Success(c, http.StatusOK, dto.PaginatedResponse{
		Data:       data,
		Pagination: dto.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}
