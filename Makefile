.PHONY: run build clean test docker-up docker-down migrate-up migrate-down migrate-status migrate-create seed

# Environment
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/xbank?sslmode=disable
MIGRATIONS_DIR = ./migrations

# Lokal ishga tushirish
run:
	go run ./cmd/api

# Binary build qilish
build:
	CGO_ENABLED=0 go build -o bin/xbank ./cmd/api

# Build artefaktlarni tozalash
clean:
	rm -rf bin/

# Testlarni ishga tushirish
test:
	go test ./... -v

# Benchmark testlar
bench:
	go test ./internal/... -bench=. -benchmem -run=^$$ -count=1

# Load test (server ishlab turishi kerak)
load-test:
	go test ./tests/load/ -v -run Load -count=1

# Docker orqali ishga tushirish
docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

# Demo ma'lumotlar bilan DB ni to'ldirish
seed:
	go run ./cmd/seed

# ── Swagger ─────────────────────────────────────

# Generate swagger docs from annotations
swagger:
	swag init -g cmd/api/main.go -o docs/swagger

# ── JWT Keys ──────────────────────────────────────

# ECDSA P-256 key pair yaratish (ES256 uchun)
jwt-keys:
	@mkdir -p keys
	openssl ecparam -genkey -name prime256v1 -noout -out keys/private.pem
	openssl ec -in keys/private.pem -pubout -out keys/public.pem
	@echo "JWT keys yaratildi: keys/private.pem, keys/public.pem"

# ── Goose Migrations ──────────────────────────────

# Barcha migratsiyalarni qo'llash
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

# Oxirgi migratsiyani qaytarish
migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

# Migratsiya holatini ko'rish
migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

# Yangi migratsiya fayl yaratish (usage: make migrate-create name=create_accounts)
migrate-create:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

# Muayyan versiyagacha migrate qilish (usage: make migrate-to version=002)
migrate-to:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up-to $(version)

# Barcha migratsiyalarni qaytarish
migrate-reset:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" reset
