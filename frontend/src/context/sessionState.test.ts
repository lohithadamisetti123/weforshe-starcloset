import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import {
  EMPTY_SESSION,
  MAX_EXCLUDE_IDS,
  MAX_SHOWN_PER_DEPARTMENT,
  MAX_TURNS_PER_DEPARTMENT,
  activeShown,
  activeTurns,
  capExclusions,
  dislikeProduct,
  genderForDepartment,
  mergeShown,
  noteShown,
  parseSession,
  productIdsOf,
  recordTurn,
  resetConversation,
  selectDepartment,
  setGender,
  storageKey,
  undislikeProduct,
  type ConversationTurn,
  type SessionState,
} from './sessionState.ts'
import type { CollectionResponse } from '@/types'

// ─── helpers ─────────────────────────────────────────────────────────────────

function turn(overrides: Partial<ConversationTurn> = {}): ConversationTurn {
  return {
    id: overrides.id ?? Math.random().toString(36).slice(2),
    department: overrides.department ?? 'Women',
    prompt: overrides.prompt ?? 'office wear',
    timestamp: overrides.timestamp ?? '2026-07-22T00:00:00.000Z',
    productIds: overrides.productIds ?? [],
    collectionTitles: overrides.collectionTitles ?? [],
  }
}

function collection(ids: string[], title = 'Office Collection'): CollectionResponse {
  return {
    title,
    items: ids.map((id) => ({
      slot: 'Top',
      // Only `product.id` is read by the code under test; the rest of the
      // Product shape is irrelevant here and asserting on it would couple this
      // suite to an unrelated type.
      product: { id } as CollectionResponse['items'][number]['product'],
    })),
    totalPrice: 0,
    explanation: [],
    department: 'Women',
  }
}

function ids(n: number, prefix = 'p'): string[] {
  return Array.from({ length: n }, (_, i) => `${prefix}${i}`)
}

// ─── department & gender locking ─────────────────────────────────────────────
//
// These are the rules the brief cares most about: a shopper in Women must never
// be handed menswear, and gender must not drift on its own.

describe('gender locking', () => {
  test('a gendered department sets the gender', () => {
    assert.equal(selectDepartment(EMPTY_SESSION, 'Women').gender, 'Women')
    assert.equal(selectDepartment(EMPTY_SESSION, 'Men').gender, 'Men')
    assert.equal(selectDepartment(EMPTY_SESSION, 'Kids - Girls').gender, 'Girls')
    assert.equal(selectDepartment(EMPTY_SESSION, 'Kids - Boys').gender, 'Boys')
  })

  test('a gender-agnostic department inherits the gender already in force', () => {
    // A woman browsing Footwear must still get women's shoes. This mirrors
    // BuildCollection's `if gender == nil && base.Gender != nil` fallback.
    const women = selectDepartment(EMPTY_SESSION, 'Women')
    for (const dept of ['Footwear', 'Accessories', 'Beauty']) {
      assert.equal(selectDepartment(women, dept).gender, 'Women', dept)
    }
  })

  test('a gender-agnostic department leaves gender null when none was set', () => {
    assert.equal(selectDepartment(EMPTY_SESSION, 'Footwear').gender, null)
  })

  test('switching from Women to Men flips the gender and never keeps both', () => {
    const women = selectDepartment(EMPTY_SESSION, 'Women')
    const men = selectDepartment(women, 'Men')
    assert.equal(men.gender, 'Men')
    assert.equal(men.department, 'Men')
  })

  test('gender never changes without a department change or explicit set', () => {
    const women = selectDepartment(EMPTY_SESSION, 'Women')
    const afterTurn = recordTurn(women, turn({ productIds: ['a'] }))
    const afterShown = noteShown(afterTurn, ['b'])
    const afterDislike = dislikeProduct(afterShown, 'c')
    const afterReset = resetConversation(afterDislike)
    assert.equal(afterReset.gender, 'Women')
  })

  test('setGender is the explicit override', () => {
    const women = selectDepartment(EMPTY_SESSION, 'Women')
    assert.equal(setGender(women, 'Men').gender, 'Men')
    assert.equal(setGender(women, null).gender, null)
  })

  test('genderForDepartment is total over unknown departments', () => {
    // An unrecognised department must inherit rather than throw or blank out.
    assert.equal(genderForDepartment('Home Essentials', 'Women'), 'Women')
    assert.equal(genderForDepartment('', null), null)
  })

  test('re-selecting the same department is a no-op by reference', () => {
    const women = selectDepartment(EMPTY_SESSION, 'Women')
    assert.equal(selectDepartment(women, 'Women'), women)
  })
})

// ─── exclusion list caps ─────────────────────────────────────────────────────
//
// The backend rejects >200 ids with a 400. Overflowing the cap would take the
// assistant from "slightly repetitive" to "completely broken", so it is pinned.

describe('exclusion caps', () => {
  test('capExclusions never exceeds the backend validation limit', () => {
    const out = capExclusions(ids(500))
    assert.equal(out.length, MAX_EXCLUDE_IDS)
  })

  test('capExclusions de-duplicates across groups', () => {
    const out = capExclusions(['a', 'b'], ['b', 'c'], ['a'])
    assert.deepEqual(out, ['a', 'b', 'c'])
  })

  test('capExclusions preserves group order so earlier groups survive the trim', () => {
    // "More Collections" passes on-screen ids first precisely so they are the
    // ones that survive — those are the repeats a shopper notices instantly.
    const visible = ids(5, 'visible')
    const older = ids(400, 'older')
    const out = capExclusions(visible, older)
    assert.deepEqual(out.slice(0, 5), visible)
    assert.equal(out.length, MAX_EXCLUDE_IDS)
  })

  test('capExclusions drops empty ids', () => {
    assert.deepEqual(capExclusions(['a', '', 'b']), ['a', 'b'])
  })

  test('capExclusions of nothing is empty', () => {
    assert.deepEqual(capExclusions(), [])
    assert.deepEqual(capExclusions([], []), [])
  })
})

describe('mergeShown', () => {
  test('newest ids come first', () => {
    assert.deepEqual(mergeShown(['old'], ['new']), ['new', 'old'])
  })

  test('de-duplicates, keeping the newest position', () => {
    assert.deepEqual(mergeShown(['a', 'b'], ['b']), ['b', 'a'])
  })

  test('caps at the per-department limit', () => {
    const out = mergeShown(ids(200, 'old'), ids(200, 'new'))
    assert.equal(out.length, MAX_SHOWN_PER_DEPARTMENT)
    assert.equal(out[0], 'new0')
  })

  test('empty incoming returns the existing list unchanged', () => {
    const existing = ['a']
    assert.equal(mergeShown(existing, []), existing)
  })
})

// ─── conversation memory ─────────────────────────────────────────────────────

describe('conversation memory', () => {
  test('recordTurn accumulates shown ids so later turns can exclude them', () => {
    let s = selectDepartment(EMPTY_SESSION, 'Women')
    s = recordTurn(s, turn({ productIds: ['p1', 'p2'] }))
    s = recordTurn(s, turn({ productIds: ['p3'] }))
    assert.deepEqual(activeShown(s), ['p3', 'p1', 'p2'])
    assert.equal(activeTurns(s).length, 2)
  })

  test('shown ids and turns are partitioned by department', () => {
    let s = selectDepartment(EMPTY_SESSION, 'Women')
    s = recordTurn(s, turn({ department: 'Women', productIds: ['w1'] }))
    s = selectDepartment(s, 'Men')
    s = recordTurn(s, turn({ department: 'Men', productIds: ['m1'] }))

    assert.deepEqual(activeShown(s), ['m1'], 'Men sees only menswear ids')
    assert.deepEqual(activeTurns(s).map((t) => t.department), ['Men'])

    const back = selectDepartment(s, 'Women')
    assert.deepEqual(activeShown(back), ['w1'], 'Women history survived the round trip')
  })

  test('turns are capped per department without touching other departments', () => {
    let s = selectDepartment(EMPTY_SESSION, 'Women')
    s = recordTurn(s, turn({ id: 'men-1', department: 'Men' }))
    for (let i = 0; i < MAX_TURNS_PER_DEPARTMENT + 10; i++) {
      s = recordTurn(s, turn({ id: `w${i}`, department: 'Women' }))
    }
    assert.equal(activeTurns(s).length, MAX_TURNS_PER_DEPARTMENT)
    // Oldest Women turns were dropped, newest kept.
    assert.equal(activeTurns(s).at(-1)?.id, `w${MAX_TURNS_PER_DEPARTMENT + 9}`)
    // The Men turn is untouched.
    assert.ok(s.turns.some((t) => t.id === 'men-1'))
  })

  test('recordTurn without a department is a no-op', () => {
    const t = turn()
    assert.equal(recordTurn(EMPTY_SESSION, t), EMPTY_SESSION)
  })

  test('noteShown without a department is a no-op', () => {
    assert.equal(noteShown(EMPTY_SESSION, ['a']), EMPTY_SESSION)
  })

  test('noteShown ignores empty input', () => {
    const s = selectDepartment(EMPTY_SESSION, 'Women')
    assert.equal(noteShown(s, []), s)
    assert.equal(noteShown(s, ['', '']), s)
  })
})

describe('resetConversation', () => {
  test('clears this department but keeps department and gender', () => {
    let s = selectDepartment(EMPTY_SESSION, 'Women')
    s = recordTurn(s, turn({ productIds: ['p1'] }))
    const reset = resetConversation(s)

    assert.equal(reset.department, 'Women')
    assert.equal(reset.gender, 'Women')
    assert.deepEqual(activeShown(reset), [])
    assert.deepEqual(activeTurns(reset), [])
  })

  test('leaves other departments intact', () => {
    let s = selectDepartment(EMPTY_SESSION, 'Men')
    s = recordTurn(s, turn({ department: 'Men', productIds: ['m1'] }))
    s = selectDepartment(s, 'Women')
    s = recordTurn(s, turn({ department: 'Women', productIds: ['w1'] }))

    const reset = resetConversation(s)
    assert.deepEqual(activeShown(reset), [])
    assert.deepEqual(reset.shown['Men'], ['m1'])
  })
})

describe('dislikes', () => {
  test('records and clears, and is idempotent', () => {
    let s = dislikeProduct(EMPTY_SESSION, 'p1')
    assert.deepEqual(s.dislikedProductIds, ['p1'])
    assert.equal(dislikeProduct(s, 'p1'), s, 'second dislike is a no-op')
    s = undislikeProduct(s, 'p1')
    assert.deepEqual(s.dislikedProductIds, [])
    assert.equal(undislikeProduct(s, 'nope'), s)
  })

  test('ignores an empty id', () => {
    assert.equal(dislikeProduct(EMPTY_SESSION, ''), EMPTY_SESSION)
  })

  test('dislikes are global, surviving a department change', () => {
    let s = selectDepartment(EMPTY_SESSION, 'Women')
    s = dislikeProduct(s, 'p1')
    s = selectDepartment(s, 'Men')
    assert.deepEqual(s.dislikedProductIds, ['p1'])
  })
})

// ─── persistence ─────────────────────────────────────────────────────────────

describe('parseSession', () => {
  test('round-trips a real session', () => {
    let s = selectDepartment(EMPTY_SESSION, 'Women')
    s = recordTurn(s, turn({ productIds: ['p1'] }))
    s = dislikeProduct(s, 'p9')
    assert.deepEqual(parseSession(JSON.stringify(s)), s)
  })

  test('falls back to a clean session on anything unrecognised', () => {
    // A half-restored session is worse than a fresh one: the shopper would be
    // in a department with no memory of how they got there.
    const bad: unknown[] = [
      null,
      '',
      'not json',
      '{',
      JSON.stringify({ version: 0 }),
      JSON.stringify({ version: 2, department: 'Women' }),
      JSON.stringify({}),
    ]
    for (const raw of bad) {
      assert.deepEqual(parseSession(raw as string | null), EMPTY_SESSION, String(raw))
    }
  })

  test('coerces malformed fields rather than trusting them', () => {
    const parsed = parseSession(
      JSON.stringify({
        version: 1,
        department: 42,
        turns: 'nope',
        shown: ['not', 'an', 'object'],
        dislikedProductIds: null,
      }),
    )
    assert.equal(parsed.department, null)
    assert.deepEqual(parsed.turns, [])
    assert.deepEqual(parsed.shown, {})
    assert.deepEqual(parsed.dislikedProductIds, [])
  })

  test('storage keys are namespaced per identity', () => {
    // A browser-global key previously let one account read another's cart.
    // Session state is identity-scoped for the same reason.
    assert.notEqual(storageKey('user-a'), storageKey('user-b'))
    assert.equal(storageKey(null), 'myntra-session:guest')
  })
})

// ─── collection helpers ──────────────────────────────────────────────────────

describe('productIdsOf', () => {
  test('flattens and de-duplicates across collections', () => {
    const out = productIdsOf([collection(['a', 'b']), collection(['b', 'c'])])
    assert.deepEqual(out, ['a', 'b', 'c'])
  })

  test('tolerates empty collections', () => {
    assert.deepEqual(productIdsOf([]), [])
    assert.deepEqual(productIdsOf([collection([])]), [])
  })
})

// ─── end-to-end conversation ─────────────────────────────────────────────────

test('a multi-turn conversation never re-offers what it already showed', () => {
  // This is the reported bug, expressed as a scenario: three turns in one
  // department, each excluding everything shown so far.
  let s = selectDepartment(EMPTY_SESSION, 'Women')
  const seen = new Set<string>()

  for (const batch of [['a', 'b', 'c'], ['d', 'e'], ['f']]) {
    const exclusions = capExclusions(activeShown(s))
    for (const id of batch) {
      assert.ok(!exclusions.includes(id), `${id} was not previously excluded`)
      assert.ok(!seen.has(id), `${id} is genuinely new`)
      seen.add(id)
    }
    s = recordTurn(s, turn({ productIds: batch }))
  }

  assert.deepEqual(activeShown(s), ['f', 'd', 'e', 'a', 'b', 'c'])
  assert.equal(capExclusions(activeShown(s)).length, 6)
})
