import {
  createContext,
  useContext,
  useCallback,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '@/lib/api'
import { useShop } from '@/context/ShopContext'
import { useAuth } from '@/context/AuthContext'
import type { StylistPersona } from '@/types/persona'

// The unlock score is the shopper's engagement currency: every product in the
// cart plus every collection added to the collection-cart. Stylists unlock as it
// grows — the loop that turns "try a stylist" into "keep shopping". It is
// deliberately derived from cart state (not persisted separately) so it can
// never disagree with what's actually in the cart.
function storageKey(userId: string | undefined): string {
  return `starcloset-persona:${userId ?? 'guest'}`
}

interface PersonaContextValue {
  personas: StylistPersona[]
  loading: boolean
  /** The shopper's current engagement score = cartCount + collectionCartCount. */
  unlockScore: number
  /** The slug the shopper has chosen, or null. */
  selectedSlug: string | null
  /** The resolved selected persona, or null. */
  selectedPersona: StylistPersona | null
  selectPersona: (slug: string) => void
  isUnlocked: (persona: StylistPersona) => boolean
  /** Items still needed to unlock a persona (0 once unlocked). */
  itemsToUnlock: (persona: StylistPersona) => number
  unlockedPersonas: StylistPersona[]
  /** The next locked persona by threshold, or null when all are unlocked. */
  nextLocked: StylistPersona | null
}

const PersonaContext = createContext<PersonaContextValue | null>(null)

export function PersonaProvider({ children }: { children: ReactNode }) {
  const { cartCount, collectionCartCount } = useShop()
  const { user } = useAuth()
  const unlockScore = cartCount + collectionCartCount

  const { data, isLoading } = useQuery<StylistPersona[]>({
    queryKey: ['stylist-personas'],
    queryFn: () => apiFetch<StylistPersona[]>('/assistant/personas'),
    staleTime: 5 * 60 * 1000,
  })
  const personas = useMemo(() => data ?? [], [data])

  const [selectedSlug, setSelectedSlug] = useState<string | null>(null)

  // Load the shopper's last-chosen stylist when identity resolves.
  useEffect(() => {
    const stored = localStorage.getItem(storageKey(user?.id))
    setSelectedSlug(stored)
  }, [user?.id])

  const selectPersona = useCallback(
    (slug: string) => {
      setSelectedSlug(slug)
      localStorage.setItem(storageKey(user?.id), slug)
    },
    [user?.id],
  )

  const isUnlocked = useCallback(
    (persona: StylistPersona) => unlockScore >= persona.unlockAtItems,
    [unlockScore],
  )

  const itemsToUnlock = useCallback(
    (persona: StylistPersona) => Math.max(0, persona.unlockAtItems - unlockScore),
    [unlockScore],
  )

  const selectedPersona = useMemo(
    () => personas.find((p) => p.slug === selectedSlug) ?? null,
    [personas, selectedSlug],
  )

  const unlockedPersonas = useMemo(
    () => personas.filter((p) => unlockScore >= p.unlockAtItems),
    [personas, unlockScore],
  )

  const nextLocked = useMemo(() => {
    const locked = personas
      .filter((p) => unlockScore < p.unlockAtItems)
      .sort((a, b) => a.unlockAtItems - b.unlockAtItems)
    return locked[0] ?? null
  }, [personas, unlockScore])

  const value: PersonaContextValue = {
    personas,
    loading: isLoading,
    unlockScore,
    selectedSlug,
    selectedPersona,
    selectPersona,
    isUnlocked,
    itemsToUnlock,
    unlockedPersonas,
    nextLocked,
  }

  return <PersonaContext.Provider value={value}>{children}</PersonaContext.Provider>
}

export function usePersona(): PersonaContextValue {
  const ctx = useContext(PersonaContext)
  if (!ctx) throw new Error('usePersona must be used within a PersonaProvider')
  return ctx
}
