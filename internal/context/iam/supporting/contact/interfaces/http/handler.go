package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/contact/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Add(c *fiber.Ctx) error {
	var req AddContactRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.ContactID == "" {
		return apperror.ErrMissingField.WithMessage("contact_id is required")
	}

	ownerID := c.Locals("user_id").(string)

	result, err := h.service.Add(c.Context(), ownerID, req.ContactID, req.CustomName)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, fiber.Map{
		"id":          result.ID,
		"contact_id":  result.ContactID,
		"custom_name": result.CustomName,
		"created_at":  result.CreatedAt,
	})
}

func (h *Handler) List(c *fiber.Ctx) error {
	ownerID := c.Locals("user_id").(string)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	offset := (page - 1) * limit

	contacts, total, err := h.service.List(c.Context(), ownerID, limit, offset)
	if err != nil {
		return err
	}

	var data []fiber.Map
	for _, ct := range contacts {
		data = append(data, fiber.Map{
			"id":          ct.ID,
			"contact_id":  ct.ContactID,
			"custom_name": ct.CustomName,
			"created_at":  ct.CreatedAt,
		})
	}
	if data == nil {
		data = []fiber.Map{}
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{
		"data": data,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	contactID := c.Query("contact_id")
	if contactID == "" {
		return apperror.ErrMissingField.WithMessage("contact_id is required")
	}

	ownerID := c.Locals("user_id").(string)

	if err := h.service.Delete(c.Context(), ownerID, contactID); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Contact deleted"})
}
