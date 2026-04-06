package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain/entity"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for site settings.
type Handler struct {
	upsert *command.UpsertHandler
	delete *command.DeleteHandler
	get    *query.GetHandler
	list   *query.ListHandler
}

// NewHandler creates a new site setting HTTP handler.
func NewHandler(
	upsert *command.UpsertHandler,
	del *command.DeleteHandler,
	get *query.GetHandler,
	list *query.ListHandler,
) *Handler {
	return &Handler{upsert: upsert, delete: del, get: get, list: list}
}

// Upsert creates or updates a site setting.
func (h *Handler) Upsert(c *fiber.Ctx) error {
	var req UpsertSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	result, err := h.upsert.Handle(c.Context(), application.CreateSettingRequest{
		Key:         req.Key,
		Value:       req.Value,
		SettingType: entity.SettingType(req.SettingType),
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, result)
}

// Delete removes a site setting by ID.
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	if err := h.delete.Handle(c.Context(), id); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"deleted": true})
}

// GetByID returns a single site setting.
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	result, err := h.get.Handle(c.Context(), id)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, result)
}

// List returns paginated site settings.
func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	result, err := h.list.Handle(c.Context(), repository.SiteSettingFilter{
		Key:         c.Query("key"),
		SettingType: c.Query("setting_type"),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, result)
}
