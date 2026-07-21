// Package database owns connection lifecycle, migrations and seeding.
package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/myntra/bharat-ai-backend/internal/config"
)

// NewPostgres opens a pooled GORM connection and verifies it is actually
// reachable.
//
// GORM's Open is lazy — it returns a handle without touching the network — so
// without the explicit Ping below, a wrong DATABASE_URL would boot cleanly and
// then fail on the first request. Failing at startup is the whole point of
// having a startup.
func NewPostgres(ctx context.Context, cfg config.DatabaseConfig, log *zap.Logger, production bool) (*gorm.DB, error) {
	level := gormlogger.Info
	if production {
		// Query logs in production are noise at best and a PII leak at worst.
		level = gormlogger.Warn
	}

	gormLog := gormlogger.New(
		zap.NewStdLog(log.Named("gorm")),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true, // A 404 is not an error worth logging.
			Colorful:                  !production,
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.URL), &gorm.Config{
		Logger: gormLog,
		// The application generates its own UUID primary keys, so there is
		// nothing to read back after an insert.
		SkipDefaultTransaction: true,
		NowFunc:                func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("database: opening postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: retrieving sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	// Bounded lifetime so connections rotate through a load balancer or a
	// failed-over primary instead of pinning to a dead host forever.
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("database: postgres unreachable: %w", err)
	}

	log.Info("postgres connected",
		zap.Int("maxOpenConns", cfg.MaxOpenConns),
		zap.Int("maxIdleConns", cfg.MaxIdleConns))

	return db, nil
}

// ClosePostgres releases the pool.
func ClosePostgres(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health reports whether the database answers within the context deadline.
func Health(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database: no connection")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
