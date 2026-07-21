// Command migrate applies or rolls back the database schema outside of the
// server's own boot-time AutoMigrate path.
//
// It exists for the two cases the server's automatic migration deliberately
// does not cover: rolling back (never automatic — see database.MigrateDown),
// and migrating a database the server is not currently pointed at, such as a
// staging copy.
//
// Usage:
//
//	go run ./cmd/migrate up
//	go run ./cmd/migrate down
package main

import (
	"fmt"
	"os"

	"github.com/myntra/bharat-ai-backend/database"
	"github.com/myntra/bharat-ai-backend/internal/config"
	"github.com/myntra/bharat-ai-backend/pkg/logger"
)

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down") {
		fmt.Fprintln(os.Stderr, "usage: migrate [up|down]")
		os.Exit(2)
	}

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

	switch os.Args[1] {
	case "up":
		err = database.Migrate(cfg.Database.URL, log)
	case "down":
		// No confirmation prompt: this binary is not wired into any
		// automatic path (see cmd/server/main.go, which only ever calls
		// Migrate), so reaching this line already required a human to type
		// "down" deliberately.
		err = database.MigrateDown(cfg.Database.URL, log)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
