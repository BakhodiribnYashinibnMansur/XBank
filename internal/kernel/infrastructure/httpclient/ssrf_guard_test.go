package httpclient

import (
	"strings"
	"testing"
)

func TestSSRFGuard_BlocksPrivateIPs(t *testing.T) {
	guard := NewSSRFGuard(nil, nil)

	blocked := []string{
		"http://127.0.0.1/admin",
		"http://127.0.0.1:8200/v1/secret",
		"http://10.0.0.1/internal",
		"http://172.16.0.1/api",
		"http://192.168.1.1/router",
		"http://169.254.169.254/latest/meta-data/", // AWS metadata
		"http://0.0.0.0/",
	}

	for _, u := range blocked {
		err := guard.Validate(u)
		if err == nil {
			t.Errorf("expected SSRF block for %s, got nil", u)
		}
	}
}

func TestSSRFGuard_BlocksProtocols(t *testing.T) {
	guard := NewSSRFGuard(nil, nil)

	blocked := []string{
		"file:///etc/passwd",
		"ftp://internal.server/data",
		"gopher://localhost:6379/_FLUSHALL",
		"dict://localhost:6379/info",
	}

	for _, u := range blocked {
		err := guard.Validate(u)
		if err == nil {
			t.Errorf("expected protocol block for %s, got nil", u)
		}
		if err != nil && !strings.Contains(err.Error(), "blocked protocol") {
			t.Errorf("expected 'blocked protocol' error for %s, got: %v", u, err)
		}
	}
}

func TestSSRFGuard_BlocksUnallowedPorts(t *testing.T) {
	guard := NewSSRFGuard(nil, nil) // default: 80, 443 only

	blocked := []string{
		"http://example.com:6379/",  // Redis
		"http://example.com:5432/",  // PostgreSQL
		"http://example.com:27017/", // MongoDB
		"http://example.com:8200/",  // Vault
		"http://example.com:9092/",  // Kafka
		"http://example.com:3000/",  // Internal app
	}

	for _, u := range blocked {
		err := guard.Validate(u)
		if err == nil {
			t.Errorf("expected port block for %s, got nil", u)
		}
		if err != nil && !strings.Contains(err.Error(), "blocked port") {
			t.Errorf("expected 'blocked port' error for %s, got: %v", u, err)
		}
	}
}

func TestSSRFGuard_AllowsPublicURLs(t *testing.T) {
	guard := NewSSRFGuard(nil, nil)

	allowed := []string{
		"https://api.telegram.org/bot123/sendMessage",
		"https://fcm.googleapis.com/v1/projects/myproject/messages:send",
		"https://example.com/webhook",
		"http://example.com/api",
	}

	for _, u := range allowed {
		err := guard.Validate(u)
		if err != nil {
			t.Errorf("expected allow for %s, got: %v", u, err)
		}
	}
}

func TestSSRFGuard_HostAllowlist(t *testing.T) {
	guard := NewSSRFGuard(
		[]string{"api.telegram.org", "fcm.googleapis.com"},
		nil,
	)

	// Allowed hosts
	err := guard.Validate("https://api.telegram.org/bot123/sendMessage")
	if err != nil {
		t.Errorf("telegram should be allowed: %v", err)
	}

	err = guard.Validate("https://fcm.googleapis.com/v1/messages")
	if err != nil {
		t.Errorf("FCM should be allowed: %v", err)
	}

	// Blocked host (not in allowlist)
	err = guard.Validate("https://evil.com/steal")
	if err == nil {
		t.Error("expected block for host not in allowlist")
	}
	if err != nil && !strings.Contains(err.Error(), "not in allowlist") {
		t.Errorf("expected 'not in allowlist' error, got: %v", err)
	}
}

func TestSSRFGuard_CustomPorts(t *testing.T) {
	guard := NewSSRFGuard(nil, []string{"80", "443", "8443"})

	// Allowed custom port
	err := guard.Validate("https://example.com:8443/api")
	if err != nil {
		t.Errorf("port 8443 should be allowed: %v", err)
	}

	// Blocked port
	err = guard.Validate("http://example.com:9999/api")
	if err == nil {
		t.Error("port 9999 should be blocked")
	}
}

func TestSSRFGuard_EmptyAndInvalidURLs(t *testing.T) {
	guard := NewSSRFGuard(nil, nil)

	invalids := []string{
		"",
		"not-a-url",
		"://missing-scheme",
		"http://",
	}

	for _, u := range invalids {
		err := guard.Validate(u)
		if err == nil {
			t.Errorf("expected error for invalid URL %q, got nil", u)
		}
	}
}

func TestSSRFGuard_BlocksLocalhost(t *testing.T) {
	guard := NewSSRFGuard(nil, nil)

	localhosts := []string{
		"http://localhost/admin",
		"http://localhost:8200/v1/secret",
	}

	for _, u := range localhosts {
		err := guard.Validate(u)
		if err == nil {
			t.Errorf("expected block for localhost URL %s", u)
		}
	}
}
