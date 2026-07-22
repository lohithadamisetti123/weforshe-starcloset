import { clearSession, getAccessToken, getRefreshToken, setTokens } from '@/lib/auth'
import type { AuthResponse } from '@/types'

const SESSION_KEY = 'myntra-session-id'

function getSessionId(): string {
  let id = localStorage.getItem(SESSION_KEY)
  if (!id) {
    id = crypto.randomUUID()
    localStorage.setItem(SESSION_KEY, id)
  }
  return id
}

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'

// buildHeaders always sends the guest session id (unchanged), and adds a Bearer
// token when the shopper is signed in. The backend prefers the bearer token
// over the guest session, so a signed-in request is authenticated as the real
// account while an anonymous one behaves exactly as it always did.
function buildHeaders(init: RequestInit, token: string | null): HeadersInit {
  return {
    'Content-Type': 'application/json',
    'X-Session-Id': getSessionId(),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...init.headers,
  }
}

// A single in-flight refresh is shared across concurrent 401s so a burst of
// requests triggers exactly one /auth/refresh, not one per request.
let refreshing: Promise<string | null> | null = null

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) return null

  if (!refreshing) {
    refreshing = (async () => {
      try {
        const res = await fetch(`${BASE_URL}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'X-Session-Id': getSessionId() },
          body: JSON.stringify({ refreshToken }),
        })
        const body = await res.json()
        if (!res.ok || body.success === false) {
          clearSession()
          return null
        }
        const data = body.data as AuthResponse
        setTokens(data.tokens)
        return data.tokens.accessToken
      } catch {
        clearSession()
        return null
      }
    })()
    // Reset the shared promise once it settles, so the next expiry can refresh
    // again.
    refreshing.finally(() => {
      refreshing = null
    })
  }
  return refreshing
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  let res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: buildHeaders(init, getAccessToken()),
  })

  // If a signed-in request is rejected because the access token expired, refresh
  // once and retry transparently. Never do this for the auth endpoints
  // themselves (they issue tokens; a 401 there is a real credential error), and
  // only when a refresh token exists — so anonymous/guest requests are wholly
  // unaffected.
  if (res.status === 401 && !path.startsWith('/auth/') && getRefreshToken()) {
    const newToken = await refreshAccessToken()
    if (newToken) {
      res = await fetch(`${BASE_URL}${path}`, {
        ...init,
        headers: buildHeaders(init, newToken),
      })
    }
  }

  // A 204 (every DELETE in this API) carries no body, and res.json() throws on
  // an empty payload — which used to make a successful delete look like a
  // failure to the caller and get optimistically reverted in the UI. Read the
  // body as text first and only parse when there is something to parse.
  const text = await res.text()
  let body: { success?: boolean; message?: string; data?: unknown } | null = null
  if (text) {
    try {
      body = JSON.parse(text)
    } catch {
      // Non-JSON payload (e.g. a proxy error page): fall through to the status
      // check below so the caller still gets a meaningful error.
      body = null
    }
  }

  if (!res.ok || body?.success === false) {
    throw new Error(body?.message ?? `Request failed: ${res.status}`)
  }
  return (body ? body.data : undefined) as T
}
