package domain

import (
	"context"
	"time"

)

type WriteRepository interface {
	Save(ctx context.Context, a *Announcement) error
	Update(ctx context.Context, a *Announcement) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Announcement, error)
}

type AnnouncementView struct {
	ID        string `json:"id"`
	TitleUz   string `json:"title_uz"`
	TitleRu   string `json:"title_ru"`
	TitleEn   string `json:"title_en"`
	BodyUz    string `json:"body_uz"`
	BodyRu    string `json:"body_ru"`
	BodyEn    string `json:"body_en"`
	Priority  int    `json:"priority"`
	Status    string `json:"status"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	CreatedAt string `json:"created_at"`
}

type AnnouncementFilter struct {
	Status string
	Limit  int
	Offset int
}

type ReadRepository interface {
	FindByID(ctx context.Context, id string) (*AnnouncementView, error)
	List(ctx context.Context, filter AnnouncementFilter) ([]*AnnouncementView, int64, error)
	ListActive(ctx context.Context, now time.Time) ([]*AnnouncementView, error)
}
