// Package main provides a standalone CLI tool for managing database schema migrations.
// It applies incremental SQL changes to the Postgres database using Goose.
//
// Usage:
//
//	go run ./cmd/migration
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const defaultMigrationsDir = "migrations"

func main() {
	dir := flag.String("dir", defaultMigrationsDir, "directory containing migration files")
	command := flag.String("command", "up", "migration command: up, down, status, reset, version")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/xbank?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	switch *command {
	case "up":
		err = goose.Up(db, *dir)
	case "down":
		err = goose.Down(db, *dir)
	case "status":
		err = goose.Status(db, *dir)
	case "reset":
		err = goose.Reset(db, *dir)
	case "version":
		err = goose.Version(db, *dir)
	default:
		log.Fatalf("Unknown command: %s (valid: up, down, status, reset, version)", *command)
	}

	if err != nil {
		log.Fatalf("Migration %s failed: %v", *command, err)
	}

	fmt.Printf("Migration %s completed successfully\n", *command)
}
