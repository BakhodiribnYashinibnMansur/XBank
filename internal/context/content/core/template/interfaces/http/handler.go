package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/template/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/template/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/httpx"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for template management.
type Handler struct {
	service *command.Service
}

// NewHandler creates a new template HTTP handler.
func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary      Create a new template
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        body body CreateTemplateRequest true "Template details"
// @Success      201 {object} TemplateResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /templates/create [post]
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Slug == "" {
		return apperror.ErrMissingField.WithMessage("slug is required")
	}
	if req.Body == "" {
		return apperror.ErrMissingField.WithMessage("body is required")
	}

	channel := domain.Channel(req.Channel)
	if channel != domain.ChannelEmail && channel != domain.ChannelSMS && channel != domain.ChannelPush {
		return apperror.ErrValidation.WithMessage("channel must be EMAIL, SMS, or PUSH")
	}

	tmpl, err := h.service.Create(c.Context(), req.Slug, channel, req.Subject, req.Body, req.Locale)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toTemplateResponse(tmpl))
}

// GetByID godoc
// @Summary      Get template by ID
// @Tags         Templates
// @Produce      json
// @Param        id query string true "Template ID"
// @Success      200 {object} TemplateResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /templates/get [get]
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	tmpl, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toTemplateResponse(tmpl))
}

// Resolve godoc
// @Summary      Resolve a template by slug and locale
// @Tags         Templates
// @Produce      json
// @Param        slug   query string true "Template slug"
// @Param        locale query string true "Locale (e.g. en, uz, ru)"
// @Success      200 {object} TemplateResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /templates/resolve [get]
func (h *Handler) Resolve(c *fiber.Ctx) error {
	slug := c.Query("slug")
	locale := c.Query("locale")
	if slug == "" || locale == "" {
		return apperror.ErrMissingField.WithMessage("slug and locale query parameters are required")
	}

	tmpl, err := h.service.GetBySlugAndLocale(c.Context(), slug, locale)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toTemplateResponse(tmpl))
}

// List godoc
// @Summary      List templates (optionally filter by channel)
// @Tags         Templates
// @Produce      json
// @Param        channel query string false "Filter by channel (EMAIL, SMS, PUSH)"
// @Param        page    query int    false "Page number" default(1)
// @Param        limit   query int    false "Items per page" default(20)
// @Success      200 {object} httpx.PaginatedResponse
// @Security     BearerAuth
// @Router       /templates/list [get]
func (h *Handler) List(c *fiber.Ctx) error {
	channel := c.Query("channel")
	pg := httpx.ParsePagination(c)

	templates, total, err := h.service.ListByChannel(c.Context(), channel, pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []TemplateResponse
	for _, tmpl := range templates {
		data = append(data, toTemplateResponse(tmpl))
	}
	if data == nil {
		data = []TemplateResponse{}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data:       data,
		Pagination: httpx.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

// UpdateBody godoc
// @Summary      Update template body
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        body body UpdateBodyRequest true "Template ID, subject, and body"
// @Success      200 {object} TemplateResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /templates/update-body [put]
func (h *Handler) UpdateBody(c *fiber.Ctx) error {
	var req UpdateBodyRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.ID == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	tmpl, err := h.service.UpdateBody(c.Context(), req.ID, req.Subject, req.Body)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toTemplateResponse(tmpl))
}

// Activate godoc
// @Summary      Activate a template
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        body body IDRequest true "Template ID"
// @Success      200 {object} TemplateResponse
// @Security     BearerAuth
// @Router       /templates/activate [post]
func (h *Handler) Activate(c *fiber.Ctx) error {
	var req IDRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.ID == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	tmpl, err := h.service.Activate(c.Context(), req.ID)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toTemplateResponse(tmpl))
}

// Archive godoc
// @Summary      Archive a template
// @Tags         Templates
// @Accept       json
// @Produce      json
// @Param        body body IDRequest true "Template ID"
// @Success      200 {object} TemplateResponse
// @Security     BearerAuth
// @Router       /templates/archive [post]
func (h *Handler) Archive(c *fiber.Ctx) error {
	var req IDRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.ID == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	tmpl, err := h.service.Archive(c.Context(), req.ID)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toTemplateResponse(tmpl))
}

func toTemplateResponse(t *domain.Template) TemplateResponse {
	return TemplateResponse{
		ID:        t.ID,
		Slug:      t.Slug,
		Channel:   string(t.Channel),
		Subject:   t.Subject,
		Body:      t.Body,
		Locale:    t.Locale,
		Status:    string(t.Status),
		Version:   t.Version,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
