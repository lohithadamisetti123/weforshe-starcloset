# Myntra Bharat AI — Backend

A production-shaped Go backend for **Myntra Bharat AI**, a fashion shopping
assistant built for **WeForShe HackerRamp 2026**. It serves the accompanying
React frontend with zero frontend changes required (see
[`frontend-integration/README.md`](frontend-integration/README.md)).

## The USP: a real recommendation engine, not a wrapper around an LLM call

The brief for this project was explicit that the recommendation engine is the
product's differentiator, so most of the engineering effort — and most of this
README — is about `internal/recommendation`, not about CRUD.

**The engine is deliberately rule-based: no model training, no embeddings, no
vector database.** That's a design decision, not a shortcut. With a catalogue
this size and a hard requirement that every recommendation be explainable to a
merchandiser in plain language, a well-tuned scoring function beats a black
box on relevance, on debuggability, and on latency — it responds in
microseconds and never hallucinates a product that isn't in stock.

It's a five-stage pipeline:

```
prompt ──▶ Understand ──▶ Retrieve ──▶ Score ──▶ Diversify ──▶ Explain ──▶ shortlist
           (ai/)          (hard        (10        (MMR)         (human
                           filters      weighted                 sentences,
                           only)        signals)                 from real
                                                                  scores)
```

1. **Understand** (`internal/ai`) — a deterministic rule-based extractor and,
   when configured, Gemini run **concurrently** over the prompt. Their
   readings are merged field-by-field (`internal/ai/merge.go`): budget
   arithmetic trusts the rules (LLMs round ₹2,499 to ₹2,500 more often than
   not), category/occasion/style trust Gemini's language understanding when
   it's confident, exclusions are unioned (a missed "not this" is worse than
   an extra one). Gemini is fully optional — with no API key configured, the
   assistant runs on the rule extractor alone and every endpoint still works.
   The rule extractor includes a **Hinglish lexicon** (shaadi, sangeet,
   mehendi, haldi, diwali, navratri, garba, eid, holi, ...) because a
   "Bharat AI" that only understands "wedding" and "festival" isn't one.

2. **Retrieve** (`internal/recommendation/engine.go`) — only genuinely
   disqualifying constraints are pushed into SQL (gender, hard price bounds,
   explicit exclusions). A **graceful relaxation ladder** guarantees the
   engine never returns an empty shortlist for a non-empty catalogue: it
   drops the price ceiling, then exclusions, then gender — in that order,
   because that's the order a shopper minds losing them — and reports exactly
   what it relaxed in the response, rather than silently ignoring part of the
   ask.

3. **Score** (`internal/recommendation/scoring.go`) — ten independent signals
   (occasion, style, gender, category, budget fit, color, brand, BM25 text
   relevance, a Bayesian quality prior, and a value/discount signal), each
   with a configurable weight. The key idea is **query-adaptive weighting**: a
   signal only participates if the query gave something to judge it against —
   asking for "blue" activates the color signal, asking for nothing activates
   only quality and value. Inactive weight is redistributed across active
   signals, so a score always answers "how well does this match what was
   actually asked," never "how well does this match a blank."

4. **Diversify** (`internal/recommendation/diversity.go`) — a greedy
   Maximal Marginal Relevance pass stops the shortlist collapsing into six
   near-identical products from one brand and category.

5. **Explain** (`internal/recommendation/explain.go`) — the match percentage
   and the human-readable reasons are derived from the actual score
   breakdown, not a second LLM call asked to rationalize a black box. What
   the shopper reads is what the engine computed.

Every stage is a pure function over plain data — no GORM, no HTTP — which is
what makes `internal/recommendation/engine_test.go` possible: the full
pipeline is tested against an in-memory fixture catalogue with no database.

## Quick start

### Option A — Docker (recommended, one command)

```bash
docker compose up --build
```

This starts Postgres, Redis, and the API; migrates the schema; and seeds a
52-product catalogue plus two demo accounts. The API is at
`http://localhost:8080`, with interactive docs at
`http://localhost:8080/swagger/index.html`.

### Option B — Run on the host (fastest edit loop)

```bash
cp .env.example .env
docker compose up -d postgres redis   # or scripts/dev.sh, which does this + waits for health
go mod download
go run ./cmd/server
```

> **Note on this repository's build environment:** this backend was generated
> in a sandbox with no access to the Go module proxy, so `go.sum` is not
> checked in and dependencies have never been fetched or compiled here. Run
> `go mod tidy` once you have network access, before the first
> `go build`/`go run` — everything after that works normally. The Docker build
> above handles this automatically since the build stage has network access.

### Demo accounts (seeded)

| Email              | Password        |
| ------------------- | ---------------- |
| `asha@myntra.dev`   | `bharat-ai-demo`  |
| `admin@myntra.dev`  | `bharat-ai-admin` |

The frontend has no login screen, so most flows don't need these — see
"Guest identity" below.

## API surface

Full interactive documentation: `GET /swagger/index.html` (backed by
`docs/openapi.yaml`, also fetchable directly for import into another tool). A
ready-to-run Postman collection is in `postman/`.

| Area | Endpoints |
| --- | --- |
| Health | `GET /health`, `/health/live`, `/health/ready` |
| Auth | `POST /auth/register`, `/auth/login`, `/auth/refresh`, `GET /profile` |
| Products | `GET /products`, `/products/search`, `/products/{id}`, `/products/category/{name}`, `/products/brand/{name}`, `/categories`, `/brands` |
| **Assistant** | `POST /assistant/chat`, `GET /assistant/recommendations/{id}` |
| History | `GET/POST /history`, `PATCH/DELETE /history/{id}` |
| Recommendations | `GET /recommendations`, `/recommendations/{id}` |
| Favorites | `GET/POST /favorites`, `DELETE /favorites/{id}` |
| Dashboard | `GET /dashboard/stats` |

`GET /products/search` and `POST /assistant/chat` are powered by the **same**
recommendation engine (via `dto.ProductQuery.ToIntent`) — the search bar and
the assistant never disagree about what "best" means for equivalent
constraints.

### Guest identity

The frontend ships no login UI, so every identity-scoped route also accepts
an `X-Session-Id` header instead of a bearer token. The first request with a
given ID mints a real guest account; every later request with the same ID
resumes it, with working history, favourites and recommendation replay. See
`internal/middleware/auth.go`.

## Configuration

All configuration is environment-first (see `.env.example` for the full,
commented list). The recommendation engine's signal weights are runtime
configuration, not constants — retuning relevance is an ops action:

```
REC_W_OCCASION=0.24   REC_W_BUDGET=0.18    REC_W_STYLE=0.15
REC_W_GENDER=0.10     REC_W_CATEGORY=0.09  REC_W_TEXT=0.09
REC_W_QUALITY=0.09    REC_W_COLOR=0.05     REC_W_VALUE=0.05
REC_W_BRAND=0.04
REC_DIVERSITY_LAMBDA=0.72   # MMR relevance/diversity trade-off
REC_TOP_N=6                 # shortlist size
REC_CANDIDATE_POOL=200      # max candidates scored per request
```

Weights are normalized at load time, so they can be tuned independently
without needing to keep the sum at 1.0 by hand.

## Architecture

Clean, layered architecture with dependency injection wired by hand in
`cmd/server/main.go` (no DI framework — the wiring is short enough to read
top-to-bottom):

```
internal/
├── handlers/        HTTP boundary: bind, validate, call one service, respond
├── services/         use cases: orchestration, no SQL, no HTTP
├── repositories/      persistence: GORM, Redis cache, nothing else
├── recommendation/    the ranking engine — pure functions, no I/O
├── ai/                intent extraction: rules + Gemini + merge
├── models/             GORM row types + shared vocabulary (enums matching
│                       the frontend's TypeScript unions exactly)
├── dto/                 the wire contract — the only thing ever serialized
├── middleware/           auth (incl. guest), CORS, rate limiting, recovery,
│                         structured logging, request ID, error mapping
├── config/                typed, validated, environment-first config
└── validator/              request validation → per-field error messages

database/
├── migrations/    versioned SQL (golang-migrate), CHECK constraints mirror
│                   the frontend's TypeScript unions
└── seed/           idempotent catalogue + demo-user seeding
```

**Why the layering matters here specifically:** the recommendation engine
depends on a single narrow interface (`recommendation.CandidateSource`), which
`ProductRepository` happens to implement — not the other way around. The
engine has no idea GORM or Postgres exist, which is what makes it fast to
test and safe to re-tune.

## Testing

```bash
go test ./...
```

Test coverage focuses where the brief asked for the effort: the
recommendation engine (`internal/recommendation/engine_test.go` — relevance
ranking, the never-empty guarantee under impossible constraints, hard
exclusions, match-percent bounds and monotonicity, explicit-sort override,
deterministic ranks) and the rule-based extractor
(`internal/ai/extractor_test.go` — the Hinglish occasion lexicon, budget
parsing including the "don't mistake an age for a price" guard, negation
handling, and the short-brand-name whole-word-only safety check for brands
like "W").

## Project structure reference

```
cmd/
├── server/   the API process
├── migrate/  standalone migration CLI (up/down)
└── seed/     standalone reseed CLI (idempotent, safe to re-run any time)
docs/         openapi.yaml + embedded Swagger UI
postman/      ready-to-run collection + environment
scripts/      migrate.sh, seed.sh, dev.sh
frontend-integration/   how to wire the existing React app to this API
```

## Troubleshooting

- **`schema is dirty at version N`** on boot — a previous migration died
  mid-flight. `database.Migrate` refuses to proceed automatically (see
  `database/migrate.go`); inspect the `schema_migrations` table, fix the
  underlying issue, and mark the version clean before restarting.
- **Assistant responses look identical regardless of the prompt** — check
  `GET /health/ready`: if Redis is down, every repository degrades to a
  direct database read (correct, if slower) rather than failing, but a stale
  cached response for an identical query is still served for its TTL
  (10 minutes by default). Products, categories and dashboard stats are
  cached; assistant answers are not.
- **Gemini calls failing** — the assistant is designed to degrade gracefully:
  an unconfigured or unreachable Gemini falls back to the rule extractor
  automatically, and `meta.source` in the assistant's response tells you
  which path actually ran for a given request.
