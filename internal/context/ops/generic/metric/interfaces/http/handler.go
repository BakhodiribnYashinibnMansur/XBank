package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/metric/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(admin fiber.Router) {
	m := admin.Group("/metrics/app")
	m.Get("", h.ListRecent)
	m.Get("/:name", h.GetByName)
}

func (h *Handler) ListRecent(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	items, err := h.service.ListRecent(c.Context(), limit)
	if err != nil {
		return err
	}

	var data []MetricResponse
	for _, m := range items {
		data = append(data, toResponse(m))
	}
	if data == nil {
		data = []MetricResponse{}
	}

	return apperror.Success(c, http.StatusOK, data)
}

func (h *Handler) GetByName(c *fiber.Ctx) error {
	name := c.Params("name")
	items, err := h.service.FindByName(c.Context(), name)
	if err != nil {
		return err
	}

	var data []MetricResponse
	for _, m := range items {
		data = append(data, toResponse(m))
	}
	if data == nil {
		data = []MetricResponse{}
	}

	return apperror.Success(c, http.StatusOK, data)
}
