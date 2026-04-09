package http

import (
	"encoding/json"
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/healthcheck/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/httpx"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for health monitoring.
type Handler struct {
	service *command.Service
}

// NewHandler creates a new healthcheck HTTP handler.
func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

// Check godoc
// @Summary      Run system health check
// @Tags         Health
// @Produce      json
// @Success      200 {object} SystemHealthResponse
// @Failure      500 {object} apperror.ErrorResponse
// @Router       /health/check [post]
func (h *Handler) Check(c *fiber.Ctx) error {
	health, err := h.service.RunCheck(c.Context())
	if err != nil {
		return err
	}

	var components []ComponentCheckResponse
	for _, comp := range health.Components {
		components = append(components, ComponentCheckResponse{
			Name:      comp.Name,
			Status:    string(comp.Status),
			LatencyMs: comp.Latency.Milliseconds(),
			Message:   comp.Message,
			CheckedAt: comp.CheckedAt,
		})
	}
	if components == nil {
		components = []ComponentCheckResponse{}
	}

	return apperror.Success(c, http.StatusOK, SystemHealthResponse{
		Status:     string(health.Status),
		Components: components,
		CheckedAt:  health.CheckedAt,
	})
}

// Latest godoc
// @Summary      Get latest health check result
// @Tags         Health
// @Produce      json
// @Success      200 {object} HealthRecordResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Router       /health/latest [get]
func (h *Handler) Latest(c *fiber.Ctx) error {
	record, err := h.service.GetLatest(c.Context())
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toRecordResponse(record))
}

// History godoc
// @Summary      List health check history
// @Tags         Health
// @Produce      json
// @Param        page  query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Success      200 {object} httpx.PaginatedResponse
// @Router       /health/history [get]
func (h *Handler) History(c *fiber.Ctx) error {
	pg := httpx.ParsePagination(c)

	records, total, err := h.service.ListHistory(c.Context(), pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []HealthRecordResponse
	for _, r := range records {
		data = append(data, toRecordResponse(r))
	}
	if data == nil {
		data = []HealthRecordResponse{}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data:       data,
		Pagination: httpx.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

func toRecordResponse(r *domain.HealthRecord) HealthRecordResponse {
	var components []ComponentCheckResponse
	_ = json.Unmarshal([]byte(r.Components), &components)
	if components == nil {
		components = []ComponentCheckResponse{}
	}
	return HealthRecordResponse{
		ID:         r.ID,
		Status:     string(r.Status),
		Components: components,
		CheckedAt:  r.CheckedAt,
	}
}
