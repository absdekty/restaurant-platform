package migrator

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Driver string

const (
	SQLite3 Driver = "sqlite3"
)

func Migrate(db *sql.DB, driver Driver, fs embed.FS, dir string) error {
	source, err := iofs.New(fs, dir)
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	switch driver {
	case SQLite3:
		databaseDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("failed to create database driver: %w", err)
		}

		m, err := migrate.NewWithInstance("iofs", source, string(driver), databaseDriver)
		if err != nil {
			return fmt.Errorf("failed to create migrator: %w", err)
		}
		// defer m.Close()

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
	default:
		return fmt.Errorf("unsupported driver: %s", driver)
	}

	return nil
}
