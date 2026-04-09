package usersetting_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	settingdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/domain"
	settingpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/infrastructure/postgres"
	userdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	userpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/infrastructure/postgres"
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

func createTestUser(t *testing.T) string {
	t.Helper()
	userRepo := userpg.NewWriteRepo(pgc.Pool)
	user, err := userdomain.NewUser("setting-test@example.com", "$2a$12$hashedpassword", "Test", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestUserSettingRepo_Upsert(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := settingpg.NewWriteRepo(pgc.Pool)

	setting, err := settingdomain.NewUserSetting(userID, "language", "en")
	if err != nil {
		t.Fatalf("NewUserSetting: %v", err)
	}
	setting.ID = uuid.New().String()

	if err := repo.Upsert(ctx, setting); err != nil {
		t.Fatalf("repo.Upsert: %v", err)
	}

	// Upsert again with different value (should update)
	setting.Value = "ru"
	setting.UpdateValue("ru")
	if err := repo.Upsert(ctx, setting); err != nil {
		t.Fatalf("repo.Upsert (update): %v", err)
	}
}

func TestUserSettingRepo_FindByUserIDAndKey(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := settingpg.NewWriteRepo(pgc.Pool)

	setting, _ := settingdomain.NewUserSetting(userID, "theme", "dark")
	setting.ID = uuid.New().String()
	repo.Upsert(ctx, setting)

	result, err := repo.FindByUserIDAndKey(ctx, userID, "theme")
	if err != nil {
		t.Fatalf("repo.FindByUserIDAndKey: %v", err)
	}
	got := result.(*settingdomain.UserSetting)
	if got.Value != "dark" {
		t.Errorf("Value = %q, want %q", got.Value, "dark")
	}
	if got.UserID != userID {
		t.Errorf("UserID = %q, want %q", got.UserID, userID)
	}
}

func TestUserSettingRepo_FindByUserIDAndKey_NotFound(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := settingpg.NewWriteRepo(pgc.Pool)

	_, err := repo.FindByUserIDAndKey(ctx, "00000000-0000-0000-0000-000000000000", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent setting")
	}
}

func TestUserSettingRepo_Delete(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := settingpg.NewWriteRepo(pgc.Pool)

	setting, _ := settingdomain.NewUserSetting(userID, "notifications", "true")
	setting.ID = uuid.New().String()
	repo.Upsert(ctx, setting)

	if err := repo.Delete(ctx, setting.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	_, err := repo.FindByUserIDAndKey(ctx, userID, "notifications")
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestUserSettingReadRepo_List(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	writeRepo := settingpg.NewWriteRepo(pgc.Pool)
	readRepo := settingpg.NewReadRepo(pgc.Pool)

	// Create multiple settings
	s1, _ := settingdomain.NewUserSetting(userID, "language", "en")
	s1.ID = uuid.New().String()
	writeRepo.Upsert(ctx, s1)

	s2, _ := settingdomain.NewUserSetting(userID, "theme", "dark")
	s2.ID = uuid.New().String()
	writeRepo.Upsert(ctx, s2)

	items, total, err := readRepo.List(ctx, settingdomain.UserSettingFilter{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("readRepo.List: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(items))
	}
}
