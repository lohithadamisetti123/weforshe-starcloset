#!/usr/bin/env bash
# Re-seeds the catalogue. Safe to run any number of times — see
# database/seed/seed.go, which upserts rather than inserts.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
go run ./cmd/seed
