package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit/application/command"
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
	rl := admin.Group("/rate-limits")
	rl.Post("", h.Create)
	rl.Get("", h.List)
	rl.Get("/:id", h.GetByID)
	rl.Put("/:id", h.Update)
	rl.Delete("/:id", h.Delete)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRateLimitRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	rule, err := h.service.Create(c.Context(), req.Key, req.MaxRequests, req.WindowSeconds, req.Description, req.Enabled)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toResponse(rule))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	rule, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toResponse(rule))
}

func (h *Handler) List(c *fiber.Ctx) error {
	rules, err := h.service.ListAll(c.Context())
	if err != nil {
		return err
	}

	var data []RateLimitResponse
	for _, r := range rules {
		data = append(data, toResponse(r))
	}
	if data == nil {
		data = []RateLimitResponse{}
	}

	return apperror.Success(c, http.StatusOK, data)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateRateLimitRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	rule, err := h.service.Update(c.Context(), id, req.MaxRequests, req.WindowSeconds, req.Description, req.Enabled)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toResponse(rule))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Rate limit rule deleted"})
}
