package config

import (
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App       AppConfig       `yaml:"app"`
	JWT       JWTConfig       `yaml:"jwt"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	CORS      CORSConfig      `yaml:"cors"`
	Kafka     KafkaConfig     `yaml:"kafka"`
	MongoDB   MongoDBConfig   `yaml:"mongodb"`
	Database  DatabaseConfig
}

type AppConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

type JWTConfig struct {
	PrivateKeyPath   string `yaml:"private_key_path"`
	PublicKeyPath    string `yaml:"public_key_path"`
	Issuer           string `yaml:"issuer"`
	Audience         string `yaml:"audience"`
	AccessTTLMinutes int    `yaml:"access_ttl_minutes"`
	RefreshTTLDays   int    `yaml:"refresh_ttl_days"`
}

type RateLimitConfig struct {
	MaxRequests   int `yaml:"max_requests"`
	WindowMinutes int `yaml:"window_minutes"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type KafkaConfig struct {
	Brokers []string          `yaml:"brokers"`
	Topics  KafkaTopicsConfig `yaml:"topics"`
}

type KafkaTopicsConfig struct {
	AccountOpened    string `yaml:"account_opened"`
	AccountCredited  string `yaml:"account_credited"`
	AccountDebited   string `yaml:"account_debited"`
	AccountFrozen    string `yaml:"account_frozen"`
	AccountClosed    string `yaml:"account_closed"`
	TransferCreated  string `yaml:"transfer_created"`
	TransferCompleted string `yaml:"transfer_completed"`
	TransferFailed   string `yaml:"transfer_failed"`
}

type MongoDBConfig struct {
	URI      string // env dan o'qiladi
	Database string `yaml:"database"`
}

type DatabaseConfig struct {
	URL string
}

func (j *JWTConfig) AccessTTL() time.Duration {
	return time.Duration(j.AccessTTLMinutes) * time.Minute
}

func (j *JWTConfig) RefreshTTL() time.Duration {
	return time.Duration(j.RefreshTTLDays) * 24 * time.Hour
}

func (r *RateLimitConfig) Window() time.Duration {
	return time.Duration(r.WindowMinutes) * time.Minute
}

func (c *CORSConfig) Origins() string {
	return strings.Join(c.AllowedOrigins, ",")
}

func Load(path string) *Config {
	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("config.yml o'qib bo'lmadi: %v", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Fatalf("config.yml parse xatolik: %v", err)
	}

	// Secrets from ENV
	cfg.Database.URL = getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/xbank?sslmode=disable")
	cfg.MongoDB.URI = getEnv("MONGODB_URI", "mongodb://localhost:27017")

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
