// Seed populates the database with demo data for development and testing.
// Usage: go run ./cmd/seed/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/xbank?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	fmt.Println("Seeding XBank database...")

	seedUsers(ctx, pool)
	seedAccounts(ctx, pool)
	seedExchangeRates(ctx, pool)
	seedBeneficiaries(ctx, pool)

	fmt.Println("Seed completed successfully!")
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash)
}

func seedUsers(ctx context.Context, pool *pgxpool.Pool) {
	users := []struct {
		id, email, password, firstName, lastName, role string
	}{
		{
			id:        "00000000-0000-0000-0000-000000000001",
			email:     "admin@xbank.uz",
			password:  hashPassword("Admin@12345"),
			firstName: "Admin",
			lastName:  "XBank",
			role:      "ADMIN",
		},
		{
			id:        "00000000-0000-0000-0000-000000000002",
			email:     "teller@xbank.uz",
			password:  hashPassword("Teller@12345"),
			firstName: "Kamola",
			lastName:  "Teller",
			role:      "TELLER",
		},
		{
			id:        "00000000-0000-0000-0000-000000000010",
			email:     "ali@example.com",
			password:  hashPassword("User@12345"),
			firstName: "Ali",
			lastName:  "Valiyev",
			role:      "CUSTOMER",
		},
		{
			id:        "00000000-0000-0000-0000-000000000011",
			email:     "vali@example.com",
			password:  hashPassword("User@12345"),
			firstName: "Vali",
			lastName:  "Karimov",
			role:      "CUSTOMER",
		},
		{
			id:        "00000000-0000-0000-0000-000000000012",
			email:     "zarina@example.com",
			password:  hashPassword("User@12345"),
			firstName: "Zarina",
			lastName:  "Mirzayeva",
			role:      "CUSTOMER",
		},
	}

	for _, u := range users {
		_, err := pool.Exec(ctx,
			`INSERT INTO users (id, email, password, first_name, last_name, role, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			 ON CONFLICT (id) DO NOTHING`,
			u.id, u.email, u.password, u.firstName, u.lastName, u.role,
		)
		if err != nil {
			log.Printf("  WARN: user %s: %v", u.email, err)
		} else {
			fmt.Printf("  User: %s (%s) [%s]\n", u.email, u.firstName, u.role)
		}
	}
}

func seedAccounts(ctx context.Context, pool *pgxpool.Pool) {
	accounts := []struct {
		id, userID, accountNumber, currency string
		balance                             int64
	}{
		{
			id:            "00000000-0000-0000-0001-000000000001",
			userID:        "00000000-0000-0000-0000-000000000010",
			accountNumber: "8600100000000001",
			currency:      "UZS",
			balance:       50000000, // 500,000 UZS (in tiyin)
		},
		{
			id:            "00000000-0000-0000-0001-000000000002",
			userID:        "00000000-0000-0000-0000-000000000010",
			accountNumber: "8600100000000002",
			currency:      "USD",
			balance:       100000, // $1,000 (in cents)
		},
		{
			id:            "00000000-0000-0000-0001-000000000003",
			userID:        "00000000-0000-0000-0000-000000000011",
			accountNumber: "8600100000000003",
			currency:      "UZS",
			balance:       25000000, // 250,000 UZS
		},
		{
			id:            "00000000-0000-0000-0001-000000000004",
			userID:        "00000000-0000-0000-0000-000000000012",
			accountNumber: "8600100000000004",
			currency:      "UZS",
			balance:       10000000, // 100,000 UZS
		},
		{
			id:            "00000000-0000-0000-0001-000000000005",
			userID:        "00000000-0000-0000-0000-000000000012",
			accountNumber: "8600100000000005",
			currency:      "EUR",
			balance:       50000, // 500 EUR (in cents)
		},
	}

	for _, a := range accounts {
		_, err := pool.Exec(ctx,
			`INSERT INTO accounts (id, user_id, account_number, balance, currency, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, 'ACTIVE', NOW(), NOW())
			 ON CONFLICT (id) DO NOTHING`,
			a.id, a.userID, a.accountNumber, a.balance, a.currency,
		)
		if err != nil {
			log.Printf("  WARN: account %s: %v", a.accountNumber, err)
		} else {
			fmt.Printf("  Account: %s (%s %d)\n", a.accountNumber, a.currency, a.balance)
		}
	}
}

func seedExchangeRates(ctx context.Context, pool *pgxpool.Pool) {
	rates := []struct {
		from, to string
		buy, sell int64
	}{
		{"USD", "UZS", 1265000, 1270000},  // 1 USD = 12,650 - 12,700 UZS (in tiyin)
		{"EUR", "UZS", 1380000, 1385000},  // 1 EUR = 13,800 - 13,850 UZS
		{"USD", "EUR", 92, 93},            // 1 USD = 0.92 - 0.93 EUR (in cents)
		{"GBP", "UZS", 1600000, 1605000},  // 1 GBP = 16,000 - 16,050 UZS
		{"RUB", "UZS", 14200, 14300},      // 1 RUB = 142 - 143 UZS
	}

	now := time.Now()
	for _, r := range rates {
		_, err := pool.Exec(ctx,
			`INSERT INTO exchange_rates (from_currency, to_currency, buy_rate, sell_rate, updated_at)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (from_currency, to_currency) DO UPDATE
			 SET buy_rate = EXCLUDED.buy_rate, sell_rate = EXCLUDED.sell_rate, updated_at = EXCLUDED.updated_at`,
			r.from, r.to, r.buy, r.sell, now,
		)
		if err != nil {
			log.Printf("  WARN: rate %s/%s: %v", r.from, r.to, err)
		} else {
			fmt.Printf("  Rate: %s/%s buy=%d sell=%d\n", r.from, r.to, r.buy, r.sell)
		}
	}
}

func seedBeneficiaries(ctx context.Context, pool *pgxpool.Pool) {
	beneficiaries := []struct {
		id, userID, name, beneficiaryType, accountNumber, currency string
	}{
		{
			id:            "00000000-0000-0000-0002-000000000001",
			userID:        "00000000-0000-0000-0000-000000000010",
			name:          "Vali Karimov",
			beneficiaryType: "DOMESTIC_ACCOUNT",
			accountNumber: "8600100000000003",
			currency:      "UZS",
		},
		{
			id:            "00000000-0000-0000-0002-000000000002",
			userID:        "00000000-0000-0000-0000-000000000011",
			name:          "Ali Valiyev",
			beneficiaryType: "DOMESTIC_ACCOUNT",
			accountNumber: "8600100000000001",
			currency:      "UZS",
		},
	}

	for _, b := range beneficiaries {
		_, err := pool.Exec(ctx,
			`INSERT INTO beneficiaries (id, user_id, name, type, account_number, currency, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW())
			 ON CONFLICT (id) DO NOTHING`,
			b.id, b.userID, b.name, b.beneficiaryType, b.accountNumber, b.currency,
		)
		if err != nil {
			log.Printf("  WARN: beneficiary %s: %v", b.name, err)
		} else {
			fmt.Printf("  Beneficiary: %s → %s\n", b.name, b.accountNumber)
		}
	}
}
