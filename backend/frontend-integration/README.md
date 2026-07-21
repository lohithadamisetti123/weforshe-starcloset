# Wiring the frontend to this backend

The frontend (`myntra.zip`) is a complete React 19 + Vite app with **no API
layer at all** — every screen reads from static fixtures in `src/data/`. This
backend was built to match those fixture shapes byte-for-byte, so wiring it up
is additive: add a fetch layer, point it at this server, and delete the mock
imports. No component, type, or route needs to change.

## 1. Point the app at the API

The dev server runs on `:5173` by default, which is already in this backend's
default CORS allow-list (see `.env.example`, `CORS_ALLOWED_ORIGINS`). Add an
env file to the frontend project:

```
# myntra/.env.local
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## 2. Identity — no login screen needed

The frontend ships no auth UI. Rather than gate every feature behind a
registration flow that doesn't exist yet, this backend accepts a
`X-Session-Id` header on every authenticated route: the first request with a
given ID mints a real guest account (a genuine `users` row, `is_guest = true`)
and every request after that resumes it. History, favourites and recommendation
replay all work immediately, with zero UI changes required.

Generate and persist an ID once per browser, and send it on every request:

```ts
// src/lib/session.ts
const KEY = "myntra-session-id";

export function getSessionId(): string {
  let id = localStorage.getItem(KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(KEY, id);
  }
  return id;
}
```

```ts
// src/lib/api.ts
import { getSessionId } from "./session";

const BASE_URL = import.meta.env.VITE_API_BASE_URL;

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      "X-Session-Id": getSessionId(),
      ...init.headers,
    },
  });
  const body = await res.json();
  if (!res.ok || body.success === false) {
    throw new Error(body?.message ?? `Request failed: ${res.status}`);
  }
  return body.data as T;
}
```

If registration/login is added later, swap the `X-Session-Id` header for
`Authorization: Bearer <accessToken>` from `POST /auth/login` — both are
accepted by the same middleware (`internal/middleware/auth.go`), so nothing
else changes.

## 3. Swap each mock for a fetch call

Every one of these is a drop-in replacement — the response shape already
matches the type the component expects.

| Frontend mock (`src/data/`)                   | Replace with                                          |
| ---------------------------------------------- | ------------------------------------------------------ |
| `products.ts` (`MOCK_PRODUCTS`)                 | `GET /products` (grid) or `GET /products/search` (search bar) |
| `mockAI.ts` (`generateMockRecommendation`)      | `POST /assistant/chat` — same `AIRecommendationResponse` shape, note the `products` key |
| `history.ts` (`MOCK_HISTORY`)                   | `GET /history`                                          |
| Dashboard stat cards                            | `GET /dashboard/stats` — same `DashboardStats` shape    |

Example, in `useChat.ts`:

```diff
- import { generateMockRecommendation } from "../data/mockAI";
+ import { apiFetch } from "../lib/api";
+ import type { AIRecommendationResponse } from "../types";

- const response = generateMockRecommendation(prompt);
+ const response = await apiFetch<AIRecommendationResponse>("/assistant/chat", {
+   method: "POST",
+   body: JSON.stringify({ prompt }),
+ });
```

## 4. Things that are additive, not breaking

The backend's response types are supersets of the frontend's mock types —
every field the frontend reads is present with the same name and shape, plus a
few extra fields (`matchReasons` on `Product`, `meta` on
`AIRecommendationResponse`, `recommendationId` for replay). Structural typing
means the frontend's existing interfaces keep compiling unchanged; the extra
fields are simply available if a component wants to use them later — for
instance, rendering `product.matchReasons` as a tooltip on the match-percent
badge, or `response.meta.relaxed` as a "we widened your budget slightly" note.

## 5. CORS in production

`CORS_ALLOWED_ORIGINS` and `CORS_ALLOW_CREDENTIALS` (see `.env.example`) must
list the frontend's real deployed origin before shipping past a laptop demo —
the default list only covers local Vite/CRA/preview ports.
