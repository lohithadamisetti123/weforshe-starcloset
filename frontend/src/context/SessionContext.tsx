import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { useAuth } from '@/context/AuthContext'
import * as session from '@/context/sessionState'
import type { CollectionResponse, Gender } from '@/types'

// SessionContext is the single source of truth for the *shopping session*:
// which department the shopper is in, which gender that implies, what the
// assistant has already shown them, and what was said in earlier turns.
//
// Why this exists
// ---------------
// Before this, `department` lived in a bare `useState` inside `useChat`. That
// made it component-local: a page refresh, or navigating to /cart and back,
// dropped it and threw the shopper back to the department picker. Worse, since
// nothing accumulated the products already shown, every turn of a conversation
// was independent — turn 2 had no idea what turn 1 had recommended, so the same
// products resurfaced repeatedly.
//
// Both are the same root cause: there was no session, only per-component state.
// This module is that session.
//
// This file is deliberately thin. Every rule with real consequences —
// gender inheritance across departments, the per-department caps, the
// exclusion-list ceiling the backend validates against — lives in
// sessionState.ts as pure functions, and is covered by sessionState.test.ts.
// What remains here is the React and storage plumbing.
//
// Persistence model
// -----------------
// State is mirrored to localStorage under a key namespaced by the signed-in
// user id (`myntra-session:<userId>`, or `:guest`). Namespacing is not
// optional: this codebase has already been bitten once by a browser-global key
// letting one account see another account's cart (see the note at the top of
// useCart.ts). Session state is identity-scoped for the same reason, and
// `clearUserScopedCaches()` in lib/auth.ts purges these keys on login and
// logout.

function readSession(userId: string | null): session.SessionState {
  try {
    return session.parseSession(localStorage.getItem(session.storageKey(userId)))
  } catch {
    // Storage disabled entirely (private mode, blocked cookies). Not a reason
    // to stop the shopper from shopping.
    return session.EMPTY_SESSION
  }
}

function writeSession(userId: string | null, state: session.SessionState): void {
  try {
    localStorage.setItem(session.storageKey(userId), JSON.stringify(state))
  } catch {
    // Quota exceeded or storage unavailable. The in-memory session remains
    // authoritative for this tab; only cross-refresh durability is lost.
  }
}

export type { ConversationTurn } from '@/context/sessionState'
export { SESSION_STORAGE_PREFIX } from '@/context/sessionState'

export interface SessionContextValue {
  /** The department currently being shopped, or null before selection. */
  department: string | null
  /** The gender constraint in force. Null for a gender-agnostic session. */
  gender: Gender | null

  /** Select a department. Derives gender for gendered departments; leaves it
   *  untouched for gender-agnostic ones. */
  selectDepartment: (department: string) => void
  /** Explicitly override the gender. The only other way it changes is by
   *  choosing a gendered department, which is itself an explicit act. */
  setGender: (gender: Gender | null) => void

  /** Conversation turns for the active department, oldest first. */
  turns: session.ConversationTurn[]
  /** Record a completed exchange. */
  recordTurn: (prompt: string, collections: CollectionResponse[]) => void

  /** Product ids already shown in the active department, most recent first.
   *  Pass through `capExclusions` before sending as `excludeProductIds`. */
  shownProductIds: string[]
  /** Note products as shown without recording a full turn. */
  noteShown: (productIds: string[]) => void

  /** Products the shopper explicitly rejected. */
  dislikedProductIds: string[]
  dislikeProduct: (productId: string) => void
  undislikeProduct: (productId: string) => void
  isDisliked: (productId: string) => boolean

  /** Clear the conversation for the active department, keeping department and
   *  gender. Backs "New Chat". */
  resetConversation: () => void
  /** Clear everything including department and gender. Backs "Change". */
  clearSession: () => void
}

const SessionContext = createContext<SessionContextValue | null>(null)

export function SessionProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth()
  const userId = user?.id ?? null

  // Initialise synchronously from storage so a refresh comes back already in
  // the right department, with no flash of the department picker.
  const [state, setState] = useState<session.SessionState>(() => readSession(userId))

  // Track which identity the in-memory session belongs to, so a change of user
  // swaps the session rather than writing one user's state under another's key.
  const loadedFor = useRef<string | null>(userId)

  useEffect(() => {
    if (loadedFor.current === userId) return
    loadedFor.current = userId
    setState(readSession(userId))
  }, [userId])

  // Mirror to storage on every change, guarded by the same identity check so an
  // in-flight write from the previous user cannot land under the new key.
  useEffect(() => {
    if (loadedFor.current !== userId) return
    writeSession(userId, state)
  }, [state, userId])

  const selectDepartment = useCallback((department: string) => {
    setState((prev) => session.selectDepartment(prev, department))
  }, [])

  const setGender = useCallback((gender: Gender | null) => {
    setState((prev) => session.setGender(prev, gender))
  }, [])

  const noteShown = useCallback((productIds: string[]) => {
    setState((prev) => session.noteShown(prev, productIds))
  }, [])

  const recordTurn = useCallback((prompt: string, collections: CollectionResponse[]) => {
    setState((prev) => {
      if (!prev.department) return prev
      return session.recordTurn(prev, {
        id:
          typeof crypto !== 'undefined' && 'randomUUID' in crypto
            ? crypto.randomUUID()
            : Math.random().toString(36).slice(2, 10),
        department: prev.department,
        prompt,
        timestamp: new Date().toISOString(),
        productIds: session.productIdsOf(collections),
        collectionTitles: collections.map((c) => c.title),
      })
    })
  }, [])

  const dislikeProduct = useCallback((productId: string) => {
    setState((prev) => session.dislikeProduct(prev, productId))
  }, [])

  const undislikeProduct = useCallback((productId: string) => {
    setState((prev) => session.undislikeProduct(prev, productId))
  }, [])

  const resetConversation = useCallback(() => {
    setState((prev) => session.resetConversation(prev))
  }, [])

  const clearSession = useCallback(() => {
    setState(session.EMPTY_SESSION)
  }, [])

  const turns = useMemo(() => session.activeTurns(state), [state])
  const shownProductIds = useMemo(() => session.activeShown(state), [state])

  const isDisliked = useCallback(
    (productId: string) => state.dislikedProductIds.includes(productId),
    [state.dislikedProductIds],
  )

  const value = useMemo<SessionContextValue>(
    () => ({
      department: state.department,
      gender: state.gender,
      selectDepartment,
      setGender,
      turns,
      recordTurn,
      shownProductIds,
      noteShown,
      dislikedProductIds: state.dislikedProductIds,
      dislikeProduct,
      undislikeProduct,
      isDisliked,
      resetConversation,
      clearSession,
    }),
    [
      state.department,
      state.gender,
      state.dislikedProductIds,
      turns,
      shownProductIds,
      selectDepartment,
      setGender,
      recordTurn,
      noteShown,
      dislikeProduct,
      undislikeProduct,
      isDisliked,
      resetConversation,
      clearSession,
    ],
  )

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>
}

/** Access the shopping session. Throws outside a SessionProvider, matching the
 *  convention used by useAuth and useShop. */
export function useSession(): SessionContextValue {
  const ctx = useContext(SessionContext)
  if (!ctx) {
    throw new Error('useSession must be used within a SessionProvider')
  }
  return ctx
}
