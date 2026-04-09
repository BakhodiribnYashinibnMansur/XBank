package announcement_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/infrastructure/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/test/integration/common/setup"
)

var pgc *setup.PostgresContainer

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION") == "" {
		fmt.Println("skipping integration tests (set INTEGRATION=1 to run)")
		os.Exit(0)
	}

	var code int
	func() {
		t := &testing.T{}
		pgc = setup.MustPostgres(t)
		defer pgc.Teardown()
		code = m.Run()
	}()
	os.Exit(code)
}

func truncate(t *testing.T) {
	t.Helper()
	setup.TruncateAll(t, pgc.Pool)
}

func TestAnnouncementRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	ann, err := domain.NewAnnouncement(
		"Yangilik", "Novost'", "News",
		"Tafsilotlar", "Podrobnosti", "Details",
		1,
	)
	if err != nil {
		t.Fatalf("creating announcement: %v", err)
	}

	if err := repo.Save(ctx, ann); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if ann.ID == "" {
		t.Fatal("expected announcement ID to be set after Save")
	}

	got, err := repo.FindByID(ctx, ann.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}

	if got.TitleUz != "Yangilik" {
		t.Errorf("titleUz = %q, want %q", got.TitleUz, "Yangilik")
	}
	if got.TitleRu != "Novost'" {
		t.Errorf("titleRu = %q, want %q", got.TitleRu, "Novost'")
	}
	if got.TitleEn != "News" {
		t.Errorf("titleEn = %q, want %q", got.TitleEn, "News")
	}
	if got.BodyUz != "Tafsilotlar" {
		t.Errorf("bodyUz = %q, want %q", got.BodyUz, "Tafsilotlar")
	}
	if got.Priority != 1 {
		t.Errorf("priority = %d, want 1", got.Priority)
	}
	if got.Status != domain.StatusDraft {
		t.Errorf("status = %q, want %q", got.Status, domain.StatusDraft)
	}
	if got.StartDate != nil {
		t.Error("expected StartDate to be nil")
	}
	if got.EndDate != nil {
		t.Error("expected EndDate to be nil")
	}
}

func TestAnnouncementRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	ann, _ := domain.NewAnnouncement("UzTitle", "RuTitle", "EnTitle", "UzBody", "RuBody", "EnBody", 0)
	repo.Save(ctx, ann)

	newTitle := "Updated Title"
	newPriority := 5
	startDate := time.Now().Add(24 * time.Hour).Truncate(time.Microsecond)
	if err := ann.Update(nil, nil, &newTitle, nil, nil, nil, &newPriority, &startDate, nil); err != nil {
		t.Fatalf("ann.Update: %v", err)
	}

	if err := repo.Update(ctx, ann); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.FindByID(ctx, ann.ID)
	if got.TitleEn != "Updated Title" {
		t.Errorf("titleEn = %q, want %q", got.TitleEn, "Updated Title")
	}
	if got.Priority != 5 {
		t.Errorf("priority = %d, want 5", got.Priority)
	}
	if got.StartDate == nil {
		t.Fatal("expected StartDate to be set")
	}
}

func TestAnnouncementRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	ann, _ := domain.NewAnnouncement("Del", "", "", "", "", "", 0)
	repo.Save(ctx, ann)

	if err := repo.Delete(ctx, ann.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, ann.ID)
	if err == nil {
		t.Fatal("expected error after deleting announcement")
	}
}
