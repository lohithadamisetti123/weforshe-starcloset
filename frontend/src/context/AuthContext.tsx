import { createContext, useContext, useState, type ReactNode } from 'react'
import { apiFetch } from '@/lib/api'
import { clearSession, clearUserScopedCaches, getStoredUser, setSession } from '@/lib/auth'
import type { AuthResponse, AuthUser } from '@/types'

// AuthContext follows the same shape as this project's ShopProvider: a context
// created here, a provider that owns the state, and a `useAuth()` accessor that
// throws outside the provider. It is the single place that holds authentication
// state; every component reads it through the hook rather than touching storage
// or the auth endpoints directly.

interface AuthContextValue {
  user: AuthUser | null
  isAuthenticated: boolean
  signUp: (name: string, email: string, password: string) => Promise<AuthResponse>
  signIn: (email: string, password: string) => Promise<AuthResponse>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  // Restore the session synchronously from storage on first render, so a page
  // refresh comes back already signed in with no flash of the signed-out UI.
  const [user, setUser] = useState<AuthUser | null>(() => getStoredUser())

  const signUp = async (name: string, email: string, password: string): Promise<AuthResponse> => {
    const res = await apiFetch<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    })
    // Drop any client-side data left by a previous identity on this browser
    // before the new session takes over.
    clearUserScopedCaches()
    setSession(res)
    setUser(res.user)
    return res
  }

  const signIn = async (email: string, password: string): Promise<AuthResponse> => {
    const res = await apiFetch<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    // Signing in as a different user must not inherit the previous user's
    // client-side data.
    clearUserScopedCaches()
    setSession(res)
    setUser(res.user)
    return res
  }

  // logout clears the stored session and reloads to Home. A full reload is
  // deliberate: identity-scoped state loaded across the app (favourites,
  // history, profile) is re-initialised cleanly under the guest identity,
  // rather than trying to invalidate every hook by hand.
  const logout = () => {
    clearSession()
    setUser(null)
    window.location.assign('/')
  }

  const value: AuthContextValue = {
    user,
    isAuthenticated: user !== null,
    signUp,
    signIn,
    logout,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>')
  return ctx
}
