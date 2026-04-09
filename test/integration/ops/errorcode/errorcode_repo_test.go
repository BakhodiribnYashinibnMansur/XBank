package errorcode_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/infrastructure/postgres"
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

func TestErrorCodeRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	ec, err := domain.NewErrorCode(
		"INSUFFICIENT_FUNDS",
		"Insufficient funds in account",
		"Hisobda mablag' yetarli emas",
		"Nedostatochno sredstv na schete",
		"BUSINESS", "MEDIUM", 400, false, "Please top up your account",
	)
	if err != nil {
		t.Fatalf("creating error code: %v", err)
	}

	if err := repo.Save(ctx, ec); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if ec.ID == "" {
		t.Fatal("expected error code ID to be set after Save")
	}

	got, err := repo.FindByID(ctx, ec.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}

	if got.Code != "INSUFFICIENT_FUNDS" {
		t.Errorf("code = %q, want %q", got.Code, "INSUFFICIENT_FUNDS")
	}
	if got.MessageEn != "Insufficient funds in account" {
		t.Errorf("messageEn = %q, want %q", got.MessageEn, "Insufficient funds in account")
	}
	if got.MessageUz != "Hisobda mablag' yetarli emas" {
		t.Errorf("messageUz = %q, want %q", got.MessageUz, "Hisobda mablag' yetarli emas")
	}
	if got.MessageRu != "Nedostatochno sredstv na schete" {
		t.Errorf("messageRu = %q, want %q", got.MessageRu, "Nedostatochno sredstv na schete")
	}
	if got.Category != "BUSINESS" {
		t.Errorf("category = %q, want %q", got.Category, "BUSINESS")
	}
	if got.Severity != "MEDIUM" {
		t.Errorf("severity = %q, want %q", got.Severity, "MEDIUM")
	}
	if got.HTTPStatus != 400 {
		t.Errorf("httpStatus = %d, want 400", got.HTTPStatus)
	}
	if got.Retryable {
		t.Error("expected Retryable to be false")
	}
	if got.Suggestion != "Please top up your account" {
		t.Errorf("suggestion = %q, want %q", got.Suggestion, "Please top up your account")
	}
}

func TestErrorCodeRepo_FindByCode(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	ec, _ := domain.NewErrorCode("AUTH_FAILED", "Authentication failed", "", "", "SECURITY", "HIGH", 401, false, "Check credentials")
	repo.Save(ctx, ec)

	got, err := repo.FindByCode(ctx, "AUTH_FAILED")
	if err != nil {
		t.Fatalf("repo.FindByCode: %v", err)
	}
	if got.ID != ec.ID {
		t.Errorf("ID = %q, want %q", got.ID, ec.ID)
	}
	if got.MessageEn != "Authentication failed" {
		t.Errorf("messageEn = %q, want %q", got.MessageEn, "Authentication failed")
	}
}

func TestErrorCodeRepo_Update(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	ec, _ := domain.NewErrorCode("RATE_LIMITED", "Too many requests", "", "", "SECURITY", "LOW", 429, true, "Wait and retry")
	repo.Save(ctx, ec)

	newMsg := "Rate limit exceeded"
	newStatus := 503
	newRetryable := false
	ec.Update(&newMsg, nil, nil, nil, &newStatus, &newRetryable)

	if err := repo.Update(ctx, ec); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.FindByID(ctx, ec.ID)
	if got.MessageEn != "Rate limit exceeded" {
		t.Errorf("messageEn = %q, want %q", got.MessageEn, "Rate limit exceeded")
	}
	if got.HTTPStatus != 503 {
		t.Errorf("httpStatus = %d, want 503", got.HTTPStatus)
	}
	if got.Retryable {
		t.Error("expected Retryable to be false after update")
	}
}

func TestErrorCodeRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	ec, _ := domain.NewErrorCode("TEMP_CODE", "Temporary", "", "", "SYSTEM", "LOW", 500, false, "")
	repo.Save(ctx, ec)

	if err := repo.Delete(ctx, ec.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, ec.ID)
	if err == nil {
		t.Fatal("expected error after deleting error code")
	}
}
