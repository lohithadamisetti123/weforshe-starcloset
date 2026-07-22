# Myntra Bharat AI — Frontend

**Your Personal AI Fashion Stylist** — built for Myntra WeForShe HackerRamp 2026.

A premium, AI-first shopping assistant: chat with an AI stylist, receive complete
slot-based outfit collections with per-item match scores and plain-English
explanations, browse a recommendations dashboard, and review your history.

The frontend is a real client for the Go backend in `../backend`. It holds no
recommendation logic of its own — ranking, outfit assembly, gender locking and
explanation all happen server-side.

## Tech stack

- React 19 + TypeScript 6 + Vite 8
- Tailwind CSS v4
- Framer Motion (animation)
- Lucide React (icons)
- React Router 7
- TanStack React Query
- React Hook Form + Zod
- Recharts (dashboard charts)

## Getting started

```bash
npm install
npm run dev
```

Open the printed local URL (usually http://localhost:5173). The backend must be
running — see `../backend/README.md`.

## Scripts

| Script | What it does |
|---|---|
| `npm run dev` | Vite dev server |
| `npm run build` | Typecheck the app, then bundle |
| `npm run typecheck` | Typecheck the app **and** the test sources |
| `npm test` | Run the unit suite (Node's built-in runner — no extra dependency) |
| `npm run lint` | oxlint |
| `npm run preview` | Serve the production bundle |

Tests run on Node's native test runner with built-in TypeScript stripping, so
there is no Vitest/Jest install to maintain. Node 22+ is required.

## Project structure

```
src/
  components/
    layout/          Navbar, Footer, page Layout wrapper
    home/            Hero, floating cards, categories, occasions
    assistant/       Chat window, bubbles, prompt input, AI loader, collections
    recommendations/ Product cards, dashboard stat cards, charts
    onboarding/      Personalization questionnaire and its gate
    history/         History table (search, filter, pagination)
    ui/              Reusable primitives: Button, Badge, Avatar, Card, Skeleton
  context/
    AuthContext      Authentication state
    SessionContext   Shopping session — React binding
    sessionState.ts  Shopping session — pure logic (unit tested)
    AppContext       Cart, favourites, saved collections, chat history
  pages/             One component per route
  data/              Static catalogue used by decorative landing-page components
  hooks/             useChat, useCart, useProfile, useFavorites, …
  types/             Shared TypeScript interfaces (mirror the backend DTOs)
  lib/               api.ts (fetch client), auth.ts (session storage), utils.ts
```

## SessionContext

`SessionContext` is the single source of truth for the shopping session:
department, gender, conversation turns, and the products already shown.

Two rules in it are load-bearing, and both are pinned by tests in
`src/context/sessionState.test.ts`:

**Gender only changes on a deliberate act.** Choosing a gendered department
(Women, Men, Kids - Girls, Kids - Boys) sets it. Choosing a gender-agnostic one
(Beauty, Footwear, Accessories) *inherits* whatever was already in force, so a
shopper who picked Women and then browsed Footwear still gets women's shoes.
This mirrors the backend, which falls back to the intent's own gender when the
department carries none. Nothing infers gender from prompt text on the client.

**Shown products accumulate across turns.** Every product the assistant surfaces
is recorded per department and sent back as `excludeProductIds` on the next
request, so the stylist does not re-offer what it just suggested. The list is
capped at 200 to match the backend's `validate:"omitempty,max=200"` — exceeding
it is a 400, not a degraded result, so the cap is a correctness constraint.

Session state is persisted to `localStorage` under a key namespaced by user id
(`myntra-session:<userId>`). The namespacing is deliberate: a browser-global key
previously let one account read another's cart, and `clearUserScopedCaches()` in
`lib/auth.ts` purges these keys on both login and logout.

The logic lives in `sessionState.ts` as pure functions over plain data;
`SessionContext.tsx` is only the React and storage plumbing. That split is what
makes the rules above testable without a renderer or a mocked DOM.

## Notes

- Product images are placeholder photography from Picsum, seeded per product so
  they stay consistent across reloads.
- `src/data/products.ts` is a static catalogue used only by decorative
  landing-page components. It is not the shopping catalogue — that comes from
  the backend.
