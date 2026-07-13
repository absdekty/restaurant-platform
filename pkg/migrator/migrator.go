package migrator

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Driver string

const (
	SQLite3  Driver = "sqlite3"
	Postgres Driver = "postgresql"
)

func Migrate(db *sql.DB, driver Driver, fs embed.FS, dir string) error {
	source, err := iofs.New(fs, dir)
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	var databaseDriver database.Driver

	switch driver {
	case SQLite3:
		databaseDriver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
	case Postgres:
		databaseDriver, err = postgres.WithInstance(db, &postgres.Config{})
	default:
		return fmt.Errorf("unsupported driver: %s", driver)
	}

	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, string(driver), databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
