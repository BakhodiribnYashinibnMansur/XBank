package notification_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/infrastructure/postgres"
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

func TestNotificationRepo_SaveAndFindByID(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	data := map[string]string{"account_id": "acc-001", "amount": "1500.00"}
	notif, err := domain.NewNotification("user-100", "Transfer received", "You received $1500", domain.NotificationInfo, data)
	if err != nil {
		t.Fatalf("creating notification: %v", err)
	}

	if err := repo.Save(ctx, notif); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	if notif.ID == "" {
		t.Fatal("expected notification ID to be set after Save")
	}

	got, err := repo.FindByID(ctx, notif.ID)
	if err != nil {
		t.Fatalf("repo.FindByID: %v", err)
	}

	if got.UserID != "user-100" {
		t.Errorf("userID = %q, want %q", got.UserID, "user-100")
	}
	if got.Title != "Transfer received" {
		t.Errorf("title = %q, want %q", got.Title, "Transfer received")
	}
	if got.Message != "You received $1500" {
		t.Errorf("message = %q, want %q", got.Message, "You received $1500")
	}
	if got.Type != domain.NotificationInfo {
		t.Errorf("type = %q, want %q", got.Type, domain.NotificationInfo)
	}
	if got.ReadAt != nil {
		t.Error("expected ReadAt to be nil for new notification")
	}
	if got.Data["account_id"] != "acc-001" {
		t.Errorf("data[account_id] = %q, want %q", got.Data["account_id"], "acc-001")
	}
	if got.Data["amount"] != "1500.00" {
		t.Errorf("data[amount] = %q, want %q", got.Data["amount"], "1500.00")
	}
}

func TestNotificationRepo_Update_MarkAsRead(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	notif, _ := domain.NewNotification("user-200", "Alert", "Security alert", domain.NotificationAlert, nil)
	repo.Save(ctx, notif)

	notif.MarkAsRead()
	if err := repo.Update(ctx, notif); err != nil {
		t.Fatalf("repo.Update: %v", err)
	}

	got, _ := repo.FindByID(ctx, notif.ID)
	if got.ReadAt == nil {
		t.Fatal("expected ReadAt to be set after MarkAsRead")
	}
}

func TestNotificationRepo_Delete(t *testing.T) {
	truncate(t)
	repo := pg.NewWriteRepo(pgc.Pool)
	ctx := context.Background()

	notif, _ := domain.NewNotification("user-300", "Info", "Test notification", domain.NotificationInfo, nil)
	repo.Save(ctx, notif)

	if err := repo.Delete(ctx, notif.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByID(ctx, notif.ID)
	if err == nil {
		t.Fatal("expected error after deleting notification")
	}
}
