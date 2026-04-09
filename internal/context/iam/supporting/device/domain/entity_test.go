package domain

import (
	"testing"
)

func TestHashDevice(t *testing.T) {
	hash1 := HashDevice("device-123")
	hash2 := HashDevice("device-123")

	if hash1 != hash2 {
		t.Error("same device ID should produce same hash")
	}

	hash3 := HashDevice("device-456")
	if hash1 == hash3 {
		t.Error("different device IDs should produce different hashes")
	}

	// SHA-256 produces 64 hex characters
	if len(hash1) != 64 {
		t.Errorf("hash length expected 64, got: %d", len(hash1))
	}
}

func TestHashDevice_EmptyString(t *testing.T) {
	hash := HashDevice("")
	if hash == "" {
		t.Error("empty device ID should still produce a hash")
	}
	if len(hash) != 64 {
		t.Errorf("hash length expected 64, got: %d", len(hash))
	}
}

func TestHashDevice_Deterministic(t *testing.T) {
	devices := []string{"iPhone-14-Pro", "Samsung-Galaxy-S23", "Pixel-7"}
	for _, d := range devices {
		h1 := HashDevice(d)
		h2 := HashDevice(d)
		if h1 != h2 {
			t.Errorf("HashDevice(%q) not deterministic: %s != %s", d, h1, h2)
		}
	}
}

func TestFingerprint_Fields(t *testing.T) {
	fp := &Fingerprint{
		ID:         "fp-1",
		UserID:     "user-1",
		DeviceHash: HashDevice("device-123"),
		DeviceName: "iPhone 14 Pro",
		IPAddress:  "192.168.1.100",
		Trusted:    true,
	}

	if fp.UserID != "user-1" {
		t.Errorf("UserID expected user-1, got: %s", fp.UserID)
	}
	if fp.Trusted != true {
		t.Error("Trusted should be true")
	}
	if fp.DeviceName != "iPhone 14 Pro" {
		t.Errorf("DeviceName expected iPhone 14 Pro, got: %s", fp.DeviceName)
	}
}
