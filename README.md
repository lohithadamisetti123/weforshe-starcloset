# StarCloset — Myntra Bharat AI

**WeForShe HackerRamp 2026 submission.**

StarCloset is an AI fashion shopping assistant built around a real,
explainable recommendation engine — and wrapped in a gamified **Luxury Fashion
District** where every celebrity stylist owns a boutique you unlock by
shopping.

| | |
|---|---|
| **Frontend** | React 19 · TypeScript · Vite 8 · Tailwind v4 · Framer Motion |
| **Backend** | Go · PostgreSQL · Redis · optional Gemini |
| **Engine** | Rule-based, 10-signal scoring pipeline — no embeddings, no vector DB |

---

## What makes it different

**A real recommendation engine, not a wrapper around an LLM call.** The engine
is a deliberately rule-based five-stage pipeline — Understand → Retrieve →
Score → Diversify → Explain — that responds in microseconds, never
hallucinates a product that isn't in stock, and can explain any ranking to a
merchandiser in plain language. Gemini is *optional*: with no API key
configured, every endpoint still works on the deterministic rule extractor
alone.

**Built for Bharat.** The prompt extractor ships a Hinglish lexicon — *shaadi,
sangeet, mehendi, haldi, diwali, navratri, garba, eid, holi* — because a
"Bharat AI" that only understands "wedding" and "festival" isn't one.
Granular festival occasions feed a cross-occasion affinity matrix, so a
sangeet query surfaces wedding wear at 0.85 affinity rather than falling back
to a generic bucket.

**A journey, not a menu.** The Stylist Journey is a vertical luxury boulevard.
Each of the five celebrity stylists owns a designer boutique that stands dark
behind frosted glass until you unlock it — then the lights come on, the doors
swing open, and the stylist's portrait appears in the entrance. Unlock all
five and the Grand Fashion Palace opens with a cinematic finale.

---

## Quick start

The one-command path — Postgres, Redis, migrations and catalogue seeding all
happen automatically:

```bash
cd backend && docker compose up
```

Then start the frontend:

```bash
cd frontend && npm install && npm run dev
```

The app runs at **http://localhost:5173**, the API at
**http://localhost:8080**. Health check: `GET /health/ready`.

### Running the backend natively

If you prefer to run Go directly, bring up only the datastores and copy the
environment template:

```bash
cd backend && docker compose up -d postgres redis && cp .env.example .env && go run ./cmd/server
```

> Postgres is mapped to host port **5433** (not 5432) to avoid colliding with
> an existing local Postgres install.

---

## Configuration

Every variable is documented in [`backend/.env.example`](backend/.env.example).
The essentials:

| Variable | Purpose |
|---|---|
| `DATABASE_URL` | Postgres connection string |
| `REDIS_URL` | Redis connection string |
| `JWT_SECRET` | Auth signing secret — **must** be overridden outside local dev |
| `GEMINI_API_KEY` | Optional. Empty means the assistant runs fully offline on the rule extractor |
| `ALLOW_GUEST` | Allows browsing without an account |

No secrets are committed: `.env` files are gitignored and only `.env.example`
placeholders are tracked.

---

## Project structure

```
backend/     Go API — recommendation engine, personas, RBAC, StarCloset progress
  internal/
    ai/                prompt understanding (rule extractor + optional Gemini merge)
    recommendation/    the engine: retrieval, scoring, diversification, explanations
    personas/          celebrity stylist roster + department configuration
    handlers/ services/ repositories/   HTTP → domain → data
  database/seed/       catalogue + migrations

frontend/    React SPA
  src/
    components/starcloset/   the Luxury Fashion District & boutiques
    components/onboarding/   mandatory profile setup + location search
    components/assistant/    the AI stylist chat
    context/                 auth, session, StarCloset journey state
```

Deeper documentation lives in [`backend/README.md`](backend/README.md) —
including the full scoring-signal breakdown — and
[`frontend/README.md`](frontend/README.md).

---

## Features

- **AI stylist chat** — natural-language and Hinglish prompts, budget-aware,
  with explanations for every recommendation
- **Collection builder** — complete head-to-toe outfits assembled within budget
- **StarCloset journey** — five celebrity stylists unlocked through shopping,
  across independent Women / Men / Kids districts
- **Personalized onboarding** — style, colour, occasion, category and location
  preferences that feed the engine
- **RBAC + partner portal** — role-gated persona onboarding for partners/admins
- **Recommendations dashboard**, saved collections, cart, and chat history

---

## Tests

```bash
cd backend && go test ./...
```

```bash
cd frontend && npm run typecheck && npm run test && npm run build
```


## 🎥 Demo Video

👉 [Watch the Demo Video](https://drive.google.com/file/d/1BXIyCdkRwKKGj3cPJ4vZM5UoxHGIwTdn/view?usp=sharing)

