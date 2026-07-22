import type { AuthResponse, AuthTokens, AuthUser } from '@/types'

// Single source of truth for the stored auth session.
//
// Tokens and the user record live in localStorage so the session survives a
// page refresh and a closed tab — the same persistence the app already uses for
// its guest session id. This module is the *only* place that reads or writes
// those keys, so there is no duplicate auth-storage logic anywhere else: the API
// client reads the access token here, and the auth context reads/writes the
// whole session here.
//
// It deliberately imports nothing but types, so both the API client and the auth
// context can depend on it without creating an import cycle.

const ACCESS_KEY = 'myntra-access-token'
const REFRESH_KEY = 'myntra-refresh-token'
const USER_KEY = 'myntra-auth-user'

export function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY)
}

export function getStoredUser(): AuthUser | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as AuthUser
  } catch {
    return null
  }
}

// setTokens updates only the token pair (used by the silent refresh path, which
// gets new tokens but the same user).
export function setTokens(tokens: AuthTokens): void {
  localStorage.setItem(ACCESS_KEY, tokens.accessToken)
  localStorage.setItem(REFRESH_KEY, tokens.refreshToken)
}

// setSession persists a full login result (user + tokens).
export function setSession(res: AuthResponse): void {
  setTokens(res.tokens)
  localStorage.setItem(USER_KEY, JSON.stringify(res.user))
}

// Legacy browser-global keys from before cart/wishlist moved to per-user
// backend storage. They are shared by every account on the machine, so they are
// purged whenever the signed-in identity changes — otherwise a stale blob from
// a previous user could still be read by an older build or a future regression.
const LEGACY_SHARED_KEYS = [
  'myntra-cart',
  'myntra-saved-collections',
  'myntra-collection-cart',
  'myntra-chat-history',
]

// clearUserScopedCaches drops any client-side data belonging to a specific
// shopper. Called on both login and logout so nothing survives an identity
// change in memory or in storage.
export function clearUserScopedCaches(): void {
  for (const key of LEGACY_SHARED_KEYS) {
    localStorage.removeItem(key)
  }
  // Per-user namespaced keys: chat history (myntra-chat-history:<userId>)
  // and the shopping session (myntra-session:<userId>). Both are identity
  // scoped, so an identity change must drop every namespace, not just the
  // outgoing user's — a stale blob under another id would otherwise be read
  // back if that user signed in again on this machine.
  const NAMESPACED_PREFIXES = ['myntra-chat-history:', 'myntra-session:']
  for (let i = localStorage.length - 1; i >= 0; i--) {
    const key = localStorage.key(i)
    if (key && NAMESPACED_PREFIXES.some((prefix) => key.startsWith(prefix))) {
      localStorage.removeItem(key)
    }
  }
}

// clearSession removes everything auth-related plus any user-scoped client
// caches. The guest session id is left intact, so logging out cleanly drops the
// app back to its guest state.
export function clearSession(): void {
  localStorage.removeItem(ACCESS_KEY)
  localStorage.removeItem(REFRESH_KEY)
  localStorage.removeItem(USER_KEY)
  clearUserScopedCaches()
}
