package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultsFromYAML(t *testing.T) {
	path := writeTestConfig(t)
	cfg := Load(path)

	if cfg.App.Name != "TestBank" {
		t.Errorf("app.name = %q, want TestBank", cfg.App.Name)
	}
	if cfg.App.Port != 4000 {
		t.Errorf("app.port = %d, want 4000", cfg.App.Port)
	}
	if cfg.JWT.Issuer != "test-api" {
		t.Errorf("jwt.issuer = %q, want test-api", cfg.JWT.Issuer)
	}
	if cfg.JWT.AccessTTL().Minutes() != 30 {
		t.Errorf("jwt.access_ttl = %v, want 30m", cfg.JWT.AccessTTL())
	}
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	path := writeTestConfig(t)

	t.Setenv("DATABASE_URL", "postgres://env:env@envhost:5432/envdb")
	t.Setenv("REDIS_URL", "redis://env-redis:6379")
	t.Setenv("CARD_ENCRYPTION_KEY", "env-card-key")
	t.Setenv("HMAC_SECRET", "env-hmac-secret")

	cfg := Load(path)

	if cfg.Database.URL != "postgres://env:env@envhost:5432/envdb" {
		t.Errorf("database.url = %q, want env value", cfg.Database.URL)
	}
	if cfg.Redis.URL != "redis://env-redis:6379" {
		t.Errorf("redis.url = %q, want env value", cfg.Redis.URL)
	}
	if cfg.Encryption.CardKey != "env-card-key" {
		t.Errorf("encryption.card_key = %q, want env value", cfg.Encryption.CardKey)
	}
	if cfg.Encryption.HMACSecret != "env-hmac-secret" {
		t.Errorf("encryption.hmac_secret = %q, want env value", cfg.Encryption.HMACSecret)
	}
}

func TestLoad_XBANKPrefixEnv(t *testing.T) {
	path := writeTestConfig(t)

	t.Setenv("XBANK_APP_PORT", "9999")

	cfg := Load(path)

	if cfg.App.Port != 9999 {
		t.Errorf("app.port = %d, want 9999 (from XBANK_APP_PORT)", cfg.App.Port)
	}
}

func TestLoad_HelperMethods(t *testing.T) {
	path := writeTestConfig(t)
	cfg := Load(path)

	if cfg.JWT.AccessTTL().Minutes() != 30 {
		t.Errorf("AccessTTL = %v, want 30m", cfg.JWT.AccessTTL())
	}
	if cfg.JWT.RefreshTTL().Hours() != 7*24 {
		t.Errorf("RefreshTTL = %v, want 168h", cfg.JWT.RefreshTTL())
	}
	if cfg.RateLimit.Window().Minutes() != 2 {
		t.Errorf("RateLimit.Window = %v, want 2m", cfg.RateLimit.Window())
	}
	if cfg.HMAC.MaxClockSkew().Minutes() != 10 {
		t.Errorf("HMAC.MaxClockSkew = %v, want 10m", cfg.HMAC.MaxClockSkew())
	}
	if cfg.CORS.Origins() != "http://test:3000" {
		t.Errorf("CORS.Origins = %q", cfg.CORS.Origins())
	}
}

func writeTestConfig(t *testing.T) string {
	t.Helper()
	content := `
app:
  name: TestBank
  port: 4000
jwt:
  issuer: test-api
  audience: test-client
  access_ttl_minutes: 30
  refresh_ttl_days: 7
  private_key_path: keys/private.pem
  public_key_path: keys/public.pem
rate_limit:
  max_requests: 100
  window_minutes: 2
hmac:
  max_clock_skew_minutes: 10
cors:
  allowed_origins:
    - http://test:3000
mongodb:
  database: test_audit
kafka:
  brokers:
    - localhost:9092
  topics:
    account_opened: test.account.opened
    account_credited: test.account.credited
    account_debited: test.account.debited
    account_frozen: test.account.frozen
    account_closed: test.account.closed
    transfer_created: test.transfer.created
    transfer_completed: test.transfer.completed
    transfer_failed: test.transfer.failed
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test config: %v", err)
	}
	return path
}
