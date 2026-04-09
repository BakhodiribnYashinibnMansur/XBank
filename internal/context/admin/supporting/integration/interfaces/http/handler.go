package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/integration/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/integration/domain"
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
	ig := admin.Group("/integrations")
	ig.Post("", h.Create)
	ig.Get("", h.List)
	ig.Get("/:id", h.GetByID)
	ig.Put("/:id", h.Update)
	ig.Delete("/:id", h.Delete)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateIntegrationRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	i, err := h.service.Create(c.Context(), req.Name, req.BaseURL, req.APIKey, domain.Status(req.Status), req.WebhookURL)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toResponse(i))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	i, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toResponse(i))
}

func (h *Handler) List(c *fiber.Ctx) error {
	items, err := h.service.ListAll(c.Context())
	if err != nil {
		return err
	}

	var data []IntegrationResponse
	for _, i := range items {
		data = append(data, toResponse(i))
	}
	if data == nil {
		data = []IntegrationResponse{}
	}

	return apperror.Success(c, http.StatusOK, data)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateIntegrationRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	i, err := h.service.Update(c.Context(), id, req.BaseURL, req.APIKey, domain.Status(req.Status), req.WebhookURL)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toResponse(i))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Integration deleted"})
}
