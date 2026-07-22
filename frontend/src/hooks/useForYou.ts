import { useEffect, useState } from 'react'
import { apiFetch } from '@/lib/api'
import type {
  AIRecommendationResponse,
  CollectionResponse,
  CollectionsResponse,
  Product,
  UserProfile,
} from '@/types'

// useForYou builds a personalized Recommendations feed from the shopper's saved
// profile — no questions asked. It derives a department and a natural-language
// brief from the profile, then calls the assistant's existing endpoints, which
// already apply the full personalization boost layer server-side (favourite
// brands, colours, purchase history, location, and the occasion/style/budget the
// brief encodes). So the products and collections it returns are ranked for
// this shopper specifically, and they update the moment the profile changes,
// because the brief is rebuilt from the latest profile every time.

// Map the profile's gender onto a supported collection department.
function departmentFor(gender: string): string {
  switch (gender) {
    case 'Men':
      return 'Men'
    case 'Girls':
      return 'Kids - Girls'
    case 'Boys':
      return 'Kids - Boys'
    case 'Women':
    default:
      return 'Women'
  }
}

// Synthesize a brief from the profile. Kept human-readable so the same extractor
// the assistant uses everywhere parses it into a proper intent; personalization
// boosts are then layered on top by the backend.
function briefFor(p: UserProfile): string {
  const parts: string[] = []
  if (p.favoriteStyles[0]) parts.push(p.favoriteStyles[0])
  parts.push(p.favoriteCategories[0] ?? 'outfits')
  if (p.favoriteOccasions[0]) parts.push(`for ${p.favoriteOccasions[0]}`)
  let brief = parts.join(' ').trim()
  if (p.budgetMax) brief += ` under ₹${p.budgetMax}`
  return brief || 'trending picks for me'
}

interface ForYouState {
  loading: boolean
  ready: boolean
  hasProfile: boolean
  profileName: string
  products: Product[]
  collections: CollectionResponse[]
  personalized: boolean
}

export function useForYou() {
  const [state, setState] = useState<ForYouState>({
    loading: true,
    ready: false,
    hasProfile: false,
    profileName: '',
    products: [],
    collections: [],
    personalized: false,
  })

  useEffect(() => {
    let alive = true

    async function run() {
      let profile: UserProfile | null = null
      try {
        profile = await apiFetch<UserProfile>('/profile/preferences')
      } catch {
        profile = null
      }

      // No completed profile → nothing to personalize on. The page falls back
      // to its standard dashboard/catalogue view.
      if (!profile || !profile.onboarded) {
        if (alive) {
          setState((s) => ({ ...s, loading: false, ready: true, hasProfile: false }))
        }
        return
      }

      const department = departmentFor(profile.gender)
      const brief = briefFor(profile)

      // Fetch a personalized product shortlist and personalized collections in
      // parallel. Each call is independently best-effort: one failing still
      // lets the other render.
      const productsP = apiFetch<AIRecommendationResponse>('/assistant/chat', {
        method: 'POST',
        body: JSON.stringify({ prompt: `${department}: ${brief}` }),
      })
        .then((r) => ({ products: r.products ?? [], personalized: r.meta?.personalized ?? false }))
        .catch(() => ({ products: [] as Product[], personalized: false }))

      const collectionsP = apiFetch<CollectionsResponse>('/assistant/collections', {
        method: 'POST',
        body: JSON.stringify({ department, prompt: brief }),
      })
        .then((r) => ({ collections: r.collections ?? [], personalized: r.meta?.personalized ?? false }))
        .catch(() => ({ collections: [] as CollectionResponse[], personalized: false }))

      const [prod, coll] = await Promise.all([productsP, collectionsP])
      if (!alive) return

      setState({
        loading: false,
        ready: true,
        hasProfile: true,
        profileName: profile.name,
        products: prod.products,
        collections: coll.collections,
        personalized: prod.personalized || coll.personalized,
      })
    }

    run()
    return () => {
      alive = false
    }
  }, [])

  return state
}
