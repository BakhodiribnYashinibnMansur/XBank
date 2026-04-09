package dataexport_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/infrastructure/postgres"
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

func TestDataExportRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	export, err := domain.NewDataExport("user-123")
	if err != nil {
		t.Fatalf("creating data export: %v", err)
	}
	export.ID = uuid.New().String()

	if err := repo.Save(ctx, export); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}

	got, err := repo.FindByID(ctx, export.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}

	if got.UserID != "user-123" {
		t.Errorf("userID = %q, want %q", got.UserID, "user-123")
	}
	if got.Status != domain.ExportPending {
		t.Errorf("status = %q, want %q", got.Status, domain.ExportPending)
	}
	if got.FileURL != "" {
		t.Errorf("fileURL = %q, want empty", got.FileURL)
	}
	if got.ErrorMsg != "" {
		t.Errorf("errorMsg = %q, want empty", got.ErrorMsg)
	}
}

func TestDataExportRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	export, _ := domain.NewDataExport("user-456")
	export.ID = uuid.New().String()
	repo.Save(ctx, export)

	// Transition to PROCESSING
	if err := export.StartProcessing(); err != nil {
		t.Fatalf("StartProcessing: %v", err)
	}
	if err := repo.Update(ctx, export); err != nil {
		t.Fatalf("repo.Update (processing): %v", err)
	}

	got, _ := repo.FindByID(ctx, export.ID)
	if got.Status != domain.ExportProcessing {
		t.Errorf("status = %q, want %q", got.Status, domain.ExportProcessing)
	}

	// Transition to COMPLETED
	if err := export.Complete("https://storage.example.com/export.zip"); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if err := repo.Update(ctx, export); err != nil {
		t.Fatalf("repo.Update (completed): %v", err)
	}

	got, _ = repo.FindByID(ctx, export.ID)
	if got.Status != domain.ExportStatusCompleted {
		t.Errorf("status = %q, want %q", got.Status, domain.ExportStatusCompleted)
	}
	if got.FileURL != "https://storage.example.com/export.zip" {
		t.Errorf("fileURL = %q, want %q", got.FileURL, "https://storage.example.com/export.zip")
	}
}
