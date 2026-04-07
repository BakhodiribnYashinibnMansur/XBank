package http

import (
	"net/http"

	contactApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/contact"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type ContactHandler struct {
	service *contactApp.Service
}

func NewContactHandler(service *contactApp.Service) *ContactHandler {
	return &ContactHandler{service: service}
}

// Add godoc
// @Summary      Add a contact
// @Tags         Contacts
// @Accept       json
// @Produce      json
// @Param        body body object true "contact_id and custom_name"
// @Success      201 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /contacts/add [post]
func (h *ContactHandler) Add(c *fiber.Ctx) error {
	var req struct {
		ContactID  string `json:"contact_id"`
		CustomName string `json:"custom_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.ContactID == "" {
		return apperror.ErrMissingField.WithMessage("contact_id is required")
	}

	ownerID := c.Locals("user_id").(string)

	contact, err := h.service.Add(c.Context(), ownerID, req.ContactID, req.CustomName)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, fiber.Map{
		"id":          contact.ID,
		"contact_id":  contact.ContactID,
		"custom_name": contact.CustomName,
		"created_at":  contact.CreatedAt,
	})
}

// List godoc
// @Summary      List contacts
// @Tags         Contacts
// @Produce      json
// @Param        page  query int false "Page" default(1)
// @Param        limit query int false "Limit" default(20)
// @Success      200 {object} dto.PaginatedResponse
// @Security     BearerAuth
// @Router       /contacts/list [get]
func (h *ContactHandler) List(c *fiber.Ctx) error {
	ownerID := c.Locals("user_id").(string)
	pg := dto.ParsePagination(c)

	contacts, total, err := h.service.List(c.Context(), ownerID, pg.Limit, pg.Offset())
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

	return apperror.Success(c, http.StatusOK, dto.PaginatedResponse{
		Data:       data,
		Pagination: dto.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

// Delete godoc
// @Summary      Delete a contact
// @Tags         Contacts
// @Param        contact_id query string true "Contact user ID"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /contacts/delete [delete]
func (h *ContactHandler) Delete(c *fiber.Ctx) error {
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
