import type { CollectionResponse, Gender } from '@/types'

// Pure session logic, deliberately free of React and of the DOM.
//
// SessionContext.tsx is the React binding: it owns the provider, the storage
// side effects and the memoised value. Everything that decides *what the
// session becomes* lives here as pure functions over plain data, which is what
// makes the rules in this file directly testable (see sessionState.test.ts)
// without a renderer, a DOM, or a mocked localStorage.
//
// The split matters because these rules are the ones with real consequences —
// gender leaking across departments, or the exclusion list overflowing the
// backend's 200-item validation cap — and those are exactly the things that
// should be pinned by tests rather than by inspection.

/** How many shown-product ids to retain per department. */
//
// The backend validates `excludeProductIds` with `max=200`
// (dto/collection.go). Exceeding it is a 400, not a degraded result, so this
// cap is a correctness constraint rather than a memory optimisation. 150
// leaves headroom for the per-request additions made by "More Collections",
// which prepends the ids visible in the current message.
export const MAX_SHOWN_PER_DEPARTMENT = 150

/** How many conversation turns to retain per department. */
export const MAX_TURNS_PER_DEPARTMENT = 40

/** Hard ceiling on any outbound `excludeProductIds` array. Mirrors the
 *  backend's `validate:"omitempty,max=200"`. */
export const MAX_EXCLUDE_IDS = 200

export const SESSION_STORAGE_PREFIX = 'myntra-session:'

/** Departments that imply a shopper gender, mirroring the backend's
 *  `DepartmentToGender` in internal/recommendation/collection.go.
 *
 *  Beauty, Footwear and Accessories are absent on purpose: they are
 *  gender-agnostic there too, and the backend falls back to the intent's own
 *  gender for them. */
export const DEPARTMENT_GENDER: Readonly<Record<string, Gender>> = Object.freeze({
  Women: 'Women',
  Men: 'Men',
  'Kids - Girls': 'Girls',
  'Kids - Boys': 'Boys',
})

/** One recorded exchange with the assistant. */
export interface ConversationTurn {
  id: string
  department: string
  prompt: string
  timestamp: string
  /** Ids of every product surfaced by this turn. */
  productIds: string[]
  /** Titles of the collections returned, for lightweight recall. */
  collectionTitles: string[]
}

/** The full session state, and the exact shape persisted to storage.
 *  Versioned so a future change can migrate rather than crash on a stale blob. */
export interface SessionState {
  version: 1
  department: string | null
  gender: Gender | null
  turns: ConversationTurn[]
  /** department -> product ids already shown in that department. */
  shown: Record<string, string[]>
  dislikedProductIds: string[]
}

export const EMPTY_SESSION: SessionState = Object.freeze({
  version: 1,
  department: null,
  gender: null,
  turns: [],
  shown: {},
  dislikedProductIds: [],
}) as SessionState

export function storageKey(userId: string | null): string {
  return `${SESSION_STORAGE_PREFIX}${userId ?? 'guest'}`
}

/** Parses a persisted blob, falling back to a clean session on anything
 *  unrecognised. Half-restoring a session is worse than starting fresh. */
export function parseSession(raw: string | null): SessionState {
  if (!raw) return EMPTY_SESSION
  try {
    const parsed = JSON.parse(raw) as Partial<SessionState>
    if (parsed.version !== 1) return EMPTY_SESSION
    return {
      version: 1,
      department: typeof parsed.department === 'string' ? parsed.department : null,
      gender: (parsed.gender ?? null) as Gender | null,
      turns: Array.isArray(parsed.turns) ? parsed.turns : [],
      shown:
        parsed.shown && typeof parsed.shown === 'object' && !Array.isArray(parsed.shown)
          ? parsed.shown
          : {},
      dislikedProductIds: Array.isArray(parsed.dislikedProductIds)
        ? parsed.dislikedProductIds
        : [],
    }
  } catch {
    return EMPTY_SESSION
  }
}

/** Appends ids to a capped, de-duplicated, most-recent-first list. */
export function mergeShown(
  existing: readonly string[],
  incoming: readonly string[],
  cap: number = MAX_SHOWN_PER_DEPARTMENT,
): string[] {
  if (incoming.length === 0) return existing as string[]
  const merged = [...incoming, ...existing].filter(Boolean)
  return Array.from(new Set(merged)).slice(0, cap)
}

/** De-duplicates and caps an outbound exclusion list. Earlier groups win the
 *  cap, so callers pass the ids the shopper would most notice repeating first. */
export function capExclusions(...groups: readonly string[][]): string[] {
  return Array.from(new Set(groups.flat().filter(Boolean))).slice(0, MAX_EXCLUDE_IDS)
}

/** Resolves the gender in force after selecting a department.
 *
 *  A gendered department sets it — choosing "Men" *is* an explicit request. A
 *  gender-agnostic department inherits whatever was already in force, so a
 *  shopper who picked Women and then browsed Footwear still gets women's shoes.
 *  This mirrors BuildCollection, which falls back to `base.Gender` when the
 *  department carries no gender of its own. */
export function genderForDepartment(
  department: string,
  current: Gender | null,
): Gender | null {
  return DEPARTMENT_GENDER[department] ?? current
}

/** Collects the distinct product ids surfaced by a set of collections. */
export function productIdsOf(collections: readonly CollectionResponse[]): string[] {
  return Array.from(
    new Set(collections.flatMap((c) => c.items.map((it) => it.product.id)).filter(Boolean)),
  )
}

// ─── Transitions ─────────────────────────────────────────────────────────────
// Each returns a new state, or the same reference when nothing changed, so the
// React binding's memoisation and the storage effect both stay cheap.

export function selectDepartment(state: SessionState, department: string): SessionState {
  if (state.department === department) return state
  return {
    ...state,
    department,
    gender: genderForDepartment(department, state.gender),
  }
}

export function setGender(state: SessionState, gender: Gender | null): SessionState {
  if (state.gender === gender) return state
  return { ...state, gender }
}

export function noteShown(state: SessionState, productIds: readonly string[]): SessionState {
  const clean = productIds.filter(Boolean)
  if (clean.length === 0 || !state.department) return state
  const dept = state.department
  return {
    ...state,
    shown: { ...state.shown, [dept]: mergeShown(state.shown[dept] ?? [], clean) },
  }
}

export function recordTurn(
  state: SessionState,
  turn: ConversationTurn,
): SessionState {
  if (!state.department) return state
  const dept = state.department

  // Keep only the most recent turns for this department; other departments'
  // turns are untouched.
  const appended = [...state.turns, turn]
  const forDept = appended.filter((t) => t.department === dept)
  const excess = forDept.length - MAX_TURNS_PER_DEPARTMENT
  const dropped = excess > 0 ? new Set(forDept.slice(0, excess).map((t) => t.id)) : null
  const turns = dropped ? appended.filter((t) => !dropped.has(t.id)) : appended

  return {
    ...state,
    turns,
    shown: { ...state.shown, [dept]: mergeShown(state.shown[dept] ?? [], turn.productIds) },
  }
}

export function dislikeProduct(state: SessionState, productId: string): SessionState {
  if (!productId || state.dislikedProductIds.includes(productId)) return state
  return { ...state, dislikedProductIds: [...state.dislikedProductIds, productId] }
}

export function undislikeProduct(state: SessionState, productId: string): SessionState {
  if (!state.dislikedProductIds.includes(productId)) return state
  return {
    ...state,
    dislikedProductIds: state.dislikedProductIds.filter((id) => id !== productId),
  }
}

/** Clears the conversation for the active department, keeping department and
 *  gender. Backs "New Chat". */
export function resetConversation(state: SessionState): SessionState {
  if (!state.department) return state
  const dept = state.department
  const shown = { ...state.shown }
  delete shown[dept]
  return {
    ...state,
    turns: state.turns.filter((t) => t.department !== dept),
    shown,
  }
}

/** Turns for the active department, oldest first. */
export function activeTurns(state: SessionState): ConversationTurn[] {
  if (!state.department) return []
  return state.turns.filter((t) => t.department === state.department)
}

/** Shown product ids for the active department, most recent first. */
export function activeShown(state: SessionState): string[] {
  if (!state.department) return []
  return state.shown[state.department] ?? []
}
