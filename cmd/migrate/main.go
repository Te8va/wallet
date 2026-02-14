package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		dsn            string
		migrationsPath string
		command        string
		versionStr     string
	)

	flag.StringVar(&dsn, "dsn", "", "Database connection string")
	flag.StringVar(&migrationsPath, "migrations-path", "file://migrations", "Path to migrations")
	flag.StringVar(&command, "command", "up", "Migration command (up, down, force, version)")
	flag.StringVar(&versionStr, "version", "", "Version for force command")
	flag.Parse()

	if dsn == "" {
		dsn = os.Getenv("DATABASE_DSN")
	}
	if dsn == "" {
		log.Fatal("DSN is required. Use -dsn flag or DATABASE_DSN environment variable")
	}

	log.Printf("Running migration command: %s", command)

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		log.Fatalf("Failed to initialize migrations: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		log.Println("Migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
		log.Println("Migrations rolled back successfully")
	case "force":
		if versionStr == "" {
			log.Fatal("Version is required for force command (use -version flag)")
		}
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			log.Fatalf("Invalid version format: %v", err)
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("Failed to force migration version: %v", err)
		}
		log.Printf("Forced migration version to %d", version)
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				log.Println("No migrations applied yet")
			} else {
				log.Fatalf("Failed to get migration version: %v", err)
			}
		} else {
			status := "clean"
			if dirty {
				status = "dirty"
			}
			log.Printf("ℹCurrent migration version: %d, status: %s", version, status)
		}
	default:
		log.Fatalf("Unknown command: %s. Available commands: up, down, force, version", command)
	}
}
