#!/usr/bin/env bash
# Starts Postgres and Redis via Docker, then runs the API on the host with
# `go run` — the fastest edit-run loop, since it skips the Docker image build
# entirely. The frontend's Vite dev server (localhost:5173) is already in the
# default CORS_ALLOWED_ORIGINS list, so no extra configuration is needed to
# point it at this server.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

docker compose up -d postgres redis

echo "Waiting for postgres and redis to become healthy..."
until [ "$(docker compose ps --format '{{.Health}}' postgres)" = "healthy" ] &&
      [ "$(docker compose ps --format '{{.Health}}' redis)" = "healthy" ]; do
  sleep 1
done

if [ ! -f .env ]; then
  echo "No .env found — copying .env.example. Review it before continuing."
  cp .env.example .env
fi

go run ./cmd/server
