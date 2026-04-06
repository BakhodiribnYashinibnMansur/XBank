package entity

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

type AnnouncementStatus string

const (
	StatusDraft     AnnouncementStatus = "DRAFT"
	StatusPublished AnnouncementStatus = "PUBLISHED"
)

// Announcement is the aggregate root for system announcements.
// Supports multi-language (uz/ru/en), priority, and publish lifecycle.
type Announcement struct {
	domain.AggregateRoot
	TitleUz   string
	TitleRu   string
	TitleEn   string
	BodyUz    string
	BodyRu    string
	BodyEn    string
	Priority  int
	Status    AnnouncementStatus
	StartDate *time.Time
	EndDate   *time.Time
}

func NewAnnouncement(titleUz, titleRu, titleEn, bodyUz, bodyRu, bodyEn string, priority int) (*Announcement, error) {
	if titleUz == "" && titleRu == "" && titleEn == "" {
		return nil, ErrEmptyTitle
	}

	now := time.Now()
	a := &Announcement{
		TitleUz:  titleUz,
		TitleRu:  titleRu,
		TitleEn:  titleEn,
		BodyUz:   bodyUz,
		BodyRu:   bodyRu,
		BodyEn:   bodyEn,
		Priority: priority,
		Status:   StatusDraft,
	}
	a.CreatedAt = now
	a.UpdatedAt = now
	return a, nil
}

// Publish makes the announcement visible. One-way transition (irreversible).
func (a *Announcement) Publish() error {
	if a.Status == StatusPublished {
		return ErrAlreadyPublished
	}
	a.Status = StatusPublished
	a.Touch()
	return nil
}

// Update modifies draft announcement content.
func (a *Announcement) Update(titleUz, titleRu, titleEn, bodyUz, bodyRu, bodyEn *string, priority *int, startDate, endDate *time.Time) error {
	if a.Status == StatusPublished {
		return ErrCannotEditPublished
	}
	if titleUz != nil {
		a.TitleUz = *titleUz
	}
	if titleRu != nil {
		a.TitleRu = *titleRu
	}
	if titleEn != nil {
		a.TitleEn = *titleEn
	}
	if bodyUz != nil {
		a.BodyUz = *bodyUz
	}
	if bodyRu != nil {
		a.BodyRu = *bodyRu
	}
	if bodyEn != nil {
		a.BodyEn = *bodyEn
	}
	if priority != nil {
		a.Priority = *priority
	}
	if startDate != nil {
		a.StartDate = startDate
	}
	if endDate != nil {
		a.EndDate = endDate
	}
	a.Touch()
	return nil
}

// IsActive returns true if the announcement is published and within the date window.
func (a *Announcement) IsActive(now time.Time) bool {
	if a.Status != StatusPublished {
		return false
	}
	if a.StartDate != nil && now.Before(*a.StartDate) {
		return false
	}
	if a.EndDate != nil && now.After(*a.EndDate) {
		return false
	}
	return true
}
