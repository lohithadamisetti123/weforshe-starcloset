package database

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	// Registers the "postgres://" database driver with golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

// migrationFS embeds the SQL files into the binary.
//
// Embedding rather than reading from disk means the container image cannot
// drift from the code that expects it: there is no path to misconfigure, no
// volume to forget to mount, and `go run cmd/server/main.go` works from a fresh
// clone with nothing but a database URL. The migrations are part of the build
// artefact, which is what they logically are.
//
//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrate applies all pending migrations.
func Migrate(databaseURL string, log *zap.Logger) error {
	src, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("database: loading embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, databaseURL)
	if err != nil {
		return fmt.Errorf("database: initialising migrator: %w", err)
	}
	defer func() {
		// Close returns both a source and a database error; neither is
		// actionable at this point and both are already reflected in the
		// migration outcome.
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Warn("closing migration source", zap.Error(srcErr))
		}
		if dbErr != nil {
			log.Warn("closing migration database", zap.Error(dbErr))
		}
	}()

	before, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("database: reading migration version: %w", err)
	}

	if dirty {
		// A dirty state means a previous migration died mid-flight. Continuing
		// would apply the next migration on top of a half-applied one, which is
		// how a recoverable incident becomes a corrupt schema. Stop and make a
		// human look.
		return fmt.Errorf("database: schema is dirty at version %d; "+
			"resolve manually before starting (see README, Troubleshooting)", before)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("database schema up to date", zap.Uint("version", before))
			return nil
		}
		return fmt.Errorf("database: applying migrations: %w", err)
	}

	after, _, err := m.Version()
	if err != nil {
		return fmt.Errorf("database: reading migration version after up: %w", err)
	}

	log.Info("database migrated",
		zap.Uint("from", before),
		zap.Uint("to", after))
	return nil
}

// MigrateDown rolls back every migration. Wired only into the CLI helper in
// scripts/, never into the server's boot path — an automatic teardown one
// mis-set env var away from production is not a feature.
func MigrateDown(databaseURL string, log *zap.Logger) error {
	src, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("database: loading embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, databaseURL)
	if err != nil {
		return fmt.Errorf("database: initialising migrator: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("database: rolling back migrations: %w", err)
	}

	log.Info("database rolled back")
	return nil
}
