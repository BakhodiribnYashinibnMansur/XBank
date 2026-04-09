package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	HMAC       HMACConfig       `mapstructure:"hmac"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	CORS       CORSConfig       `mapstructure:"cors"`
	Kafka      KafkaConfig      `mapstructure:"kafka"`
	MongoDB    MongoDBConfig    `mapstructure:"mongodb"`
	Jaeger     JaegerConfig     `mapstructure:"jaeger"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Vault      VaultConfig      `mapstructure:"vault"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

type JWTConfig struct {
	PrivateKeyPath   string `mapstructure:"private_key_path"`
	PublicKeyPath    string `mapstructure:"public_key_path"`
	Issuer           string `mapstructure:"issuer"`
	Audience         string `mapstructure:"audience"`
	AccessTTLMinutes int    `mapstructure:"access_ttl_minutes"`
	RefreshTTLDays   int    `mapstructure:"refresh_ttl_days"`
}

func (j *JWTConfig) AccessTTL() time.Duration {
	return time.Duration(j.AccessTTLMinutes) * time.Minute
}

func (j *JWTConfig) RefreshTTL() time.Duration {
	return time.Duration(j.RefreshTTLDays) * 24 * time.Hour
}

type HMACConfig struct {
	MaxClockSkewMinutes int `mapstructure:"max_clock_skew_minutes"`
}

func (h *HMACConfig) MaxClockSkew() time.Duration {
	return time.Duration(h.MaxClockSkewMinutes) * time.Minute
}

type RateLimitConfig struct {
	MaxRequests   int `mapstructure:"max_requests"`
	WindowMinutes int `mapstructure:"window_minutes"`
}

func (r *RateLimitConfig) Window() time.Duration {
	return time.Duration(r.WindowMinutes) * time.Minute
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

func (c *CORSConfig) Origins() string {
	return strings.Join(c.AllowedOrigins, ",")
}

type KafkaConfig struct {
	Brokers           []string          `mapstructure:"brokers"`
	Topics            KafkaTopicsConfig `mapstructure:"topics"`
	SchemaRegistryURL string            `mapstructure:"schema_registry_url"`
}

type KafkaTopicsConfig struct {
	AccountOpened         string `mapstructure:"account_opened"`
	AccountCredited       string `mapstructure:"account_credited"`
	AccountDebited        string `mapstructure:"account_debited"`
	AccountFrozen         string `mapstructure:"account_frozen"`
	AccountClosed         string `mapstructure:"account_closed"`
	TransferCreated       string `mapstructure:"transfer_created"`
	TransferCompleted     string `mapstructure:"transfer_completed"`
	TransferFailed        string `mapstructure:"transfer_failed"`
	CardIssued            string `mapstructure:"card_issued"`
	CardBlocked           string `mapstructure:"card_blocked"`
	CardActivated         string `mapstructure:"card_activated"`
	KYCSubmitted          string `mapstructure:"kyc_submitted"`
	KYCApproved           string `mapstructure:"kyc_approved"`
	KYCRejected           string `mapstructure:"kyc_rejected"`
	NotificationRequested string `mapstructure:"notification_requested"`
}

type MongoDBConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type EncryptionConfig struct {
	CardKey    string `mapstructure:"card_key"`
	HMACSecret string `mapstructure:"hmac_secret"`
}

type VaultConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Address   string `mapstructure:"address"`
	Token     string `mapstructure:"token"`
	MountPath string `mapstructure:"mount_path"`
}

type JaegerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

// Load reads configuration from YAML file and environment variables.
// ENV variables override YAML values. Prefix: XBANK_
// Example: XBANK_DATABASE_URL overrides database.url
func Load(path string) *Config {
	v := viper.New()

	// Defaults
	v.SetDefault("app.name", "XBank")
	v.SetDefault("app.port", 3000)
	v.SetDefault("jwt.issuer", "xbank-api")
	v.SetDefault("jwt.audience", "xbank-client")
	v.SetDefault("jwt.access_ttl_minutes", 15)
	v.SetDefault("jwt.refresh_ttl_days", 30)
	v.SetDefault("jwt.private_key_path", "keys/private.pem")
	v.SetDefault("jwt.public_key_path", "keys/public.pem")
	v.SetDefault("rate_limit.max_requests", 60)
	v.SetDefault("rate_limit.window_minutes", 1)
	v.SetDefault("hmac.max_clock_skew_minutes", 5)
	v.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	v.SetDefault("database.url", "postgres://postgres:postgres@localhost:5432/xbank?sslmode=disable")
	v.SetDefault("redis.url", "redis://localhost:6379/0")
	v.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	v.SetDefault("mongodb.database", "xbank_audit")
	v.SetDefault("jaeger.enabled", true)
	v.SetDefault("jaeger.endpoint", "localhost:4318")
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("vault.address", "http://localhost:8200")
	v.SetDefault("vault.mount_path", "secret")

	// YAML config file
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("config file o'qib bo'lmadi (%s): %v", path, err)
	}

	// Environment variables (XBANK_ prefix)
	v.SetEnvPrefix("XBANK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Legacy ENV variable bindings (backward compatibility)
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("redis.url", "REDIS_URL")
	v.BindEnv("mongodb.uri", "MONGODB_URI")
	v.BindEnv("encryption.card_key", "CARD_ENCRYPTION_KEY")
	v.BindEnv("encryption.hmac_secret", "HMAC_SECRET")
	v.BindEnv("kafka.schema_registry_url", "SCHEMA_REGISTRY_URL")
	v.BindEnv("jaeger.endpoint", "JAEGER_ENDPOINT")
	v.BindEnv("vault.enabled", "VAULT_ENABLED")
	v.BindEnv("vault.address", "VAULT_ADDR")
	v.BindEnv("vault.token", "VAULT_TOKEN")
	v.BindEnv("vault.mount_path", "VAULT_MOUNT_PATH")

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		log.Fatalf("config unmarshal xatolik: %v", err)
	}

	// Production SSL warning
	if !strings.Contains(cfg.Database.URL, "sslmode=verify") {
		env := v.GetString("app_env")
		if env == "" {
			env = "development"
		}
		if env == "production" {
			log.Println("WARNING: DATABASE_URL does not use sslmode=verify-full. This is insecure for production!")
		}
	}

	return cfg
}
