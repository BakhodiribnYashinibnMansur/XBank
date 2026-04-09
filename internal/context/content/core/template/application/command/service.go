package command

import (
	"context"
	"time"

	template "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/template/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

// Service provides template management operations.
type Service struct {
	repo template.Repository
}

// NewService creates a new template service.
func NewService(repo template.Repository) *Service {
	return &Service{repo: repo}
}

// Create adds a new template.
func (s *Service) Create(ctx context.Context, slug string, channel template.Channel, subject, body, locale string) (_ *template.Template, err error) {
	defer metrics.ObserveService("TemplateService", "Create", time.Now(), &err)

	tmpl, err := template.NewTemplate(slug, channel, subject, body, locale)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

// GetByID returns a template by its ID.
func (s *Service) GetByID(ctx context.Context, id string) (_ *template.Template, err error) {
	defer metrics.ObserveService("TemplateService", "GetByID", time.Now(), &err)
	return s.repo.GetByID(ctx, id)
}

// GetBySlugAndLocale returns a template by slug and locale.
func (s *Service) GetBySlugAndLocale(ctx context.Context, slug, locale string) (_ *template.Template, err error) {
	defer metrics.ObserveService("TemplateService", "GetBySlugAndLocale", time.Now(), &err)
	return s.repo.GetBySlugAndLocale(ctx, slug, locale)
}

// ListByChannel returns templates filtered by channel with pagination.
func (s *Service) ListByChannel(ctx context.Context, channel string, limit, offset int) (_ []*template.Template, _ int64, err error) {
	defer metrics.ObserveService("TemplateService", "ListByChannel", time.Now(), &err)

	templates, err := s.repo.ListByChannel(ctx, channel, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByChannel(ctx, channel)
	if err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

// UpdateBody updates the template body and subject.
func (s *Service) UpdateBody(ctx context.Context, id, subject, body string) (_ *template.Template, err error) {
	defer metrics.ObserveService("TemplateService", "UpdateBody", time.Now(), &err)

	tmpl, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := tmpl.UpdateBody(subject, body); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Activate publishes a template for use.
func (s *Service) Activate(ctx context.Context, id string) (_ *template.Template, err error) {
	defer metrics.ObserveService("TemplateService", "Activate", time.Now(), &err)

	tmpl, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tmpl.Activate()
	if err := s.repo.Update(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Archive marks a template as archived.
func (s *Service) Archive(ctx context.Context, id string) (_ *template.Template, err error) {
	defer metrics.ObserveService("TemplateService", "Archive", time.Now(), &err)

	tmpl, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tmpl.Archive()
	if err := s.repo.Update(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}
