# syntax=docker/dockerfile:1

# --- Build stage --------------------------------------------------------
# A separate build stage means the shipped image carries a compiled binary
# and nothing else: no Go toolchain, no module cache, no source tree for
# anyone with shell access to read. It also isolates the slow part
# (downloading modules, compiling) into its own cached layer so it isn't
# repeated when only application code changes.
FROM golang:1.24-bookworm AS build

WORKDIR /src

# Dependencies first, so this layer only invalidates when go.mod/go.sum
# actually change, not on every source edit.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# `go mod tidy` before building rather than assuming go.mod/go.sum are already
# complete: this repository's go.mod was hand-written rather than produced by
# the toolchain, so it may be missing indirect requires or a go.sum entirely.
# `go build`'s default -mod=readonly refuses to proceed in that state ("go:
# updates to go.mod needed") — tidy is what actually resolves and pins the
# full dependency graph. It's a no-op on a build environment where go.mod was
# already tidy.
RUN go mod tidy

# CGO_ENABLED=0 for a fully static binary — no libc dependency, so it runs
# unmodified on the minimal Alpine base below.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/server \
    ./cmd/server

# --- Runtime stage --------------------------------------------------------
# Alpine over distroless here on purpose, even though it costs a little attack
# surface: it keeps a shell and wget available, which is what lets Docker
# Compose's HEALTHCHECK actually probe /health from inside the container, and
# what lets a reviewer `docker compose exec backend sh` when something needs
# debugging live during judging. For a longer-lived production deployment,
# swapping to a distroless base is a one-line change and nothing else in this
# file would need to move.
FROM alpine:3.20 AS runtime

# ca-certificates: outbound HTTPS to the Gemini API needs a trust store.
# wget: backs the HEALTHCHECK below.
# tzdata: unused by the app (IST is a fixed offset, see dashboard_service.go)
# but harmless and occasionally expected by dependencies' time formatting.
RUN apk add --no-cache ca-certificates wget tzdata \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=build /out/server /app/server

USER app
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=3 \
    CMD wget -qO- http://localhost:8080/health/ready || exit 1

ENTRYPOINT ["/app/server"]
