// Command seed populates the catalogue outside of the server's own boot-time
// AutoSeed path — most commonly to reseed a database the server is already
// running against, since the seeder is idempotent (see database/seed/seed.go)
// and safe to re-run at any time.
//
// Usage:
//
//	go run ./cmd/seed
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/myntra/bharat-ai-backend/database"
	"github.com/myntra/bharat-ai-backend/database/seed"
	"github.com/myntra/bharat-ai-backend/internal/config"
	"github.com/myntra/bharat-ai-backend/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: loading config:", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Server.LogLevel, cfg.Server.IsProduction())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: building logger:", err)
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := database.NewPostgres(ctx, cfg.Database, log, cfg.Server.IsProduction())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: connecting to postgres:", err)
		os.Exit(1)
	}
	defer func() { _ = database.ClosePostgres(db) }()

	if err := seed.NewSeeder(db, log).Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "fatal: seeding:", err)
		os.Exit(1)
	}
	log.Info("seed complete")
}
