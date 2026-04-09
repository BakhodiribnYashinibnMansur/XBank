package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/domain"
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
	ip := admin.Group("/ip-rules")
	ip.Post("", h.Create)
	ip.Get("", h.List)
	ip.Get("/:id", h.GetByID)
	ip.Delete("/:id", h.Delete)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateIPRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	createdBy := c.Locals("user_id").(string)
	rule, err := h.service.Create(c.Context(), req.IPAddress, domain.RuleType(req.RuleType), req.Reason, createdBy, req.ExpiresAt)
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

	var data []IPRuleResponse
	for _, r := range rules {
		data = append(data, toResponse(r))
	}
	if data == nil {
		data = []IPRuleResponse{}
	}

	return apperror.Success(c, http.StatusOK, data)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "IP rule deleted"})
}
