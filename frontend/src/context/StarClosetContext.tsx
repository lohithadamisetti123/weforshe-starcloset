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
import { useAuth } from '@/context/AuthContext'
import type { StarDepartment, StarStylist, ProgressResponse } from '@/types/starcloset'

// StarClosetContext owns the V2 department journey: three independent shopping
// departments (Women / Men / Kids), each with its own stylist roster and its own
// unlock progress. Progress is authoritative on the backend (star_progress),
// fetched here and advanced through recordCartAdd — one cart-add advances only
// the currently-selected department, which is what keeps the journeys
// independent. Selection (department + per-department stylist) is persisted per
// user in localStorage so a refresh or a trip to the cart returns you where you
// left off.

function deptKey(userId: string | undefined) {
  return `starcloset-dept:${userId ?? 'guest'}`
}
function picksKey(userId: string | undefined) {
  return `starcloset-picks:${userId ?? 'guest'}`
}

function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key)
    return raw ? (JSON.parse(raw) as T) : fallback
  } catch {
    return fallback
  }
}

interface StarClosetContextValue {
  departments: StarDepartment[]
  loading: boolean
  /** Per-department unlock counters, e.g. { Women: 2, Men: 0, Kids: 0 }. */
  progress: Record<string, number>

  selectedDepartment: string | null
  currentDept: StarDepartment | null
  selectDepartment: (key: string) => void
  clearDepartment: () => void

  selectedSlug: string | null
  selectedPersona: StarStylist | null
  selectPersona: (slug: string) => void
  clearPersona: () => void

  /** Progress for the selected department (0 when none selected). */
  unlockScore: number
  isUnlocked: (stylist: StarStylist) => boolean
  itemsToUnlock: (stylist: StarStylist) => number
  nextLocked: StarStylist | null

  /** The engine department to send to the collections API for the current
   *  department + stylist (Women / Men / Kids - Girls / Kids - Boys). */
  collectionsDepartment: string | null

  /** Record one cart-add against the current department. Returns any stylists
   *  that just crossed their unlock threshold, for the celebration. */
  recordCartAdd: () => Promise<StarStylist[]>
}

const StarClosetContext = createContext<StarClosetContextValue | null>(null)

export function StarClosetProvider({ children }: { children: ReactNode }) {
  const { user, isAuthenticated } = useAuth()

  const { data: deptData, isLoading } = useQuery<StarDepartment[]>({
    queryKey: ['starcloset-departments'],
    queryFn: () => apiFetch<StarDepartment[]>('/starcloset/departments'),
    staleTime: 10 * 60 * 1000,
  })
  const departments = useMemo(() => deptData ?? [], [deptData])

  // Progress lives on the backend and is only meaningful for a real identity.
  const { data: progressData, refetch: refetchProgress } = useQuery<ProgressResponse>({
    queryKey: ['starcloset-progress', user?.id],
    queryFn: () => apiFetch<ProgressResponse>('/starcloset/progress'),
    enabled: isAuthenticated,
    staleTime: 30 * 1000,
  })
  // Optimistic local mirror so unlocks feel instant even before the refetch.
  const [progress, setProgress] = useState<Record<string, number>>({})
  useEffect(() => {
    if (progressData?.progress) setProgress(progressData.progress)
  }, [progressData])

  const [selectedDepartment, setSelectedDepartment] = useState<string | null>(null)
  const [picks, setPicks] = useState<Record<string, string>>({})

  // Reload selection when identity changes.
  useEffect(() => {
    setSelectedDepartment(localStorage.getItem(deptKey(user?.id)))
    setPicks(readJSON<Record<string, string>>(picksKey(user?.id), {}))
  }, [user?.id])

  const selectDepartment = useCallback(
    (key: string) => {
      setSelectedDepartment(key)
      localStorage.setItem(deptKey(user?.id), key)
    },
    [user?.id],
  )

  const clearDepartment = useCallback(() => {
    setSelectedDepartment(null)
    localStorage.removeItem(deptKey(user?.id))
  }, [user?.id])

  const selectPersona = useCallback(
    (slug: string) => {
      if (!selectedDepartment) return
      setPicks((prev) => {
        const next = { ...prev, [selectedDepartment]: slug }
        localStorage.setItem(picksKey(user?.id), JSON.stringify(next))
        return next
      })
    },
    [selectedDepartment, user?.id],
  )

  const clearPersona = useCallback(() => {
    if (!selectedDepartment) return
    setPicks((prev) => {
      const next = { ...prev }
      delete next[selectedDepartment]
      localStorage.setItem(picksKey(user?.id), JSON.stringify(next))
      return next
    })
  }, [selectedDepartment, user?.id])

  const currentDept = useMemo(
    () => departments.find((d) => d.key === selectedDepartment) ?? null,
    [departments, selectedDepartment],
  )

  const selectedSlug = selectedDepartment ? (picks[selectedDepartment] ?? null) : null

  const unlockScore = selectedDepartment ? (progress[selectedDepartment] ?? 0) : 0

  const selectedPersona = useMemo(() => {
    const p = currentDept?.stylists.find((s) => s.slug === selectedSlug) ?? null
    // A persisted pick is honoured only while it remains unlocked at the current
    // progress. A stale or now-locked pick falls back to the progression map,
    // so a locked stylist can never appear as the active one.
    if (!p || unlockScore < p.unlockAtItems) return null
    return p
  }, [currentDept, selectedSlug, unlockScore])

  const isUnlocked = useCallback(
    (stylist: StarStylist) => unlockScore >= stylist.unlockAtItems,
    [unlockScore],
  )
  const itemsToUnlock = useCallback(
    (stylist: StarStylist) => Math.max(0, stylist.unlockAtItems - unlockScore),
    [unlockScore],
  )

  const nextLocked = useMemo(() => {
    if (!currentDept) return null
    return (
      currentDept.stylists
        .filter((s) => unlockScore < s.unlockAtItems)
        .sort((a, b) => a.unlockAtItems - b.unlockAtItems)[0] ?? null
    )
  }, [currentDept, unlockScore])

  const collectionsDepartment = useMemo(() => {
    if (!currentDept) return null
    if (currentDept.collectionsDepartment) return currentDept.collectionsDepartment
    // Kids: resolve girls/boys from the chosen stylist's gender.
    return selectedPersona?.defaultGender === 'Boys' ? 'Kids - Boys' : 'Kids - Girls'
  }, [currentDept, selectedPersona])

  const recordCartAdd = useCallback(async (): Promise<StarStylist[]> => {
    if (!selectedDepartment || !currentDept) return []
    const before = progress[selectedDepartment] ?? 0
    try {
      const res = await apiFetch<ProgressResponse>('/starcloset/progress', {
        method: 'POST',
        body: JSON.stringify({ department: selectedDepartment, delta: 1 }),
      })
      const after = res.progress[selectedDepartment] ?? before
      setProgress(res.progress)
      refetchProgress()
      // Stylists whose threshold sits in (before, after] just unlocked.
      return currentDept.stylists.filter(
        (s) => s.unlockAtItems > before && s.unlockAtItems <= after,
      )
    } catch {
      return []
    }
  }, [selectedDepartment, currentDept, progress, refetchProgress])

  const value: StarClosetContextValue = {
    departments,
    loading: isLoading,
    progress,
    selectedDepartment,
    currentDept,
    selectDepartment,
    clearDepartment,
    selectedSlug,
    selectedPersona,
    selectPersona,
    clearPersona,
    unlockScore,
    isUnlocked,
    itemsToUnlock,
    nextLocked,
    collectionsDepartment,
    recordCartAdd,
  }

  return <StarClosetContext.Provider value={value}>{children}</StarClosetContext.Provider>
}

export function useStarCloset(): StarClosetContextValue {
  const ctx = useContext(StarClosetContext)
  if (!ctx) throw new Error('useStarCloset must be used within a StarClosetProvider')
  return ctx
}
