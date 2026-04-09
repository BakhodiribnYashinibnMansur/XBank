package device_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	devicedomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/domain"
	devicepg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/infrastructure/postgres"
	userdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/domain"
	userpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/infrastructure/postgres"
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

func createTestUser(t *testing.T) string {
	t.Helper()
	userRepo := userpg.NewWriteRepo(pgc.Pool)
	user, err := userdomain.NewUser("device-test@example.com", "$2a$12$hashedpassword", "Test", "User")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("inserting test user: %v", err)
	}
	return user.ID
}

func TestDeviceRepo_Upsert(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := devicepg.NewWriteRepo(pgc.Pool)

	deviceHash := devicedomain.HashDevice("my-device-id-123")
	fp := &devicedomain.Fingerprint{
		UserID:     userID,
		DeviceHash: deviceHash,
		DeviceName: "Chrome on MacOS",
		IPAddress:  "192.168.1.100",
		Trusted:    false,
	}

	if err := repo.Upsert(ctx, fp); err != nil {
		t.Fatalf("repo.Upsert: %v", err)
	}
	if fp.ID == "" {
		t.Fatal("expected fingerprint ID to be set after Upsert")
	}
}

func TestDeviceRepo_Upsert_UpdateOnConflict(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := devicepg.NewWriteRepo(pgc.Pool)

	deviceHash := devicedomain.HashDevice("same-device")
	fp := &devicedomain.Fingerprint{
		UserID:     userID,
		DeviceHash: deviceHash,
		DeviceName: "Firefox",
		IPAddress:  "10.0.0.1",
		Trusted:    false,
	}
	repo.Upsert(ctx, fp)
	firstID := fp.ID

	// Upsert again with different IP — should update, not insert
	fp2 := &devicedomain.Fingerprint{
		UserID:     userID,
		DeviceHash: deviceHash,
		DeviceName: "Firefox",
		IPAddress:  "10.0.0.2",
		Trusted:    false,
	}
	if err := repo.Upsert(ctx, fp2); err != nil {
		t.Fatalf("repo.Upsert (conflict): %v", err)
	}
	if fp2.ID != firstID {
		t.Errorf("ID changed after upsert: got %q, want %q", fp2.ID, firstID)
	}
}

func TestDeviceRepo_GetByUserAndDevice(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := devicepg.NewWriteRepo(pgc.Pool)

	deviceHash := devicedomain.HashDevice("lookup-device")
	fp := &devicedomain.Fingerprint{
		UserID:     userID,
		DeviceHash: deviceHash,
		DeviceName: "Safari on iOS",
		IPAddress:  "172.16.0.1",
		Trusted:    true,
	}
	repo.Upsert(ctx, fp)

	got, err := repo.GetByUserAndDevice(ctx, userID, deviceHash)
	if err != nil {
		t.Fatalf("repo.GetByUserAndDevice: %v", err)
	}
	if got == nil {
		t.Fatal("expected fingerprint, got nil")
	}
	if got.DeviceName != "Safari on iOS" {
		t.Errorf("DeviceName = %q, want %q", got.DeviceName, "Safari on iOS")
	}
	if !got.Trusted {
		t.Error("expected Trusted to be true")
	}
}

func TestDeviceRepo_GetByUserAndDevice_NotFound(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := devicepg.NewWriteRepo(pgc.Pool)

	got, err := repo.GetByUserAndDevice(ctx, "00000000-0000-0000-0000-000000000000", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil for nonexistent device")
	}
}

func TestDeviceRepo_ListByUserID(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := devicepg.NewWriteRepo(pgc.Pool)

	fp1 := &devicedomain.Fingerprint{
		UserID:     userID,
		DeviceHash: devicedomain.HashDevice("device-1"),
		DeviceName: "Chrome",
		IPAddress:  "10.0.0.1",
		Trusted:    false,
	}
	repo.Upsert(ctx, fp1)

	fp2 := &devicedomain.Fingerprint{
		UserID:     userID,
		DeviceHash: devicedomain.HashDevice("device-2"),
		DeviceName: "Firefox",
		IPAddress:  "10.0.0.2",
		Trusted:    true,
	}
	repo.Upsert(ctx, fp2)

	fps, err := repo.ListByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("repo.ListByUserID: %v", err)
	}
	if len(fps) != 2 {
		t.Errorf("len(fps) = %d, want 2", len(fps))
	}
}

func TestDeviceRepo_Delete(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	userID := createTestUser(t)
	repo := devicepg.NewWriteRepo(pgc.Pool)

	deviceHash := devicedomain.HashDevice("delete-device")
	fp := &devicedomain.Fingerprint{
		UserID:     userID,
		DeviceHash: deviceHash,
		DeviceName: "Temp Device",
		IPAddress:  "10.0.0.1",
		Trusted:    false,
	}
	repo.Upsert(ctx, fp)

	if err := repo.Delete(ctx, fp.ID); err != nil {
		t.Fatalf("repo.Delete: %v", err)
	}

	got, _ := repo.GetByUserAndDevice(ctx, userID, deviceHash)
	if got != nil {
		t.Error("expected nil after deletion")
	}
}
