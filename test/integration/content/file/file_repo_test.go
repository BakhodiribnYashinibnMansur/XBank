package file_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/infrastructure/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/test/integration/common/setup"
	"github.com/google/uuid"
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

func TestFileRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	file, err := domain.NewFile(
		"abc123.pdf", "report.pdf", "application/pdf",
		102400, "uploads/abc123.pdf", "https://cdn.example.com/abc123.pdf", "user-50",
	)
	if err != nil {
		t.Fatalf("creating file: %v", err)
	}
	file.ID = uuid.New().String()

	if err := repo.Save(ctx, file); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}

	result, err := repo.FindByID(ctx, file.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}
	got := result.(*domain.File)

	if got.Name != "abc123.pdf" {
		t.Errorf("name = %q, want %q", got.Name, "abc123.pdf")
	}
	if got.OriginalName != "report.pdf" {
		t.Errorf("originalName = %q, want %q", got.OriginalName, "report.pdf")
	}
	if got.MimeType != "application/pdf" {
		t.Errorf("mimeType = %q, want %q", got.MimeType, "application/pdf")
	}
	if got.Size != 102400 {
		t.Errorf("size = %d, want 102400", got.Size)
	}
	if got.Path != "uploads/abc123.pdf" {
		t.Errorf("path = %q, want %q", got.Path, "uploads/abc123.pdf")
	}
	if got.URL != "https://cdn.example.com/abc123.pdf" {
		t.Errorf("url = %q, want %q", got.URL, "https://cdn.example.com/abc123.pdf")
	}
	if got.UploadedBy != "user-50" {
		t.Errorf("uploadedBy = %q, want %q", got.UploadedBy, "user-50")
	}
}

func TestFileRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	file, _ := domain.NewFile("del.png", "photo.png", "image/png", 2048, "uploads/del.png", "https://cdn.example.com/del.png", "user-60")
	file.ID = uuid.New().String()
	repo.Save(ctx, file)

	if err := repo.Delete(ctx, file.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, file.ID)
	if err == nil {
		t.Fatal("expected error after deleting file")
	}
}
