package http

import (
	"net/http"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/audit/application/command"
	auditDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/audit/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/httpx"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListAuditLogs(c *fiber.Ctx) error {
	pg := httpx.ParsePagination(c)

	filter := auditDomain.AuditFilter{
		AggregateType: c.Query("aggregate_type"),
		AggregateID:   c.Query("aggregate_id"),
		Action:        c.Query("action"),
		ActorID:       c.Query("actor_id"),
		Limit:         pg.Limit,
		Offset:        pg.Offset(),
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

	logs, total, err := h.service.ListAuditLogs(c.Context(), filter)
	if err != nil {
		return err
	}

	resp := make([]AuditLogResponse, len(logs))
	for i, l := range logs {
		resp[i] = AuditLogResponse{
			ID: l.ID, AggregateType: l.AggregateType, AggregateID: l.AggregateID,
			Action: l.Action, ActorID: l.ActorID, Attributes: l.Attributes,
			IPAddress: l.IPAddress, UserAgent: l.UserAgent, CreatedAt: l.CreatedAt,
		}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data: resp,
		Pagination: httpx.PaginationResponse{
			Page:  pg.Page,
			Limit: pg.Limit,
			Total: total,
		},
	})
}

func (h *Handler) ListEndpointHistory(c *fiber.Ctx) error {
	pg := httpx.ParsePagination(c)

	filter := auditDomain.EndpointFilter{
		Method: c.Query("method"),
		Path:   c.Query("path"),
		UserID: c.Query("user_id"),
		Limit:  pg.Limit,
		Offset: pg.Offset(),
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

	entries, total, err := h.service.ListEndpointHistory(c.Context(), filter)
	if err != nil {
		return err
	}

	resp := make([]EndpointHistoryResponse, len(entries))
	for i, e := range entries {
		resp[i] = EndpointHistoryResponse{
			ID: e.ID, Method: e.Method, Path: e.Path, StatusCode: e.StatusCode,
			UserID: e.UserID, IPAddress: e.IPAddress, DurationMs: e.DurationMs, CreatedAt: e.CreatedAt,
		}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data: resp,
		Pagination: httpx.PaginationResponse{
			Page:  pg.Page,
			Limit: pg.Limit,
			Total: total,
		},
	})
}
