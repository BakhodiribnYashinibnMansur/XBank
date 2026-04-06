package event

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

type AnnouncementPublished struct {
	domain.BaseEvent
}

func NewAnnouncementPublished(id string) AnnouncementPublished {
	return AnnouncementPublished{BaseEvent: domain.NewBaseEvent("announcement.published", id)}
}
