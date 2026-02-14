package repository

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//go:generate mockgen -destination=mocks/migrator_mock.gen.go -package=mocks . Migrator
type Migrator interface {
	Up() error
	Close() (sourceErr, databaseErr error)
}

func ApplyMigrations(m Migrator) error {
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("repository.ApplyMigrations(): %w", err)
	}

	sourceErr, databaseErr := m.Close()

	if sourceErr != nil {
		return fmt.Errorf("repository.ApplyMigrations(): %w", sourceErr)
	}

	if databaseErr != nil {
		return fmt.Errorf("repository.ApplyMigrations(): %w", databaseErr)
	}

	return nil
}
