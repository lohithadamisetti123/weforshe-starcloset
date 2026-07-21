#!/usr/bin/env bash
# Applies (or rolls back) the database schema against whatever DATABASE_URL
# resolves to — a local .env, or real environment variables.
#
# Usage:
#   ./scripts/migrate.sh up
#   ./scripts/migrate.sh down
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
go run ./cmd/migrate "${1:-up}"
