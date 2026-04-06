package application

import "time"

type CreateAnnouncementRequest struct {
	TitleUz   string     `json:"title_uz"`
	TitleRu   string     `json:"title_ru"`
	TitleEn   string     `json:"title_en"`
	BodyUz    string     `json:"body_uz"`
	BodyRu    string     `json:"body_ru"`
	BodyEn    string     `json:"body_en"`
	Priority  int        `json:"priority"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}
