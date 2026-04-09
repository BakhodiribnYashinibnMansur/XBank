package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

var (
	ErrTemplateNotFound = domain.NewDomainError("TEMPLATE_NOT_FOUND", "template not found")
	ErrTemplateExists   = domain.NewDomainError("TEMPLATE_EXISTS", "template with this slug already exists")
	ErrMissingSlug      = domain.NewDomainError("MISSING_FIELD", "slug cannot be empty")
	ErrMissingBody      = domain.NewDomainError("MISSING_FIELD", "body cannot be empty")
)

// Channel represents the delivery channel for a template.
type Channel string

const (
	ChannelEmail Channel = "EMAIL"
	ChannelSMS   Channel = "SMS"
	ChannelPush  Channel = "PUSH"
)

// Status represents the template status.
type Status string

const (
	StatusDraft    Status = "DRAFT"
	StatusActive   Status = "ACTIVE"
	StatusArchived Status = "ARCHIVED"
)

// Template represents a message or email template.
type Template struct {
	ID        string
	Slug      string // unique identifier, e.g. "welcome_email", "otp_sms"
	Channel   Channel
	Subject   string // for EMAIL channel
	Body      string // template body with placeholders, e.g. "Hello {{.Name}}"
	Locale    string // e.g. "en", "uz", "ru"
	Status    Status
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTemplate creates a new template with validation.
func NewTemplate(slug string, channel Channel, subject, body, locale string) (*Template, error) {
	if slug == "" {
		return nil, ErrMissingSlug
	}
	if body == "" {
		return nil, ErrMissingBody
	}

	return &Template{
		Slug:    slug,
		Channel: channel,
		Subject: subject,
		Body:    body,
		Locale:  locale,
		Status:  StatusDraft,
		Version: 1,
	}, nil
}

// Activate publishes the template for use.
func (t *Template) Activate() {
	t.Status = StatusActive
}

// Archive marks the template as archived.
func (t *Template) Archive() {
	t.Status = StatusArchived
}

// UpdateBody updates the template body and increments the version.
func (t *Template) UpdateBody(subject, body string) error {
	if body == "" {
		return ErrMissingBody
	}
	t.Subject = subject
	t.Body = body
	t.Version++
	return nil
}

// Repository defines the persistence interface for templates.
type Repository interface {
	Create(ctx context.Context, template *Template) error
	GetByID(ctx context.Context, id string) (*Template, error)
	GetBySlugAndLocale(ctx context.Context, slug, locale string) (*Template, error)
	ListByChannel(ctx context.Context, channel string, limit, offset int) ([]*Template, error)
	CountByChannel(ctx context.Context, channel string) (int64, error)
	Update(ctx context.Context, template *Template) error
}
