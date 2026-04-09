// Seed populates the database with demo data for development and testing.
// Usage: go run ./cmd/seeder/
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
	seedCards(ctx, pool)
	seedTransfers(ctx, pool)
	seedLedgerEntries(ctx, pool)
	seedExchangeRates(ctx, pool)
	seedBeneficiaries(ctx, pool)
	seedKYC(ctx, pool)
	seedFraudChecks(ctx, pool)
	seedDeviceFingerprints(ctx, pool)
	seedNotifications(ctx, pool)
	seedAuditLogs(ctx, pool)
	seedFeatureFlags(ctx, pool)
	seedSiteSettings(ctx, pool)
	seedIPRules(ctx, pool)
	seedErrorCodes(ctx, pool)
	seedRBACExtras(ctx, pool)
	seedUserSettings(ctx, pool)
	seedUserContacts(ctx, pool)
	seedScheduledTransfers(ctx, pool)
	seedAnnouncements(ctx, pool)
	seedTranslations(ctx, pool)
	seedRateLimitRules(ctx, pool)
	seedIntegrations(ctx, pool)
	seedStatisticsSnapshots(ctx, pool)
	seedReconciliationRuns(ctx, pool)
	seedSystemErrors(ctx, pool)
	seedDataExports(ctx, pool)
	seedFiles(ctx, pool)
	seedEndpointHistory(ctx, pool)
	seedAppMetrics(ctx, pool)
	seedFeatureFlagRules(ctx, pool)
	seedCardTokens(ctx, pool)
	seedCardHolds(ctx, pool)

	fmt.Println("Seed completed successfully!")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash)
}

func must(ctx context.Context, pool *pgxpool.Pool, label string, query string, args ...any) {
	_, err := pool.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("  WARN: %s: %v", label, err)
	} else {
		fmt.Printf("  OK: %s\n", label)
	}
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

func seedUsers(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Users ---")

	users := []struct {
		id, email, password, firstName, lastName, role string
	}{
		// Staff
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
			lastName:  "Abdullayeva",
			role:      "TELLER",
		},
		{
			id:        "00000000-0000-0000-0000-000000000003",
			email:     "auditor@xbank.uz",
			password:  hashPassword("Auditor@12345"),
			firstName: "Bobur",
			lastName:  "Toshmatov",
			role:      "ADMIN",
		},
		// Customers
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
		{
			id:        "00000000-0000-0000-0000-000000000013",
			email:     "jamshid@example.com",
			password:  hashPassword("User@12345"),
			firstName: "Jamshid",
			lastName:  "Rakhimov",
			role:      "CUSTOMER",
		},
		{
			id:        "00000000-0000-0000-0000-000000000014",
			email:     "nilufar@example.com",
			password:  hashPassword("User@12345"),
			firstName: "Nilufar",
			lastName:  "Ergasheva",
			role:      "CUSTOMER",
		},
		// Test / blocked user
		{
			id:        "00000000-0000-0000-0000-000000000020",
			email:     "blocked@example.com",
			password:  hashPassword("User@12345"),
			firstName: "Test",
			lastName:  "Blocked",
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
			fmt.Printf("  User: %s (%s %s) [%s]\n", u.email, u.firstName, u.lastName, u.role)
		}
	}
}

// ---------------------------------------------------------------------------
// Accounts
// ---------------------------------------------------------------------------

func seedAccounts(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Accounts ---")

	accounts := []struct {
		id, userID, accountNumber, currency, status string
		balance                                     int64
	}{
		// Ali — savings UZS, checking USD
		{
			id: "00000000-0000-0000-0001-000000000001", userID: "00000000-0000-0000-0000-000000000010",
			accountNumber: "8600100000000001", currency: "UZS", balance: 50000000, status: "ACTIVE",
		},
		{
			id: "00000000-0000-0000-0001-000000000002", userID: "00000000-0000-0000-0000-000000000010",
			accountNumber: "8600100000000002", currency: "USD", balance: 100000, status: "ACTIVE",
		},
		// Vali — UZS
		{
			id: "00000000-0000-0000-0001-000000000003", userID: "00000000-0000-0000-0000-000000000011",
			accountNumber: "8600100000000003", currency: "UZS", balance: 25000000, status: "ACTIVE",
		},
		// Zarina — UZS, EUR
		{
			id: "00000000-0000-0000-0001-000000000004", userID: "00000000-0000-0000-0000-000000000012",
			accountNumber: "8600100000000004", currency: "UZS", balance: 10000000, status: "ACTIVE",
		},
		{
			id: "00000000-0000-0000-0001-000000000005", userID: "00000000-0000-0000-0000-000000000012",
			accountNumber: "8600100000000005", currency: "EUR", balance: 50000, status: "ACTIVE",
		},
		// Jamshid — business UZS, USD
		{
			id: "00000000-0000-0000-0001-000000000006", userID: "00000000-0000-0000-0000-000000000013",
			accountNumber: "8600100000000006", currency: "UZS", balance: 150000000, status: "ACTIVE",
		},
		{
			id: "00000000-0000-0000-0001-000000000007", userID: "00000000-0000-0000-0000-000000000013",
			accountNumber: "8600100000000007", currency: "USD", balance: 500000, status: "ACTIVE",
		},
		// Nilufar — UZS
		{
			id: "00000000-0000-0000-0001-000000000008", userID: "00000000-0000-0000-0000-000000000014",
			accountNumber: "8600100000000008", currency: "UZS", balance: 8500000, status: "ACTIVE",
		},
		// Blocked user — frozen account
		{
			id: "00000000-0000-0000-0001-000000000009", userID: "00000000-0000-0000-0000-000000000020",
			accountNumber: "8600100000000009", currency: "UZS", balance: 500000, status: "FROZEN",
		},
	}

	for _, a := range accounts {
		_, err := pool.Exec(ctx,
			`INSERT INTO accounts (id, user_id, account_number, balance, currency, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			 ON CONFLICT (id) DO NOTHING`,
			a.id, a.userID, a.accountNumber, a.balance, a.currency, a.status,
		)
		if err != nil {
			log.Printf("  WARN: account %s: %v", a.accountNumber, err)
		} else {
			fmt.Printf("  Account: %s (%s %d) [%s]\n", a.accountNumber, a.currency, a.balance, a.status)
		}
	}
}

// ---------------------------------------------------------------------------
// Cards
// ---------------------------------------------------------------------------

func seedCards(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Cards ---")

	pinHash := hashPassword("1234")

	cards := []struct {
		id, accountID, pan, maskedPan, cardType, status string
		expiryMonth, expiryYear                         int
	}{
		// Ali — active UZCARD debit
		{
			id: "00000000-0000-0000-0003-000000000001", accountID: "00000000-0000-0000-0001-000000000001",
			pan: "[E2EE_REDACTED]0001", maskedPan: "8600 **** **** 0001",
			cardType: "DEBIT", status: "ACTIVE", expiryMonth: 12, expiryYear: 2028,
		},
		// Ali — active Visa USD
		{
			id: "00000000-0000-0000-0003-000000000002", accountID: "00000000-0000-0000-0001-000000000002",
			pan: "[E2EE_REDACTED]0002", maskedPan: "4278 **** **** 0002",
			cardType: "DEBIT", status: "ACTIVE", expiryMonth: 6, expiryYear: 2029,
		},
		// Vali — active Humo debit
		{
			id: "00000000-0000-0000-0003-000000000003", accountID: "00000000-0000-0000-0001-000000000003",
			pan: "[E2EE_REDACTED]0003", maskedPan: "9860 **** **** 0003",
			cardType: "DEBIT", status: "ACTIVE", expiryMonth: 3, expiryYear: 2028,
		},
		// Zarina — blocked card
		{
			id: "00000000-0000-0000-0003-000000000004", accountID: "00000000-0000-0000-0001-000000000004",
			pan: "[E2EE_REDACTED]0004", maskedPan: "8600 **** **** 0004",
			cardType: "DEBIT", status: "BLOCKED", expiryMonth: 9, expiryYear: 2027,
		},
		// Zarina — expired card
		{
			id: "00000000-0000-0000-0003-000000000005", accountID: "00000000-0000-0000-0001-000000000004",
			pan: "[E2EE_REDACTED]0005", maskedPan: "8600 **** **** 0005",
			cardType: "DEBIT", status: "EXPIRED", expiryMonth: 1, expiryYear: 2025,
		},
		// Jamshid — active business debit
		{
			id: "00000000-0000-0000-0003-000000000006", accountID: "00000000-0000-0000-0001-000000000006",
			pan: "[E2EE_REDACTED]0006", maskedPan: "8600 **** **** 0006",
			cardType: "DEBIT", status: "ACTIVE", expiryMonth: 11, expiryYear: 2029,
		},
		// Nilufar — inactive (not yet activated)
		{
			id: "00000000-0000-0000-0003-000000000007", accountID: "00000000-0000-0000-0001-000000000008",
			pan: "[E2EE_REDACTED]0007", maskedPan: "8600 **** **** 0007",
			cardType: "DEBIT", status: "INACTIVE", expiryMonth: 5, expiryYear: 2030,
		},
	}

	for _, c := range cards {
		_, err := pool.Exec(ctx,
			`INSERT INTO cards (id, account_id, pan, masked_pan, pin_hash, expiry_month, expiry_year, card_type, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
			 ON CONFLICT (id) DO NOTHING`,
			c.id, c.accountID, c.pan, c.maskedPan, pinHash,
			c.expiryMonth, c.expiryYear, c.cardType, c.status,
		)
		if err != nil {
			log.Printf("  WARN: card %s: %v", c.maskedPan, err)
		} else {
			fmt.Printf("  Card: %s [%s] %s exp %02d/%d\n", c.maskedPan, c.cardType, c.status, c.expiryMonth, c.expiryYear)
		}
	}
}

// ---------------------------------------------------------------------------
// Transfers
// ---------------------------------------------------------------------------

func seedTransfers(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Transfers ---")

	now := time.Now()

	transfers := []struct {
		id, fromAccountID, toAccountID, currency, status, description, failureReason string
		amount                                                                       int64
		createdAt                                                                    time.Time
	}{
		// Completed: Ali -> Vali (1,000,000 UZS)
		{
			id: "00000000-0000-0000-0004-000000000001",
			fromAccountID: "00000000-0000-0000-0001-000000000001", toAccountID: "00000000-0000-0000-0001-000000000003",
			amount: 1000000, currency: "UZS", status: "COMPLETED",
			description: "Kvartira ijarasi uchun", createdAt: now.Add(-72 * time.Hour),
		},
		// Completed: Vali -> Zarina (500,000 UZS)
		{
			id: "00000000-0000-0000-0004-000000000002",
			fromAccountID: "00000000-0000-0000-0001-000000000003", toAccountID: "00000000-0000-0000-0001-000000000004",
			amount: 500000, currency: "UZS", status: "COMPLETED",
			description: "Oylik tolov", createdAt: now.Add(-48 * time.Hour),
		},
		// Completed: Jamshid -> Ali (2,500,000 UZS)
		{
			id: "00000000-0000-0000-0004-000000000003",
			fromAccountID: "00000000-0000-0000-0001-000000000006", toAccountID: "00000000-0000-0000-0001-000000000001",
			amount: 2500000, currency: "UZS", status: "COMPLETED",
			description: "Dasturchilik xizmati uchun", createdAt: now.Add(-24 * time.Hour),
		},
		// Pending: Ali -> Nilufar
		{
			id: "00000000-0000-0000-0004-000000000004",
			fromAccountID: "00000000-0000-0000-0001-000000000001", toAccountID: "00000000-0000-0000-0001-000000000008",
			amount: 300000, currency: "UZS", status: "PENDING",
			description: "Kitob uchun", createdAt: now.Add(-1 * time.Hour),
		},
		// Failed: Zarina -> Jamshid (insufficient funds scenario)
		{
			id: "00000000-0000-0000-0004-000000000005",
			fromAccountID: "00000000-0000-0000-0001-000000000004", toAccountID: "00000000-0000-0000-0001-000000000006",
			amount: 90000000, currency: "UZS", status: "FAILED",
			description: "Katta summa", failureReason: "Hisobda yetarli mablag' yo'q",
			createdAt: now.Add(-6 * time.Hour),
		},
		// Completed: Jamshid USD -> Ali USD ($200)
		{
			id: "00000000-0000-0000-0004-000000000006",
			fromAccountID: "00000000-0000-0000-0001-000000000007", toAccountID: "00000000-0000-0000-0001-000000000002",
			amount: 20000, currency: "USD", status: "COMPLETED",
			description: "Freelance payment", createdAt: now.Add(-12 * time.Hour),
		},
	}

	for _, t := range transfers {
		_, err := pool.Exec(ctx,
			`INSERT INTO transfers (id, from_account_id, to_account_id, amount, currency, status, description, failure_reason, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (id) DO NOTHING`,
			t.id, t.fromAccountID, t.toAccountID, t.amount, t.currency, t.status,
			t.description, t.failureReason, t.createdAt,
		)
		if err != nil {
			log.Printf("  WARN: transfer %s: %v", t.id, err)
		} else {
			fmt.Printf("  Transfer: %s %s %d [%s]\n", t.id[:8], t.currency, t.amount, t.status)
		}
	}
}

// ---------------------------------------------------------------------------
// Ledger Entries
// ---------------------------------------------------------------------------

func seedLedgerEntries(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Ledger Entries ---")

	now := time.Now()

	entries := []struct {
		id, accountID, transferID, entryType, currency string
		amount                                         int64
		createdAt                                      time.Time
	}{
		// Transfer 1: Ali -> Vali 1,000,000 UZS
		{
			id: "00000000-0000-0000-0005-000000000001", accountID: "00000000-0000-0000-0001-000000000001",
			transferID: "00000000-0000-0000-0004-000000000001", entryType: "DEBIT",
			amount: 1000000, currency: "UZS", createdAt: now.Add(-72 * time.Hour),
		},
		{
			id: "00000000-0000-0000-0005-000000000002", accountID: "00000000-0000-0000-0001-000000000003",
			transferID: "00000000-0000-0000-0004-000000000001", entryType: "CREDIT",
			amount: 1000000, currency: "UZS", createdAt: now.Add(-72 * time.Hour),
		},
		// Transfer 2: Vali -> Zarina 500,000 UZS
		{
			id: "00000000-0000-0000-0005-000000000003", accountID: "00000000-0000-0000-0001-000000000003",
			transferID: "00000000-0000-0000-0004-000000000002", entryType: "DEBIT",
			amount: 500000, currency: "UZS", createdAt: now.Add(-48 * time.Hour),
		},
		{
			id: "00000000-0000-0000-0005-000000000004", accountID: "00000000-0000-0000-0001-000000000004",
			transferID: "00000000-0000-0000-0004-000000000002", entryType: "CREDIT",
			amount: 500000, currency: "UZS", createdAt: now.Add(-48 * time.Hour),
		},
		// Transfer 3: Jamshid -> Ali 2,500,000 UZS
		{
			id: "00000000-0000-0000-0005-000000000005", accountID: "00000000-0000-0000-0001-000000000006",
			transferID: "00000000-0000-0000-0004-000000000003", entryType: "DEBIT",
			amount: 2500000, currency: "UZS", createdAt: now.Add(-24 * time.Hour),
		},
		{
			id: "00000000-0000-0000-0005-000000000006", accountID: "00000000-0000-0000-0001-000000000001",
			transferID: "00000000-0000-0000-0004-000000000003", entryType: "CREDIT",
			amount: 2500000, currency: "UZS", createdAt: now.Add(-24 * time.Hour),
		},
		// Transfer 6: Jamshid USD -> Ali USD $200
		{
			id: "00000000-0000-0000-0005-000000000007", accountID: "00000000-0000-0000-0001-000000000007",
			transferID: "00000000-0000-0000-0004-000000000006", entryType: "DEBIT",
			amount: 20000, currency: "USD", createdAt: now.Add(-12 * time.Hour),
		},
		{
			id: "00000000-0000-0000-0005-000000000008", accountID: "00000000-0000-0000-0001-000000000002",
			transferID: "00000000-0000-0000-0004-000000000006", entryType: "CREDIT",
			amount: 20000, currency: "USD", createdAt: now.Add(-12 * time.Hour),
		},
	}

	for _, e := range entries {
		_, err := pool.Exec(ctx,
			`INSERT INTO ledger_entries (id, account_id, transfer_id, entry_type, amount, currency, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (id) DO NOTHING`,
			e.id, e.accountID, e.transferID, e.entryType, e.amount, e.currency, e.createdAt,
		)
		if err != nil {
			log.Printf("  WARN: ledger %s: %v", e.id[:8], err)
		} else {
			fmt.Printf("  Ledger: %s %s %d %s\n", e.id[:8], e.entryType, e.amount, e.currency)
		}
	}
}

// ---------------------------------------------------------------------------
// Exchange Rates
// ---------------------------------------------------------------------------

func seedExchangeRates(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Exchange Rates ---")

	rates := []struct {
		from, to string
		buy, sell int64
	}{
		{"USD", "UZS", 1265000, 1270000},
		{"EUR", "UZS", 1380000, 1385000},
		{"GBP", "UZS", 1600000, 1605000},
		{"RUB", "UZS", 14200, 14300},
		{"KZT", "UZS", 2700, 2750},
		{"TRY", "UZS", 39000, 39500},
		{"CNY", "UZS", 174000, 175000},
		{"JPY", "UZS", 8600, 8700},
		{"USD", "EUR", 92, 93},
		{"USD", "GBP", 79, 80},
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

// ---------------------------------------------------------------------------
// Beneficiaries
// ---------------------------------------------------------------------------

func seedBeneficiaries(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Beneficiaries ---")

	beneficiaries := []struct {
		id, userID, name, benType, accountNumber, bankName, bankCode, currency string
		verified                                                               bool
	}{
		// Ali's beneficiaries
		{
			id: "00000000-0000-0000-0002-000000000001", userID: "00000000-0000-0000-0000-000000000010",
			name: "Vali Karimov", benType: "INTERNAL", accountNumber: "8600100000000003",
			currency: "UZS", verified: true,
		},
		{
			id: "00000000-0000-0000-0002-000000000002", userID: "00000000-0000-0000-0000-000000000010",
			name: "Jamshid Rakhimov", benType: "INTERNAL", accountNumber: "8600100000000006",
			currency: "UZS", verified: true,
		},
		{
			id: "00000000-0000-0000-0002-000000000003", userID: "00000000-0000-0000-0000-000000000010",
			name: "Toshkent Elektr Tarmoqlari", benType: "EXTERNAL",
			accountNumber: "20208000900100001010", bankName: "Milliy Bank", bankCode: "01158",
			currency: "UZS", verified: true,
		},
		// Vali's beneficiaries
		{
			id: "00000000-0000-0000-0002-000000000004", userID: "00000000-0000-0000-0000-000000000011",
			name: "Ali Valiyev", benType: "INTERNAL", accountNumber: "8600100000000001",
			currency: "UZS", verified: true,
		},
		// Jamshid's beneficiaries
		{
			id: "00000000-0000-0000-0002-000000000005", userID: "00000000-0000-0000-0000-000000000013",
			name: "ООО Silk Road Trading", benType: "EXTERNAL",
			accountNumber: "20208000300100045012", bankName: "Asaka Bank", bankCode: "00873",
			currency: "UZS", verified: true,
		},
		{
			id: "00000000-0000-0000-0002-000000000006", userID: "00000000-0000-0000-0000-000000000013",
			name: "Nilufar Ergasheva", benType: "INTERNAL", accountNumber: "8600100000000008",
			currency: "UZS", verified: false,
		},
	}

	for _, b := range beneficiaries {
		_, err := pool.Exec(ctx,
			`INSERT INTO beneficiaries (id, user_id, name, ben_type, account_number, bank_name, bank_code, currency, verified, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			 ON CONFLICT (id) DO NOTHING`,
			b.id, b.userID, b.name, b.benType, b.accountNumber, b.bankName, b.bankCode, b.currency, b.verified,
		)
		if err != nil {
			log.Printf("  WARN: beneficiary %s: %v", b.name, err)
		} else {
			fmt.Printf("  Beneficiary: %s -> %s [%s]\n", b.name, b.accountNumber, b.benType)
		}
	}
}

// ---------------------------------------------------------------------------
// KYC Verifications
// ---------------------------------------------------------------------------

func seedKYC(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- KYC Verifications ---")

	kycs := []struct {
		id, userID, docType, docNumber, firstName, lastName, dob, status, rejectedReason string
		reviewedBy                                                                       *string
	}{
		{
			id: "00000000-0000-0000-0006-000000000001", userID: "00000000-0000-0000-0000-000000000010",
			docType: "PASSPORT", docNumber: "AA1234567", firstName: "Ali", lastName: "Valiyev",
			dob: "1990-05-15", status: "APPROVED",
		},
		{
			id: "00000000-0000-0000-0006-000000000002", userID: "00000000-0000-0000-0000-000000000011",
			docType: "PASSPORT", docNumber: "AB7654321", firstName: "Vali", lastName: "Karimov",
			dob: "1988-11-22", status: "APPROVED",
		},
		{
			id: "00000000-0000-0000-0006-000000000003", userID: "00000000-0000-0000-0000-000000000012",
			docType: "ID_CARD", docNumber: "AC1112233", firstName: "Zarina", lastName: "Mirzayeva",
			dob: "1995-03-08", status: "PENDING",
		},
		{
			id: "00000000-0000-0000-0006-000000000004", userID: "00000000-0000-0000-0000-000000000013",
			docType: "PASSPORT", docNumber: "AD4455667", firstName: "Jamshid", lastName: "Rakhimov",
			dob: "1985-07-30", status: "APPROVED",
		},
		{
			id: "00000000-0000-0000-0006-000000000005", userID: "00000000-0000-0000-0000-000000000014",
			docType: "PASSPORT", docNumber: "AE9988776", firstName: "Nilufar", lastName: "Ergasheva",
			dob: "1992-12-01", status: "PENDING",
		},
		{
			id: "00000000-0000-0000-0006-000000000006", userID: "00000000-0000-0000-0000-000000000020",
			docType: "PASSPORT", docNumber: "AF0000001", firstName: "Test", lastName: "Blocked",
			dob: "2000-01-01", status: "REJECTED", rejectedReason: "Hujjat yaroqsiz — muddati o'tgan",
		},
	}

	adminID := "00000000-0000-0000-0000-000000000001"
	for _, k := range kycs {
		reviewedBy := (*string)(nil)
		if k.status == "APPROVED" || k.status == "REJECTED" {
			reviewedBy = &adminID
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO kyc_verifications (id, user_id, document_type, document_number, first_name, last_name, date_of_birth, status, rejected_reason, reviewed_by, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
			 ON CONFLICT (id) DO NOTHING`,
			k.id, k.userID, k.docType, k.docNumber, k.firstName, k.lastName, k.dob, k.status, k.rejectedReason, reviewedBy,
		)
		if err != nil {
			log.Printf("  WARN: kyc %s: %v", k.firstName, err)
		} else {
			fmt.Printf("  KYC: %s %s [%s] %s\n", k.firstName, k.lastName, k.status, k.docType)
		}
	}
}

// ---------------------------------------------------------------------------
// Fraud Checks
// ---------------------------------------------------------------------------

func seedFraudChecks(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Fraud Checks ---")

	checks := []struct {
		id, transferID, userID, riskLevel, action string
		riskScore                                 int
		flags                                     []string
	}{
		{
			id: "00000000-0000-0000-0007-000000000001",
			transferID: "00000000-0000-0000-0004-000000000001", userID: "00000000-0000-0000-0000-000000000010",
			riskScore: 5, riskLevel: "LOW", action: "APPROVE", flags: []string{},
		},
		{
			id: "00000000-0000-0000-0007-000000000002",
			transferID: "00000000-0000-0000-0004-000000000003", userID: "00000000-0000-0000-0000-000000000013",
			riskScore: 25, riskLevel: "MEDIUM", action: "APPROVE", flags: []string{"LARGE_AMOUNT"},
		},
		{
			id: "00000000-0000-0000-0007-000000000003",
			transferID: "00000000-0000-0000-0004-000000000005", userID: "00000000-0000-0000-0000-000000000012",
			riskScore: 75, riskLevel: "HIGH", action: "BLOCK",
			flags: []string{"LARGE_AMOUNT", "INSUFFICIENT_FUNDS", "UNUSUAL_PATTERN"},
		},
	}

	for _, c := range checks {
		_, err := pool.Exec(ctx,
			`INSERT INTO fraud_checks (id, transfer_id, user_id, risk_score, risk_level, action, flags, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
			 ON CONFLICT (id) DO NOTHING`,
			c.id, c.transferID, c.userID, c.riskScore, c.riskLevel, c.action, c.flags,
		)
		if err != nil {
			log.Printf("  WARN: fraud check %s: %v", c.id[:8], err)
		} else {
			fmt.Printf("  Fraud: transfer=%s risk=%d [%s] -> %s\n", c.transferID[:8], c.riskScore, c.riskLevel, c.action)
		}
	}
}

// ---------------------------------------------------------------------------
// Device Fingerprints
// ---------------------------------------------------------------------------

func seedDeviceFingerprints(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Device Fingerprints ---")

	devices := []struct {
		id, userID, deviceHash, deviceName, ipAddress string
		trusted                                       bool
	}{
		{
			id: "00000000-0000-0000-0008-000000000001", userID: "00000000-0000-0000-0000-000000000010",
			deviceHash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			deviceName: "Samsung Galaxy S24", ipAddress: "195.158.1.100", trusted: true,
		},
		{
			id: "00000000-0000-0000-0008-000000000002", userID: "00000000-0000-0000-0000-000000000010",
			deviceHash: "f6e5d4c3b2a1f6e5d4c3b2a1f6e5d4c3b2a1f6e5d4c3b2a1f6e5d4c3b2a1f6e5",
			deviceName: "MacBook Pro (Chrome)", ipAddress: "195.158.1.101", trusted: true,
		},
		{
			id: "00000000-0000-0000-0008-000000000003", userID: "00000000-0000-0000-0000-000000000011",
			deviceHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			deviceName: "iPhone 15 Pro", ipAddress: "213.230.64.50", trusted: true,
		},
		{
			id: "00000000-0000-0000-0008-000000000004", userID: "00000000-0000-0000-0000-000000000013",
			deviceHash: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			deviceName: "Huawei P60 Pro", ipAddress: "195.158.2.200", trusted: false,
		},
	}

	for _, d := range devices {
		_, err := pool.Exec(ctx,
			`INSERT INTO device_fingerprints (id, user_id, device_hash, device_name, ip_address, trusted, last_used_at, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			 ON CONFLICT (id) DO NOTHING`,
			d.id, d.userID, d.deviceHash, d.deviceName, d.ipAddress, d.trusted,
		)
		if err != nil {
			log.Printf("  WARN: device %s: %v", d.deviceName, err)
		} else {
			fmt.Printf("  Device: %s [%s] trusted=%v\n", d.deviceName, d.ipAddress, d.trusted)
		}
	}
}

// ---------------------------------------------------------------------------
// Notifications
// ---------------------------------------------------------------------------

func seedNotifications(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Notifications ---")

	now := time.Now()

	notifications := []struct {
		id, userID, title, message, nType, data string
		readAt                                  *time.Time
		createdAt                               time.Time
	}{
		{
			id: "00000000-0000-0000-0009-000000000001", userID: "00000000-0000-0000-0000-000000000010",
			title: "Pul o'tkazma muvaffaqiyatli", message: "1,000,000 UZS Vali Karimovga yuborildi.",
			nType: "TRANSFER", data: `{"transfer_id":"00000000-0000-0000-0004-000000000001"}`,
			createdAt: now.Add(-72 * time.Hour),
		},
		{
			id: "00000000-0000-0000-0009-000000000002", userID: "00000000-0000-0000-0000-000000000010",
			title: "Yangi qurilma aniqlandi", message: "MacBook Pro orqali tizimga kirildi. Agar bu siz bo'lmasangiz, darhol parolingizni o'zgartiring.",
			nType: "SECURITY", data: `{"device_id":"00000000-0000-0000-0008-000000000002"}`,
			createdAt: now.Add(-48 * time.Hour),
		},
		{
			id: "00000000-0000-0000-0009-000000000003", userID: "00000000-0000-0000-0000-000000000011",
			title: "Pul qabul qilindi", message: "1,000,000 UZS Ali Valiyevdan qabul qilindi.",
			nType: "TRANSFER", data: `{"transfer_id":"00000000-0000-0000-0004-000000000001"}`,
			createdAt: now.Add(-72 * time.Hour),
		},
		{
			id: "00000000-0000-0000-0009-000000000004", userID: "00000000-0000-0000-0000-000000000012",
			title: "Karta bloklandi", message: "8600 **** **** 0004 raqamli kartangiz xavfsizlik sababli bloklandi.",
			nType: "ALERT", data: `{"card_id":"00000000-0000-0000-0003-000000000004"}`,
			createdAt: now.Add(-24 * time.Hour),
		},
		{
			id: "00000000-0000-0000-0009-000000000005", userID: "00000000-0000-0000-0000-000000000013",
			title: "KYC tasdiqlandi", message: "Shaxsingiz muvaffaqiyatli tasdiqlandi. Barcha xizmatlar faol.",
			nType: "INFO", data: `{}`,
			createdAt: now.Add(-120 * time.Hour),
		},
		{
			id: "00000000-0000-0000-0009-000000000006", userID: "00000000-0000-0000-0000-000000000014",
			title: "XBank-ga xush kelibsiz!", message: "Hisobingiz yaratildi. KYC tekshiruvini yakunlang.",
			nType: "INFO", data: `{}`,
			createdAt: now.Add(-168 * time.Hour),
		},
	}

	for _, n := range notifications {
		_, err := pool.Exec(ctx,
			`INSERT INTO notifications (id, user_id, title, message, type, data, read_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $8)
			 ON CONFLICT (id) DO NOTHING`,
			n.id, n.userID, n.title, n.message, n.nType, n.data, n.readAt, n.createdAt,
		)
		if err != nil {
			log.Printf("  WARN: notification %s: %v", n.title, err)
		} else {
			fmt.Printf("  Notification: [%s] %s\n", n.nType, n.title)
		}
	}
}

// ---------------------------------------------------------------------------
// Audit Logs
// ---------------------------------------------------------------------------

func seedAuditLogs(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Audit Logs ---")

	now := time.Now()

	logs := []struct {
		aggregateType, aggregateID, action, actorID, attributes, ipAddress, userAgent string
		createdAt                                                                     time.Time
	}{
		{
			aggregateType: "user", aggregateID: "00000000-0000-0000-0000-000000000010",
			action: "user.login", actorID: "00000000-0000-0000-0000-000000000010",
			attributes: `{"method":"password","device":"Samsung Galaxy S24"}`,
			ipAddress: "195.158.1.100", userAgent: "XBank-Android/2.1.0",
			createdAt: now.Add(-24 * time.Hour),
		},
		{
			aggregateType: "transfer", aggregateID: "00000000-0000-0000-0004-000000000001",
			action: "transfer.created", actorID: "00000000-0000-0000-0000-000000000010",
			attributes: `{"amount":1000000,"currency":"UZS","to_account":"8600100000000003"}`,
			ipAddress: "195.158.1.100", userAgent: "XBank-Android/2.1.0",
			createdAt: now.Add(-72 * time.Hour),
		},
		{
			aggregateType: "transfer", aggregateID: "00000000-0000-0000-0004-000000000001",
			action: "transfer.completed", actorID: "system",
			attributes: `{"amount":1000000,"currency":"UZS"}`,
			ipAddress: "", userAgent: "system",
			createdAt: now.Add(-72 * time.Hour),
		},
		{
			aggregateType: "card", aggregateID: "00000000-0000-0000-0003-000000000004",
			action: "card.blocked", actorID: "00000000-0000-0000-0000-000000000001",
			attributes: `{"reason":"suspicious_activity","masked_pan":"8600 **** **** 0004"}`,
			ipAddress: "10.0.0.1", userAgent: "XBank-Admin/1.0",
			createdAt: now.Add(-24 * time.Hour),
		},
		{
			aggregateType: "kyc", aggregateID: "00000000-0000-0000-0006-000000000006",
			action: "kyc.rejected", actorID: "00000000-0000-0000-0000-000000000001",
			attributes: `{"reason":"expired_document","user_id":"00000000-0000-0000-0000-000000000020"}`,
			ipAddress: "10.0.0.1", userAgent: "XBank-Admin/1.0",
			createdAt: now.Add(-48 * time.Hour),
		},
		{
			aggregateType: "user", aggregateID: "00000000-0000-0000-0000-000000000001",
			action: "admin.login", actorID: "00000000-0000-0000-0000-000000000001",
			attributes: `{"method":"password","2fa":"totp"}`,
			ipAddress: "10.0.0.1", userAgent: "Mozilla/5.0 Chrome/125.0",
			createdAt: now.Add(-1 * time.Hour),
		},
		{
			aggregateType: "account", aggregateID: "00000000-0000-0000-0001-000000000009",
			action: "account.frozen", actorID: "00000000-0000-0000-0000-000000000001",
			attributes: `{"reason":"fraud_investigation","account_number":"8600100000000009"}`,
			ipAddress: "10.0.0.1", userAgent: "XBank-Admin/1.0",
			createdAt: now.Add(-12 * time.Hour),
		},
	}

	for _, l := range logs {
		_, err := pool.Exec(ctx,
			`INSERT INTO audit_logs (aggregate_type, aggregate_id, action, actor_id, attributes, ip_address, user_agent, created_at)
			 VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8)`,
			l.aggregateType, l.aggregateID, l.action, l.actorID, l.attributes, l.ipAddress, l.userAgent, l.createdAt,
		)
		if err != nil {
			log.Printf("  WARN: audit %s: %v", l.action, err)
		} else {
			fmt.Printf("  Audit: [%s] %s by %s\n", l.aggregateType, l.action, l.actorID[:8])
		}
	}
}

// ---------------------------------------------------------------------------
// Feature Flags
// ---------------------------------------------------------------------------

func seedFeatureFlags(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Feature Flags ---")

	flags := []struct {
		key, description, flagType, defaultValue string
		active                                   bool
		rolloutPct                               int
	}{
		{"international_transfers", "Xalqaro pul o'tkazmalarini yoqish", "bool", "false", false, 0},
		{"card_tokenization", "Karta tokenizatsiyasi (Apple Pay/Google Pay)", "bool", "true", true, 100},
		{"biometric_login", "Biometrik tizimga kirish", "bool", "true", true, 100},
		{"scheduled_transfers", "Rejalashtirilgan pul o'tkazmalar", "bool", "true", true, 80},
		{"dark_mode", "Qorong'u rejim (UI)", "bool", "true", true, 100},
		{"new_dashboard", "Yangi boshqaruv paneli", "bool", "false", true, 25},
		{"push_notifications", "Push bildirishnomalar", "bool", "true", true, 100},
		{"qr_payments", "QR-kod orqali to'lov", "bool", "false", true, 10},
		{"maintenance_mode", "Texnik xizmat rejimi", "bool", "false", false, 0},
		{"max_transfer_amount", "Bir martalik o'tkazma limiti (UZS tiyinda)", "number", "50000000", true, 100},
	}

	for _, f := range flags {
		_, err := pool.Exec(ctx,
			`INSERT INTO feature_flags (key, description, flag_type, default_value, active, rollout_pct, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			 ON CONFLICT (key) DO NOTHING`,
			f.key, f.description, f.flagType, f.defaultValue, f.active, f.rolloutPct,
		)
		if err != nil {
			log.Printf("  WARN: feature flag %s: %v", f.key, err)
		} else {
			fmt.Printf("  Flag: %s active=%v rollout=%d%%\n", f.key, f.active, f.rolloutPct)
		}
	}
}

// ---------------------------------------------------------------------------
// Site Settings
// ---------------------------------------------------------------------------

func seedSiteSettings(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Site Settings ---")

	settings := []struct {
		key, value, settingType, description string
	}{
		{"app.name", "XBank", "general", "Ilova nomi"},
		{"app.version", "2.1.0", "general", "Joriy versiya"},
		{"app.default_currency", "UZS", "general", "Standart valyuta"},
		{"app.default_language", "uz", "general", "Standart til"},
		{"app.support_phone", "+998712000000", "contact", "Qo'llab-quvvatlash telefon raqami"},
		{"app.support_email", "support@xbank.uz", "contact", "Qo'llab-quvvatlash email"},
		{"transfer.daily_limit_uzs", "100000000", "transfer", "Kunlik o'tkazma limiti (UZS tiyinda)"},
		{"transfer.daily_limit_usd", "500000", "transfer", "Kunlik o'tkazma limiti (USD sentda)"},
		{"transfer.max_single_uzs", "50000000", "transfer", "Bir martalik o'tkazma limiti (UZS)"},
		{"kyc.auto_approve_threshold", "70", "kyc", "KYC avto-tasdiqlash chegarasi (%)"},
		{"security.max_login_attempts", "5", "security", "Maksimal kirish urinishlari"},
		{"security.lockout_duration_min", "30", "security", "Bloklash davomiyligi (daqiqa)"},
		{"security.session_ttl_hours", "24", "security", "Sessiya amal qilish muddati (soat)"},
		{"security.2fa_required_admin", "true", "security", "Admin uchun 2FA majburiy"},
		{"card.max_per_account", "3", "card", "Hisob uchun maksimal kartalar soni"},
		{"card.default_daily_limit", "5000000", "card", "Karta kunlik limiti (UZS tiyinda)"},
		{"notification.email_enabled", "true", "notification", "Email bildirishnomalari"},
		{"notification.sms_enabled", "true", "notification", "SMS bildirishnomalari"},
		{"notification.push_enabled", "true", "notification", "Push bildirishnomalari"},
	}

	for _, s := range settings {
		_, err := pool.Exec(ctx,
			`INSERT INTO site_settings (key, value, setting_type, description, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, NOW(), NOW())
			 ON CONFLICT (key) DO NOTHING`,
			s.key, s.value, s.settingType, s.description,
		)
		if err != nil {
			log.Printf("  WARN: setting %s: %v", s.key, err)
		} else {
			fmt.Printf("  Setting: [%s] %s = %s\n", s.settingType, s.key, s.value)
		}
	}
}

// ---------------------------------------------------------------------------
// IP Rules
// ---------------------------------------------------------------------------

func seedIPRules(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- IP Rules ---")

	rules := []struct {
		ipAddress, ruleType, reason, createdBy string
	}{
		// Whitelist: office and VPN IPs
		{"10.0.0.0/8", "ALLOW", "Ichki tarmoq (ofis)", "admin@xbank.uz"},
		{"195.158.1.0/24", "ALLOW", "Tashqi ofis IP diapazoni", "admin@xbank.uz"},
		{"172.16.0.0/12", "ALLOW", "VPN tarmoq", "admin@xbank.uz"},
		// Blacklist: known malicious IPs
		{"45.33.32.156", "DENY", "Brute-force hujum (2026-03-15)", "admin@xbank.uz"},
		{"198.51.100.42", "DENY", "Botnet faoliyati aniqlandi", "admin@xbank.uz"},
		{"203.0.113.99", "DENY", "Firibgarlik urinishi", "admin@xbank.uz"},
	}

	for _, r := range rules {
		_, err := pool.Exec(ctx,
			`INSERT INTO ip_rules (ip_address, rule_type, reason, created_by, created_at)
			 VALUES ($1, $2, $3, $4, NOW())`,
			r.ipAddress, r.ruleType, r.reason, r.createdBy,
		)
		if err != nil {
			log.Printf("  WARN: ip rule %s: %v", r.ipAddress, err)
		} else {
			fmt.Printf("  IP Rule: %s [%s] %s\n", r.ipAddress, r.ruleType, r.reason)
		}
	}
}

// ---------------------------------------------------------------------------
// Error Codes
// ---------------------------------------------------------------------------

func seedErrorCodes(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Error Codes ---")

	codes := []struct {
		code, messageEn, messageUz, messageRu, category, severity, suggestion string
		httpStatus                                                             int
		retryable                                                              bool
	}{
		{
			code: "AUTH_001", messageEn: "Invalid credentials", messageUz: "Noto'g'ri ma'lumotlar", messageRu: "Неверные данные",
			category: "AUTH", severity: "LOW", httpStatus: 401, retryable: true,
			suggestion: "Check email and password",
		},
		{
			code: "AUTH_002", messageEn: "Account locked", messageUz: "Hisob bloklangan", messageRu: "Аккаунт заблокирован",
			category: "AUTH", severity: "MEDIUM", httpStatus: 423, retryable: false,
			suggestion: "Wait for lockout period to expire or contact support",
		},
		{
			code: "AUTH_003", messageEn: "Session expired", messageUz: "Sessiya muddati tugagan", messageRu: "Сессия истекла",
			category: "AUTH", severity: "LOW", httpStatus: 401, retryable: true,
			suggestion: "Re-authenticate with valid credentials",
		},
		{
			code: "TRF_001", messageEn: "Insufficient funds", messageUz: "Hisobda yetarli mablag' yo'q", messageRu: "Недостаточно средств",
			category: "TRANSFER", severity: "LOW", httpStatus: 422, retryable: false,
			suggestion: "Top up account balance before retrying",
		},
		{
			code: "TRF_002", messageEn: "Daily limit exceeded", messageUz: "Kunlik limit oshirildi", messageRu: "Превышен дневной лимит",
			category: "TRANSFER", severity: "MEDIUM", httpStatus: 422, retryable: false,
			suggestion: "Try again tomorrow or request limit increase",
		},
		{
			code: "TRF_003", messageEn: "Recipient account not found", messageUz: "Qabul qiluvchi hisob topilmadi", messageRu: "Счёт получателя не найден",
			category: "TRANSFER", severity: "LOW", httpStatus: 404, retryable: false,
			suggestion: "Verify recipient account number",
		},
		{
			code: "CARD_001", messageEn: "Card blocked", messageUz: "Karta bloklangan", messageRu: "Карта заблокирована",
			category: "CARD", severity: "MEDIUM", httpStatus: 403, retryable: false,
			suggestion: "Contact bank support to unblock card",
		},
		{
			code: "CARD_002", messageEn: "Card expired", messageUz: "Karta muddati tugagan", messageRu: "Срок действия карты истёк",
			category: "CARD", severity: "LOW", httpStatus: 422, retryable: false,
			suggestion: "Request a new card",
		},
		{
			code: "CARD_003", messageEn: "Invalid PIN", messageUz: "Noto'g'ri PIN-kod", messageRu: "Неверный PIN-код",
			category: "CARD", severity: "MEDIUM", httpStatus: 401, retryable: true,
			suggestion: "Check PIN and try again. 3 failed attempts will block the card",
		},
		{
			code: "KYC_001", messageEn: "KYC verification required", messageUz: "KYC tekshiruvi talab qilinadi", messageRu: "Требуется KYC верификация",
			category: "KYC", severity: "MEDIUM", httpStatus: 403, retryable: false,
			suggestion: "Complete KYC verification in the app",
		},
		{
			code: "SYS_001", messageEn: "Internal server error", messageUz: "Ichki server xatosi", messageRu: "Внутренняя ошибка сервера",
			category: "SYSTEM", severity: "CRITICAL", httpStatus: 500, retryable: true,
			suggestion: "Try again later. If issue persists, contact support",
		},
		{
			code: "SYS_002", messageEn: "Service temporarily unavailable", messageUz: "Xizmat vaqtincha mavjud emas", messageRu: "Сервис временно недоступен",
			category: "SYSTEM", severity: "HIGH", httpStatus: 503, retryable: true,
			suggestion: "Try again in a few minutes",
		},
		{
			code: "SYS_003", messageEn: "Rate limit exceeded", messageUz: "So'rovlar limiti oshirildi", messageRu: "Превышен лимит запросов",
			category: "SYSTEM", severity: "LOW", httpStatus: 429, retryable: true,
			suggestion: "Wait before sending more requests",
		},
		{
			code: "FRAUD_001", messageEn: "Transaction flagged for review", messageUz: "Tranzaksiya tekshirish uchun belgilandi", messageRu: "Транзакция отмечена для проверки",
			category: "FRAUD", severity: "HIGH", httpStatus: 422, retryable: false,
			suggestion: "Contact support if this is a legitimate transaction",
		},
	}

	for _, c := range codes {
		_, err := pool.Exec(ctx,
			`INSERT INTO error_codes (code, message_en, message_uz, message_ru, category, severity, http_status, retryable, suggestion, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
			 ON CONFLICT (code) DO NOTHING`,
			c.code, c.messageEn, c.messageUz, c.messageRu, c.category, c.severity, c.httpStatus, c.retryable, c.suggestion,
		)
		if err != nil {
			log.Printf("  WARN: error code %s: %v", c.code, err)
		} else {
			fmt.Printf("  Error: %s [%s/%s] HTTP %d\n", c.code, c.category, c.severity, c.httpStatus)
		}
	}
}

// ---------------------------------------------------------------------------
// RBAC Extras (additional permissions for new BCs)
// ---------------------------------------------------------------------------

func seedRBACExtras(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- RBAC Extras ---")

	// Add extra permissions not covered by migration 027
	perms := []struct {
		resource, action string
	}{
		{"feature_flags", "read"}, {"feature_flags", "write"},
		{"site_settings", "read"}, {"site_settings", "write"},
		{"ip_rules", "read"}, {"ip_rules", "write"},
		{"error_codes", "read"}, {"error_codes", "write"},
		{"notifications", "read"}, {"notifications", "write"},
		{"audit_logs", "read"},
		{"integrations", "read"}, {"integrations", "write"},
		{"statistics", "read"},
		{"announcements", "read"}, {"announcements", "write"},
	}

	for _, p := range perms {
		_, err := pool.Exec(ctx,
			`INSERT INTO rbac_permissions (resource, action, description, created_at)
			 VALUES ($1, $2, $3, NOW())
			 ON CONFLICT (resource, action) DO NOTHING`,
			p.resource, p.action, fmt.Sprintf("%s:%s", p.resource, p.action),
		)
		if err != nil {
			log.Printf("  WARN: perm %s:%s: %v", p.resource, p.action, err)
		} else {
			fmt.Printf("  Permission: %s:%s\n", p.resource, p.action)
		}
	}

	// Grant all new permissions to ADMIN
	must(ctx, pool, "admin policies for new perms",
		`INSERT INTO rbac_policies (role_id, permission_id, scope)
		 SELECT r.id, p.id, 'all'
		 FROM rbac_roles r, rbac_permissions p
		 WHERE r.name = 'ADMIN'
		   AND NOT EXISTS (
		       SELECT 1 FROM rbac_policies rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
		   )`)

	// TELLER gets read on audit_logs, notifications, statistics
	must(ctx, pool, "teller read policies",
		`INSERT INTO rbac_policies (role_id, permission_id, scope)
		 SELECT r.id, p.id, 'all'
		 FROM rbac_roles r, rbac_permissions p
		 WHERE r.name = 'TELLER'
		   AND p.resource IN ('audit_logs', 'notifications', 'statistics')
		   AND p.action = 'read'
		   AND NOT EXISTS (
		       SELECT 1 FROM rbac_policies rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
		   )`)

	// CUSTOMER gets read on own notifications
	must(ctx, pool, "customer notification policy",
		`INSERT INTO rbac_policies (role_id, permission_id, scope)
		 SELECT r.id, p.id, 'own'
		 FROM rbac_roles r, rbac_permissions p
		 WHERE r.name = 'CUSTOMER'
		   AND p.resource = 'notifications' AND p.action = 'read'
		   AND NOT EXISTS (
		       SELECT 1 FROM rbac_policies rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
		   )`)
}

// ---------------------------------------------------------------------------
// User Settings
// ---------------------------------------------------------------------------

func seedUserSettings(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- User Settings ---")

	settings := []struct {
		userID, key, value string
	}{
		// Ali
		{"00000000-0000-0000-0000-000000000010", "language", "uz"},
		{"00000000-0000-0000-0000-000000000010", "theme", "light"},
		{"00000000-0000-0000-0000-000000000010", "notification.push", "true"},
		{"00000000-0000-0000-0000-000000000010", "notification.email", "true"},
		{"00000000-0000-0000-0000-000000000010", "notification.sms", "false"},
		// Vali
		{"00000000-0000-0000-0000-000000000011", "language", "ru"},
		{"00000000-0000-0000-0000-000000000011", "theme", "dark"},
		{"00000000-0000-0000-0000-000000000011", "notification.push", "true"},
		// Zarina
		{"00000000-0000-0000-0000-000000000012", "language", "uz"},
		{"00000000-0000-0000-0000-000000000012", "theme", "light"},
		// Jamshid
		{"00000000-0000-0000-0000-000000000013", "language", "en"},
		{"00000000-0000-0000-0000-000000000013", "theme", "dark"},
		{"00000000-0000-0000-0000-000000000013", "notification.push", "true"},
		{"00000000-0000-0000-0000-000000000013", "notification.email", "true"},
	}

	for _, s := range settings {
		_, err := pool.Exec(ctx,
			`INSERT INTO user_settings (user_id, key, value, created_at, updated_at)
			 VALUES ($1, $2, $3, NOW(), NOW())
			 ON CONFLICT ON CONSTRAINT uq_user_settings_user_key DO NOTHING`,
			s.userID, s.key, s.value,
		)
		if err != nil {
			log.Printf("  WARN: user setting %s/%s: %v", s.userID[:8], s.key, err)
		} else {
			fmt.Printf("  UserSetting: user=%s %s=%s\n", s.userID[:8], s.key, s.value)
		}
	}
}

// ---------------------------------------------------------------------------
// User Contacts
// ---------------------------------------------------------------------------

func seedUserContacts(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- User Contacts ---")

	contacts := []struct {
		ownerID, contactID, customName string
		isBlocked                      bool
	}{
		{"00000000-0000-0000-0000-000000000010", "00000000-0000-0000-0000-000000000011", "Vali aka", false},
		{"00000000-0000-0000-0000-000000000010", "00000000-0000-0000-0000-000000000012", "Zarina", false},
		{"00000000-0000-0000-0000-000000000010", "00000000-0000-0000-0000-000000000013", "Jamshid (ish)", false},
		{"00000000-0000-0000-0000-000000000011", "00000000-0000-0000-0000-000000000010", "Ali", false},
		{"00000000-0000-0000-0000-000000000013", "00000000-0000-0000-0000-000000000010", "Ali dev", false},
		{"00000000-0000-0000-0000-000000000013", "00000000-0000-0000-0000-000000000014", "Nilufar", false},
		{"00000000-0000-0000-0000-000000000013", "00000000-0000-0000-0000-000000000020", "", true},
	}

	for _, c := range contacts {
		_, err := pool.Exec(ctx,
			`INSERT INTO user_contacts (owner_id, contact_id, custom_name, is_blocked, created_at)
			 VALUES ($1, $2, $3, $4, NOW())
			 ON CONFLICT ON CONSTRAINT uq_user_contacts_owner_contact DO NOTHING`,
			c.ownerID, c.contactID, c.customName, c.isBlocked,
		)
		if err != nil {
			log.Printf("  WARN: contact %s->%s: %v", c.ownerID[:8], c.contactID[:8], err)
		} else {
			label := c.customName
			if label == "" {
				label = "(unnamed)"
			}
			fmt.Printf("  Contact: %s -> %s [%s] blocked=%v\n", c.ownerID[:8], c.contactID[:8], label, c.isBlocked)
		}
	}
}

// ---------------------------------------------------------------------------
// Scheduled Transfers
// ---------------------------------------------------------------------------

func seedScheduledTransfers(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Scheduled Transfers ---")

	now := time.Now()

	transfers := []struct {
		id, userID, fromAccountID, toAccountID, currency, status, description string
		amount                                                                int64
		executeAt                                                             time.Time
	}{
		// Pending: Ali -> Vali, scheduled for tomorrow
		{
			id: "00000000-0000-0000-000a-000000000001",
			userID: "00000000-0000-0000-0000-000000000010",
			fromAccountID: "00000000-0000-0000-0001-000000000001", toAccountID: "00000000-0000-0000-0001-000000000003",
			amount: 2000000, currency: "UZS", status: "PENDING",
			description: "Har oylik ijara to'lovi", executeAt: now.Add(24 * time.Hour),
		},
		// Pending: Jamshid -> Nilufar, scheduled for next week
		{
			id: "00000000-0000-0000-000a-000000000002",
			userID: "00000000-0000-0000-0000-000000000013",
			fromAccountID: "00000000-0000-0000-0001-000000000006", toAccountID: "00000000-0000-0000-0001-000000000008",
			amount: 500000, currency: "UZS", status: "PENDING",
			description: "Haftalik to'lov", executeAt: now.Add(7 * 24 * time.Hour),
		},
		// Cancelled: Zarina -> Ali
		{
			id: "00000000-0000-0000-000a-000000000003",
			userID: "00000000-0000-0000-0000-000000000012",
			fromAccountID: "00000000-0000-0000-0001-000000000004", toAccountID: "00000000-0000-0000-0001-000000000001",
			amount: 750000, currency: "UZS", status: "CANCELLED",
			description: "Bekor qilingan to'lov", executeAt: now.Add(-24 * time.Hour),
		},
	}

	for _, t := range transfers {
		_, err := pool.Exec(ctx,
			`INSERT INTO scheduled_transfers (id, user_id, from_account_id, to_account_id, amount, currency, description, status, execute_at, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			 ON CONFLICT (id) DO NOTHING`,
			t.id, t.userID, t.fromAccountID, t.toAccountID, t.amount, t.currency, t.description, t.status, t.executeAt,
		)
		if err != nil {
			log.Printf("  WARN: scheduled transfer %s: %v", t.id[:8], err)
		} else {
			fmt.Printf("  Scheduled: %s %d %s [%s] at %s\n", t.id[:8], t.amount, t.currency, t.status, t.executeAt.Format("2006-01-02"))
		}
	}
}

// ---------------------------------------------------------------------------
// Announcements
// ---------------------------------------------------------------------------

func seedAnnouncements(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Announcements ---")

	now := time.Now()

	announcements := []struct {
		titleUz, titleRu, titleEn       string
		bodyUz, bodyRu, bodyEn          string
		priority                        int
		status                          string
		startDate, endDate              time.Time
	}{
		{
			titleUz: "Yangi mobil ilova versiyasi", titleRu: "Новая версия мобильного приложения", titleEn: "New Mobile App Version",
			bodyUz:  "XBank 2.1.0 versiyasi chiqdi. Yangi funksiyalar: QR to'lov, qorong'u rejim.",
			bodyRu:  "Вышла версия XBank 2.1.0. Новые функции: QR-оплата, тёмная тема.",
			bodyEn:  "XBank 2.1.0 released. New features: QR payments, dark mode.",
			priority: 1, status: "ACTIVE",
			startDate: now.Add(-48 * time.Hour), endDate: now.Add(720 * time.Hour),
		},
		{
			titleUz: "Texnik xizmat ko'rsatish", titleRu: "Техническое обслуживание", titleEn: "Scheduled Maintenance",
			bodyUz:  "2026-04-12, soat 02:00-04:00 da texnik ishlar olib boriladi. Xizmatlar vaqtincha to'xtatiladi.",
			bodyRu:  "12.04.2026 с 02:00 до 04:00 будут проведены технические работы. Сервисы будут временно недоступны.",
			bodyEn:  "Scheduled maintenance on 2026-04-12, 02:00-04:00. Services will be temporarily unavailable.",
			priority: 2, status: "DRAFT",
			startDate: now.Add(72 * time.Hour), endDate: now.Add(96 * time.Hour),
		},
		{
			titleUz: "Bayram tabrigi", titleRu: "Праздничное поздравление", titleEn: "Holiday Greetings",
			bodyUz:  "Navro'z bayrami muborak! Barcha xizmatlar odatdagidek ishlaydi.",
			bodyRu:  "С праздником Навруз! Все сервисы работают в обычном режиме.",
			bodyEn:  "Happy Navruz! All services operate normally.",
			priority: 0, status: "ARCHIVED",
			startDate: now.Add(-720 * time.Hour), endDate: now.Add(-648 * time.Hour),
		},
	}

	for _, a := range announcements {
		_, err := pool.Exec(ctx,
			`INSERT INTO announcements (title_uz, title_ru, title_en, body_uz, body_ru, body_en, priority, status, start_date, end_date, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`,
			a.titleUz, a.titleRu, a.titleEn, a.bodyUz, a.bodyRu, a.bodyEn,
			a.priority, a.status, a.startDate, a.endDate,
		)
		if err != nil {
			log.Printf("  WARN: announcement %s: %v", a.titleEn, err)
		} else {
			fmt.Printf("  Announcement: [%s] %s\n", a.status, a.titleEn)
		}
	}
}

// ---------------------------------------------------------------------------
// Translations
// ---------------------------------------------------------------------------

func seedTranslations(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Translations ---")

	translations := []struct {
		key, language, value, group string
	}{
		// Common UI
		{"common.welcome", "uz", "Xush kelibsiz", "common"},
		{"common.welcome", "ru", "Добро пожаловать", "common"},
		{"common.welcome", "en", "Welcome", "common"},
		{"common.logout", "uz", "Chiqish", "common"},
		{"common.logout", "ru", "Выход", "common"},
		{"common.logout", "en", "Logout", "common"},
		// Transfer
		{"transfer.success", "uz", "Pul muvaffaqiyatli o'tkazildi", "transfer"},
		{"transfer.success", "ru", "Перевод выполнен успешно", "transfer"},
		{"transfer.success", "en", "Transfer completed successfully", "transfer"},
		{"transfer.insufficient_funds", "uz", "Hisobda yetarli mablag' yo'q", "transfer"},
		{"transfer.insufficient_funds", "ru", "Недостаточно средств", "transfer"},
		{"transfer.insufficient_funds", "en", "Insufficient funds", "transfer"},
		// Card
		{"card.blocked", "uz", "Karta bloklangan", "card"},
		{"card.blocked", "ru", "Карта заблокирована", "card"},
		{"card.blocked", "en", "Card blocked", "card"},
	}

	for _, t := range translations {
		_, err := pool.Exec(ctx,
			`INSERT INTO translations (key, language, value, "group", created_at, updated_at)
			 VALUES ($1, $2, $3, $4, NOW(), NOW())
			 ON CONFLICT (key, language) DO NOTHING`,
			t.key, t.language, t.value, t.group,
		)
		if err != nil {
			log.Printf("  WARN: translation %s/%s: %v", t.key, t.language, err)
		} else {
			fmt.Printf("  Translation: [%s] %s = %s\n", t.language, t.key, t.value)
		}
	}
}

// ---------------------------------------------------------------------------
// Rate Limit Rules
// ---------------------------------------------------------------------------

func seedRateLimitRules(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Rate Limit Rules ---")

	rules := []struct {
		key, description  string
		maxRequests, windowSeconds int
		enabled           bool
	}{
		{"api.auth.login", "Login so'rovlari limiti", 5, 300, true},
		{"api.auth.register", "Ro'yxatdan o'tish limiti", 3, 3600, true},
		{"api.transfers.create", "Pul o'tkazma yaratish limiti", 20, 60, true},
		{"api.cards.pin_verify", "PIN tekshiruv limiti", 3, 600, true},
		{"api.general", "Umumiy API limiti", 100, 60, true},
		{"api.admin.export", "Ma'lumotlar eksport limiti", 5, 3600, true},
		{"api.otp.send", "OTP jo'natish limiti", 3, 300, true},
	}

	for _, r := range rules {
		_, err := pool.Exec(ctx,
			`INSERT INTO rate_limit_rules (key, description, max_requests, window_seconds, enabled, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			 ON CONFLICT (key) DO NOTHING`,
			r.key, r.description, r.maxRequests, r.windowSeconds, r.enabled,
		)
		if err != nil {
			log.Printf("  WARN: rate limit %s: %v", r.key, err)
		} else {
			fmt.Printf("  RateLimit: %s max=%d window=%ds\n", r.key, r.maxRequests, r.windowSeconds)
		}
	}
}

// ---------------------------------------------------------------------------
// Integrations
// ---------------------------------------------------------------------------

func seedIntegrations(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Integrations ---")

	integrations := []struct {
		name, baseURL, status, webhookURL string
	}{
		{
			name: "Paynet", baseURL: "https://api.paynet.uz/v1",
			status: "ACTIVE", webhookURL: "https://api.xbank.uz/webhooks/paynet",
		},
		{
			name: "Click", baseURL: "https://api.click.uz/v2",
			status: "ACTIVE", webhookURL: "https://api.xbank.uz/webhooks/click",
		},
		{
			name: "Payme", baseURL: "https://checkout.paycom.uz/api",
			status: "ACTIVE", webhookURL: "https://api.xbank.uz/webhooks/payme",
		},
		{
			name: "CBU Exchange", baseURL: "https://cbu.uz/uzc/arkhiv-kursov-valyut/json",
			status: "ACTIVE", webhookURL: "",
		},
		{
			name: "MyID", baseURL: "https://myid.uz/api/v1",
			status: "ACTIVE", webhookURL: "https://api.xbank.uz/webhooks/myid",
		},
		{
			name: "SMS Provider (Eskiz)", baseURL: "https://notify.eskiz.uz/api",
			status: "ACTIVE", webhookURL: "",
		},
		{
			name: "SWIFT (Test)", baseURL: "https://sandbox.swift.com/v1",
			status: "INACTIVE", webhookURL: "",
		},
	}

	for _, i := range integrations {
		_, err := pool.Exec(ctx,
			`INSERT INTO integrations (name, base_url, status, webhook_url, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, NOW(), NOW())
			 ON CONFLICT (name) DO NOTHING`,
			i.name, i.baseURL, i.status, i.webhookURL,
		)
		if err != nil {
			log.Printf("  WARN: integration %s: %v", i.name, err)
		} else {
			fmt.Printf("  Integration: %s [%s] %s\n", i.name, i.status, i.baseURL)
		}
	}
}

// ---------------------------------------------------------------------------
// Statistics Snapshots
// ---------------------------------------------------------------------------

func seedStatisticsSnapshots(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Statistics Snapshots ---")

	now := time.Now()

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		_, err := pool.Exec(ctx,
			`INSERT INTO statistics_snapshots (date, total_users, total_accounts, active_accounts, total_transfers, total_cards, pending_kyc, flagged_fraud, system_errors, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (date) DO NOTHING`,
			dateStr,
			9+i,     // total_users grows
			12+i,    // total_accounts grows
			10+i,    // active_accounts
			50+i*10, // total_transfers
			7+i,     // total_cards
			2,       // pending_kyc
			1,       // flagged_fraud
			i,       // system_errors decreasing
			date,
		)
		if err != nil {
			log.Printf("  WARN: stats %s: %v", dateStr, err)
		} else {
			fmt.Printf("  Stats: %s users=%d accounts=%d transfers=%d\n", dateStr, 9+i, 12+i, 50+i*10)
		}
	}
}

// ---------------------------------------------------------------------------
// Reconciliation Runs
// ---------------------------------------------------------------------------

func seedReconciliationRuns(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Reconciliation Runs ---")

	runs := []struct {
		userID                              string
		totalChecked, matches, mismatches   int
		status                              string
	}{
		{
			userID: "00000000-0000-0000-0000-000000000001",
			totalChecked: 150, matches: 150, mismatches: 0, status: "COMPLETED",
		},
		{
			userID: "00000000-0000-0000-0000-000000000001",
			totalChecked: 200, matches: 198, mismatches: 2, status: "COMPLETED",
		},
		{
			userID: "00000000-0000-0000-0000-000000000003",
			totalChecked: 75, matches: 75, mismatches: 0, status: "COMPLETED",
		},
	}

	for _, r := range runs {
		_, err := pool.Exec(ctx,
			`INSERT INTO reconciliation_runs (user_id, total_checked, matches, mismatches, status, created_at)
			 VALUES ($1, $2, $3, $4, $5, NOW())`,
			r.userID, r.totalChecked, r.matches, r.mismatches, r.status,
		)
		if err != nil {
			log.Printf("  WARN: reconciliation: %v", err)
		} else {
			fmt.Printf("  Reconciliation: checked=%d matches=%d mismatches=%d [%s]\n", r.totalChecked, r.matches, r.mismatches, r.status)
		}
	}
}

// ---------------------------------------------------------------------------
// System Errors
// ---------------------------------------------------------------------------

func seedSystemErrors(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- System Errors ---")

	now := time.Now()

	errors := []struct {
		code, message, severity, category, stackTrace, requestID, userID, ipAddress, path, method, resolution string
		metadata                                                                                              string
		createdAt                                                                                             time.Time
	}{
		{
			code: "SYS_001", message: "Connection pool exhausted: pgxpool max connections reached",
			severity: "CRITICAL", category: "DATABASE",
			stackTrace: "internal/infrastructure/db/pool.go:45\ninternal/application/account/query_handler.go:32",
			requestID: "req-a1b2c3d4", userID: "", ipAddress: "10.0.0.5",
			path: "/api/v1/accounts", method: "GET",
			metadata:   `{"pool_size":20,"active":20,"idle":0}`,
			resolution: "RESOLVED", createdAt: now.Add(-72 * time.Hour),
		},
		{
			code: "KAFKA_TIMEOUT", message: "Kafka producer timeout: topic xbank.transfer.completed",
			severity: "HIGH", category: "MESSAGING",
			stackTrace: "internal/infrastructure/kafka/producer.go:88\ninternal/application/transfer/command_handler.go:120",
			requestID: "req-e5f6a7b8", userID: "00000000-0000-0000-0000-000000000010", ipAddress: "195.158.1.100",
			path: "/api/v1/transfers", method: "POST",
			metadata:   `{"topic":"xbank.transfer.completed","timeout_ms":5000}`,
			resolution: "RESOLVED", createdAt: now.Add(-48 * time.Hour),
		},
		{
			code: "VAULT_UNREACHABLE", message: "HashiCorp Vault connection refused on card encryption",
			severity: "CRITICAL", category: "SECURITY",
			stackTrace: "internal/infrastructure/crypto/vault_client.go:62\ninternal/application/card/command_handler.go:55",
			requestID: "req-c9d0e1f2", userID: "00000000-0000-0000-0000-000000000013", ipAddress: "195.158.2.200",
			path: "/api/v1/cards", method: "POST",
			metadata:   `{"vault_addr":"https://vault.xbank.internal:8200","retry_count":3}`,
			resolution: "RESOLVED", createdAt: now.Add(-36 * time.Hour),
		},
		{
			code: "REDIS_SENTINEL_FAILOVER", message: "Redis sentinel triggered failover, session cache temporarily unavailable",
			severity: "HIGH", category: "CACHE",
			stackTrace: "internal/infrastructure/redis/client.go:30",
			requestID: "", userID: "", ipAddress: "10.0.0.5",
			path: "", method: "",
			metadata:   `{"sentinel":"redis-sentinel:26379","new_master":"redis-node-2:6379","failover_ms":1200}`,
			resolution: "RESOLVED", createdAt: now.Add(-24 * time.Hour),
		},
		{
			code: "OOM_PROJECTION", message: "Out of memory rebuilding account_balance_projection for partition 2026-03",
			severity: "HIGH", category: "SYSTEM",
			stackTrace: "internal/infrastructure/projection/rebuilder.go:102",
			requestID: "req-g3h4i5j6", userID: "", ipAddress: "10.0.0.5",
			path: "/admin/projections/rebuild", method: "POST",
			metadata:   `{"projection":"account_balance","partition":"2026-03","heap_mb":3840}`,
			resolution: "PENDING", createdAt: now.Add(-6 * time.Hour),
		},
		{
			code: "TLS_CERT_EXPIRY", message: "TLS certificate for api.xbank.uz expires in 7 days",
			severity: "MEDIUM", category: "INFRASTRUCTURE",
			stackTrace: "",
			requestID: "", userID: "", ipAddress: "10.0.0.1",
			path: "", method: "",
			metadata:   `{"domain":"api.xbank.uz","expires_at":"2026-04-16T00:00:00Z","issuer":"Let's Encrypt"}`,
			resolution: "PENDING", createdAt: now.Add(-2 * time.Hour),
		},
	}

	for _, e := range errors {
		resolvedAt := (*time.Time)(nil)
		resolvedBy := ""
		if e.resolution == "RESOLVED" {
			t := e.createdAt.Add(30 * time.Minute)
			resolvedAt = &t
			resolvedBy = "00000000-0000-0000-0000-000000000001"
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO system_errors (code, message, severity, category, stack_trace, request_id, user_id, ip_address, path, method, metadata, resolution, resolved_at, resolved_by, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12, $13, $14, $15, $15)`,
			e.code, e.message, e.severity, e.category, e.stackTrace, e.requestID, e.userID, e.ipAddress,
			e.path, e.method, e.metadata, e.resolution, resolvedAt, resolvedBy, e.createdAt,
		)
		if err != nil {
			log.Printf("  WARN: system error %s: %v", e.code, err)
		} else {
			fmt.Printf("  SystemError: [%s/%s] %s [%s]\n", e.category, e.severity, e.code, e.resolution)
		}
	}
}

// ---------------------------------------------------------------------------
// Data Exports
// ---------------------------------------------------------------------------

func seedDataExports(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Data Exports ---")

	now := time.Now()

	exports := []struct {
		userID, status, fileURL, errorMsg string
		createdAt                         time.Time
	}{
		{
			userID: "00000000-0000-0000-0000-000000000010", status: "COMPLETED",
			fileURL:   "https://files.xbank.uz/exports/user_00000010_20260407.csv",
			createdAt: now.Add(-48 * time.Hour),
		},
		{
			userID: "00000000-0000-0000-0000-000000000013", status: "COMPLETED",
			fileURL:   "https://files.xbank.uz/exports/user_00000013_20260408.csv",
			createdAt: now.Add(-24 * time.Hour),
		},
		{
			userID: "00000000-0000-0000-0000-000000000001", status: "COMPLETED",
			fileURL:   "https://files.xbank.uz/exports/admin_audit_20260408.csv",
			createdAt: now.Add(-12 * time.Hour),
		},
		{
			userID: "00000000-0000-0000-0000-000000000012", status: "FAILED",
			errorMsg:  "Timeout generating report: account event count exceeded threshold",
			createdAt: now.Add(-6 * time.Hour),
		},
		{
			userID: "00000000-0000-0000-0000-000000000014", status: "PENDING",
			createdAt: now.Add(-1 * time.Hour),
		},
	}

	for _, e := range exports {
		_, err := pool.Exec(ctx,
			`INSERT INTO data_exports (user_id, status, file_url, error_msg, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $5)`,
			e.userID, e.status, e.fileURL, e.errorMsg, e.createdAt,
		)
		if err != nil {
			log.Printf("  WARN: data export %s: %v", e.userID[:8], err)
		} else {
			fmt.Printf("  DataExport: user=%s [%s]\n", e.userID[:8], e.status)
		}
	}
}

// ---------------------------------------------------------------------------
// Files
// ---------------------------------------------------------------------------

func seedFiles(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Files ---")

	files := []struct {
		name, originalName, mimeType, path, url, uploadedBy string
		size                                                int64
	}{
		{
			name: "kyc_passport_ali_001.jpg", originalName: "passport_scan.jpg",
			mimeType: "image/jpeg", size: 245760,
			path: "/uploads/kyc/2026/04/kyc_passport_ali_001.jpg",
			url:  "https://files.xbank.uz/uploads/kyc/2026/04/kyc_passport_ali_001.jpg",
			uploadedBy: "00000000-0000-0000-0000-000000000010",
		},
		{
			name: "kyc_passport_vali_002.jpg", originalName: "IMG_20260322.jpg",
			mimeType: "image/jpeg", size: 312400,
			path: "/uploads/kyc/2026/03/kyc_passport_vali_002.jpg",
			url:  "https://files.xbank.uz/uploads/kyc/2026/03/kyc_passport_vali_002.jpg",
			uploadedBy: "00000000-0000-0000-0000-000000000011",
		},
		{
			name: "kyc_id_card_zarina_003.png", originalName: "id_card_front.png",
			mimeType: "image/png", size: 189500,
			path: "/uploads/kyc/2026/04/kyc_id_card_zarina_003.png",
			url:  "https://files.xbank.uz/uploads/kyc/2026/04/kyc_id_card_zarina_003.png",
			uploadedBy: "00000000-0000-0000-0000-000000000012",
		},
		{
			name: "admin_audit_report_20260408.pdf", originalName: "audit_report_20260408.pdf",
			mimeType: "application/pdf", size: 1048576,
			path: "/uploads/reports/2026/04/admin_audit_report_20260408.pdf",
			url:  "https://files.xbank.uz/uploads/reports/2026/04/admin_audit_report_20260408.pdf",
			uploadedBy: "00000000-0000-0000-0000-000000000001",
		},
		{
			name: "announcement_navruz_banner.webp", originalName: "navruz_2026.webp",
			mimeType: "image/webp", size: 524288,
			path: "/uploads/announcements/2026/03/announcement_navruz_banner.webp",
			url:  "https://files.xbank.uz/uploads/announcements/2026/03/announcement_navruz_banner.webp",
			uploadedBy: "00000000-0000-0000-0000-000000000001",
		},
	}

	for _, f := range files {
		_, err := pool.Exec(ctx,
			`INSERT INTO files (name, original_name, mime_type, size, path, url, uploaded_by, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`,
			f.name, f.originalName, f.mimeType, f.size, f.path, f.url, f.uploadedBy,
		)
		if err != nil {
			log.Printf("  WARN: file %s: %v", f.name, err)
		} else {
			fmt.Printf("  File: %s (%s, %d bytes)\n", f.name, f.mimeType, f.size)
		}
	}
}

// ---------------------------------------------------------------------------
// Endpoint History
// ---------------------------------------------------------------------------

func seedEndpointHistory(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Endpoint History ---")

	now := time.Now()

	entries := []struct {
		method, path, userID, ipAddress string
		statusCode, durationMs          int
		createdAt                       time.Time
	}{
		{"POST", "/api/v1/auth/login", "00000000-0000-0000-0000-000000000010", "195.158.1.100", 200, 120, now.Add(-24 * time.Hour)},
		{"GET", "/api/v1/accounts", "00000000-0000-0000-0000-000000000010", "195.158.1.100", 200, 45, now.Add(-23 * time.Hour)},
		{"POST", "/api/v1/transfers", "00000000-0000-0000-0000-000000000010", "195.158.1.100", 201, 350, now.Add(-22 * time.Hour)},
		{"GET", "/api/v1/cards", "00000000-0000-0000-0000-000000000010", "195.158.1.100", 200, 38, now.Add(-21 * time.Hour)},
		{"POST", "/api/v1/auth/login", "00000000-0000-0000-0000-000000000011", "213.230.64.50", 200, 115, now.Add(-20 * time.Hour)},
		{"GET", "/api/v1/accounts", "00000000-0000-0000-0000-000000000011", "213.230.64.50", 200, 42, now.Add(-19 * time.Hour)},
		{"POST", "/api/v1/transfers", "00000000-0000-0000-0000-000000000012", "195.158.3.55", 422, 180, now.Add(-18 * time.Hour)},
		{"POST", "/api/v1/auth/login", "", "45.33.32.156", 401, 95, now.Add(-17 * time.Hour)},
		{"POST", "/api/v1/auth/login", "", "45.33.32.156", 401, 88, now.Add(-17*time.Hour + 5*time.Minute)},
		{"POST", "/api/v1/auth/login", "", "45.33.32.156", 429, 12, now.Add(-17*time.Hour + 10*time.Minute)},
		{"GET", "/api/v1/exchange-rates", "00000000-0000-0000-0000-000000000013", "195.158.2.200", 200, 25, now.Add(-16 * time.Hour)},
		{"POST", "/api/v1/auth/login", "00000000-0000-0000-0000-000000000001", "10.0.0.1", 200, 230, now.Add(-12 * time.Hour)},
		{"GET", "/admin/dashboard", "00000000-0000-0000-0000-000000000001", "10.0.0.1", 200, 150, now.Add(-11 * time.Hour)},
		{"GET", "/api/v1/notifications", "00000000-0000-0000-0000-000000000010", "195.158.1.100", 200, 55, now.Add(-10 * time.Hour)},
		{"POST", "/api/v1/cards/pin/verify", "00000000-0000-0000-0000-000000000014", "195.158.4.10", 401, 65, now.Add(-8 * time.Hour)},
		{"GET", "/api/v1/beneficiaries", "00000000-0000-0000-0000-000000000013", "195.158.2.200", 200, 32, now.Add(-6 * time.Hour)},
		{"GET", "/health", "", "10.0.0.1", 200, 2, now.Add(-1 * time.Hour)},
		{"GET", "/metrics", "", "10.0.0.5", 200, 8, now.Add(-30 * time.Minute)},
	}

	for _, e := range entries {
		_, err := pool.Exec(ctx,
			`INSERT INTO endpoint_history (method, path, status_code, user_id, ip_address, duration_ms, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			e.method, e.path, e.statusCode, e.userID, e.ipAddress, e.durationMs, e.createdAt,
		)
		if err != nil {
			log.Printf("  WARN: endpoint %s %s: %v", e.method, e.path, err)
		} else {
			fmt.Printf("  Endpoint: %s %s -> %d (%dms)\n", e.method, e.path, e.statusCode, e.durationMs)
		}
	}
}

// ---------------------------------------------------------------------------
// App Metrics
// ---------------------------------------------------------------------------

func seedAppMetrics(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- App Metrics ---")

	now := time.Now()

	metrics := []struct {
		name   string
		value  float64
		labels string
		collectedAt time.Time
	}{
		{"http_requests_total", 15234, `{"method":"GET","path":"/api/v1/accounts","status":"200"}`, now.Add(-1 * time.Hour)},
		{"http_requests_total", 8421, `{"method":"POST","path":"/api/v1/transfers","status":"201"}`, now.Add(-1 * time.Hour)},
		{"http_requests_total", 342, `{"method":"POST","path":"/api/v1/auth/login","status":"401"}`, now.Add(-1 * time.Hour)},
		{"http_request_duration_seconds", 0.045, `{"method":"GET","path":"/api/v1/accounts","quantile":"0.95"}`, now.Add(-30 * time.Minute)},
		{"http_request_duration_seconds", 0.350, `{"method":"POST","path":"/api/v1/transfers","quantile":"0.95"}`, now.Add(-30 * time.Minute)},
		{"db_pool_active_connections", 12, `{"db":"primary"}`, now.Add(-15 * time.Minute)},
		{"db_pool_active_connections", 5, `{"db":"replica"}`, now.Add(-15 * time.Minute)},
		{"redis_connected_clients", 28, `{"instance":"session"}`, now.Add(-15 * time.Minute)},
		{"redis_connected_clients", 15, `{"instance":"cache"}`, now.Add(-15 * time.Minute)},
		{"kafka_consumer_lag", 0, `{"topic":"xbank.transfer.completed","group":"transfer-saga"}`, now.Add(-10 * time.Minute)},
		{"kafka_consumer_lag", 3, `{"topic":"xbank.account.credited","group":"notification-sender"}`, now.Add(-10 * time.Minute)},
		{"transfer_amount_total_uzs", 256500000, `{"status":"COMPLETED"}`, now.Add(-5 * time.Minute)},
		{"active_sessions", 47, `{}`, now.Add(-5 * time.Minute)},
		{"kyc_pending_queue_size", 2, `{}`, now.Add(-5 * time.Minute)},
		{"fraud_checks_blocked_total", 1, `{"risk_level":"HIGH"}`, now.Add(-5 * time.Minute)},
	}

	for _, m := range metrics {
		_, err := pool.Exec(ctx,
			`INSERT INTO app_metrics (name, value, labels, collected_at)
			 VALUES ($1, $2, $3::jsonb, $4)`,
			m.name, m.value, m.labels, m.collectedAt,
		)
		if err != nil {
			log.Printf("  WARN: metric %s: %v", m.name, err)
		} else {
			fmt.Printf("  Metric: %s = %.2f\n", m.name, m.value)
		}
	}
}

// ---------------------------------------------------------------------------
// Feature Flag Rule Groups + Conditions
// ---------------------------------------------------------------------------

func seedFeatureFlagRules(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Feature Flag Rules ---")

	// new_dashboard: 25% rollout, but force-enable for ADMIN role
	var flagID string
	err := pool.QueryRow(ctx,
		`SELECT id FROM feature_flags WHERE key = 'new_dashboard'`).Scan(&flagID)
	if err != nil {
		log.Printf("  WARN: feature flag 'new_dashboard' not found: %v", err)
		return
	}

	var ruleGroupID string
	err = pool.QueryRow(ctx,
		`INSERT INTO feature_flag_rule_groups (flag_id, name, priority, value)
		 VALUES ($1, 'Admin force-enable', 1, 'true')
		 ON CONFLICT DO NOTHING
		 RETURNING id`,
		flagID).Scan(&ruleGroupID)
	if err != nil {
		log.Printf("  WARN: rule group insert: %v", err)
		return
	}
	fmt.Printf("  RuleGroup: new_dashboard -> 'Admin force-enable'\n")

	_, err = pool.Exec(ctx,
		`INSERT INTO feature_flag_conditions (rule_group_id, attribute, operator, value)
		 VALUES ($1, 'user.role', 'eq', 'ADMIN')`,
		ruleGroupID)
	if err != nil {
		log.Printf("  WARN: condition insert: %v", err)
	} else {
		fmt.Printf("  Condition: user.role eq ADMIN\n")
	}

	// qr_payments: enable for specific users (beta testers)
	var qrFlagID string
	err = pool.QueryRow(ctx,
		`SELECT id FROM feature_flags WHERE key = 'qr_payments'`).Scan(&qrFlagID)
	if err != nil {
		log.Printf("  WARN: feature flag 'qr_payments' not found: %v", err)
		return
	}

	var qrRuleGroupID string
	err = pool.QueryRow(ctx,
		`INSERT INTO feature_flag_rule_groups (flag_id, name, priority, value)
		 VALUES ($1, 'Beta testers', 1, 'true')
		 ON CONFLICT DO NOTHING
		 RETURNING id`,
		qrFlagID).Scan(&qrRuleGroupID)
	if err != nil {
		log.Printf("  WARN: qr rule group insert: %v", err)
		return
	}
	fmt.Printf("  RuleGroup: qr_payments -> 'Beta testers'\n")

	_, err = pool.Exec(ctx,
		`INSERT INTO feature_flag_conditions (rule_group_id, attribute, operator, value)
		 VALUES ($1, 'user.id', 'in', '00000000-0000-0000-0000-000000000010,00000000-0000-0000-0000-000000000013')`,
		qrRuleGroupID)
	if err != nil {
		log.Printf("  WARN: qr condition insert: %v", err)
	} else {
		fmt.Printf("  Condition: user.id in [Ali, Jamshid]\n")
	}
}

// ---------------------------------------------------------------------------
// Card Tokens
// ---------------------------------------------------------------------------

func seedCardTokens(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Card Tokens ---")

	now := time.Now()

	tokens := []struct {
		token, cardID, panEncrypted, lastFour string
		expiresAt                             time.Time
		isActive                              bool
	}{
		{
			token: "tok_a1b2c3d4e5f6a7b8c9d0", cardID: "00000000-0000-0000-0003-000000000001",
			panEncrypted: "[AES256_ENCRYPTED:8600xxxxxxxxxxxx0001]", lastFour: "0001",
			expiresAt: now.Add(365 * 24 * time.Hour), isActive: true,
		},
		{
			token: "tok_e1f2a3b4c5d6e7f8a9b0", cardID: "00000000-0000-0000-0003-000000000002",
			panEncrypted: "[AES256_ENCRYPTED:4278xxxxxxxxxxxx0002]", lastFour: "0002",
			expiresAt: now.Add(365 * 24 * time.Hour), isActive: true,
		},
		{
			token: "tok_c1d2e3f4a5b6c7d8e9f0", cardID: "00000000-0000-0000-0003-000000000003",
			panEncrypted: "[AES256_ENCRYPTED:9860xxxxxxxxxxxx0003]", lastFour: "0003",
			expiresAt: now.Add(180 * 24 * time.Hour), isActive: true,
		},
		{
			token: "tok_f1a2b3c4d5e6f7a8b9c0", cardID: "00000000-0000-0000-0003-000000000006",
			panEncrypted: "[AES256_ENCRYPTED:8600xxxxxxxxxxxx0006]", lastFour: "0006",
			expiresAt: now.Add(365 * 24 * time.Hour), isActive: true,
		},
		{
			token: "tok_d1e2f3a4b5c6d7e8f9a0", cardID: "00000000-0000-0000-0003-000000000004",
			panEncrypted: "[AES256_ENCRYPTED:8600xxxxxxxxxxxx0004]", lastFour: "0004",
			expiresAt: now.Add(-30 * 24 * time.Hour), isActive: false, // blocked card, token deactivated
		},
	}

	for _, t := range tokens {
		_, err := pool.Exec(ctx,
			`INSERT INTO card_tokens (token, card_id, pan_encrypted, last_four, expires_at, is_active, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW())
			 ON CONFLICT (token) DO NOTHING`,
			t.token, t.cardID, t.panEncrypted, t.lastFour, t.expiresAt, t.isActive,
		)
		if err != nil {
			log.Printf("  WARN: card token %s: %v", t.token, err)
		} else {
			fmt.Printf("  CardToken: %s -> card ****%s active=%v\n", t.token, t.lastFour, t.isActive)
		}
	}
}

// ---------------------------------------------------------------------------
// Card Holds
// ---------------------------------------------------------------------------

func seedCardHolds(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n--- Card Holds ---")

	now := time.Now()

	holds := []struct {
		cardID, accountID, merchant, currency, status string
		amount                                        int64
		heldAt, expiresAt                             time.Time
	}{
		{
			cardID: "00000000-0000-0000-0003-000000000001", accountID: "00000000-0000-0000-0001-000000000001",
			merchant: "Korzinka.uz", amount: 350000, currency: "UZS", status: "HELD",
			heldAt: now.Add(-2 * time.Hour), expiresAt: now.Add(70 * time.Hour),
		},
		{
			cardID: "00000000-0000-0000-0003-000000000001", accountID: "00000000-0000-0000-0001-000000000001",
			merchant: "Yandex Go Tashkent", amount: 85000, currency: "UZS", status: "CAPTURED",
			heldAt: now.Add(-26 * time.Hour), expiresAt: now.Add(46 * time.Hour),
		},
		{
			cardID: "00000000-0000-0000-0003-000000000002", accountID: "00000000-0000-0000-0001-000000000002",
			merchant: "Booking.com", amount: 15000, currency: "USD", status: "HELD",
			heldAt: now.Add(-6 * time.Hour), expiresAt: now.Add(66 * time.Hour),
		},
		{
			cardID: "00000000-0000-0000-0003-000000000006", accountID: "00000000-0000-0000-0001-000000000006",
			merchant: "Macro Supermarket", amount: 1250000, currency: "UZS", status: "RELEASED",
			heldAt: now.Add(-48 * time.Hour), expiresAt: now.Add(24 * time.Hour),
		},
		{
			cardID: "00000000-0000-0000-0003-000000000003", accountID: "00000000-0000-0000-0001-000000000003",
			merchant: "Uzum Market", amount: 420000, currency: "UZS", status: "HELD",
			heldAt: now.Add(-1 * time.Hour), expiresAt: now.Add(71 * time.Hour),
		},
	}

	for _, h := range holds {
		capturedAt := (*time.Time)(nil)
		releasedAt := (*time.Time)(nil)
		if h.status == "CAPTURED" {
			t := h.heldAt.Add(4 * time.Hour)
			capturedAt = &t
		}
		if h.status == "RELEASED" {
			t := h.heldAt.Add(2 * time.Hour)
			releasedAt = &t
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO card_holds (card_id, account_id, merchant, amount, currency, status, held_at, expires_at, captured_at, released_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			h.cardID, h.accountID, h.merchant, h.amount, h.currency, h.status, h.heldAt, h.expiresAt, capturedAt, releasedAt,
		)
		if err != nil {
			log.Printf("  WARN: card hold %s: %v", h.merchant, err)
		} else {
			fmt.Printf("  CardHold: %s %d %s [%s]\n", h.merchant, h.amount, h.currency, h.status)
		}
	}
}
