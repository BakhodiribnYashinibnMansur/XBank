package domain

import (
	"testing"
	"time"
)

func TestNewAnnouncement_Success(t *testing.T) {
	a, err := NewAnnouncement("Title UZ", "Title RU", "Title EN", "Body UZ", "Body RU", "Body EN", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.TitleUz != "Title UZ" {
		t.Errorf("TitleUz mismatch, got: %s", a.TitleUz)
	}
	if a.Status != StatusDraft {
		t.Errorf("Status expected DRAFT, got: %s", a.Status)
	}
	if a.Priority != 1 {
		t.Errorf("Priority expected 1, got: %d", a.Priority)
	}
}

func TestNewAnnouncement_AllTitlesEmpty(t *testing.T) {
	_, err := NewAnnouncement("", "", "", "Body", "Body", "Body", 0)
	if err != ErrEmptyTitle {
		t.Errorf("expected ErrEmptyTitle, got: %v", err)
	}
}

func TestNewAnnouncement_SingleLanguageTitle(t *testing.T) {
	// Should succeed with at least one title
	a, err := NewAnnouncement("", "", "English Only", "", "", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.TitleEn != "English Only" {
		t.Errorf("TitleEn mismatch, got: %s", a.TitleEn)
	}
}

func TestAnnouncement_Publish(t *testing.T) {
	a, _ := NewAnnouncement("Title", "", "", "", "", "", 0)

	if err := a.Publish(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Status != StatusPublished {
		t.Errorf("Status expected PUBLISHED, got: %s", a.Status)
	}
}

func TestAnnouncement_Publish_AlreadyPublished(t *testing.T) {
	a, _ := NewAnnouncement("Title", "", "", "", "", "", 0)
	a.Publish()

	err := a.Publish()
	if err != ErrAlreadyPublished {
		t.Errorf("expected ErrAlreadyPublished, got: %v", err)
	}
}

func TestAnnouncement_Update_Draft(t *testing.T) {
	a, _ := NewAnnouncement("Old", "", "", "", "", "", 0)

	newTitle := "New Title"
	newPriority := 5
	err := a.Update(&newTitle, nil, nil, nil, nil, nil, &newPriority, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.TitleUz != "New Title" {
		t.Errorf("TitleUz expected New Title, got: %s", a.TitleUz)
	}
	if a.Priority != 5 {
		t.Errorf("Priority expected 5, got: %d", a.Priority)
	}
}

func TestAnnouncement_Update_Published(t *testing.T) {
	a, _ := NewAnnouncement("Title", "", "", "", "", "", 0)
	a.Publish()

	newTitle := "Updated"
	err := a.Update(&newTitle, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != ErrCannotEditPublished {
		t.Errorf("expected ErrCannotEditPublished, got: %v", err)
	}
}

func TestAnnouncement_IsActive(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name      string
		status    AnnouncementStatus
		startDate *time.Time
		endDate   *time.Time
		want      bool
	}{
		{"draft is not active", StatusDraft, nil, nil, false},
		{"published no dates", StatusPublished, nil, nil, true},
		{"published within window", StatusPublished, &past, &future, true},
		{"published before start", StatusPublished, &future, nil, false},
		{"published after end", StatusPublished, nil, &past, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := NewAnnouncement("Title", "", "", "", "", "", 0)
			a.Status = tt.status
			a.StartDate = tt.startDate
			a.EndDate = tt.endDate

			if got := a.IsActive(now); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}
