import { useCallback, useEffect, useState } from 'react'
import { apiFetch } from '@/lib/api'
import type { UpdateProfileRequest, UserProfile } from '@/types'

// useProfile manages the shopper's personalization profile against the backend,
// which is the source of truth (the profile is stored server-side, keyed to the
// session/user, so it survives refreshes and follows the shopper across the
// app). It mirrors the shape of useFavorites: load on mount, expose a status,
// and sync writes.
//
// Two things are loaded independently:
//   - onboarding status (cheap; drives whether onboarding is shown at all)
//   - the full profile (loaded lazily, for the edit screen)
export function useProfile() {
  const [onboarded, setOnboarded] = useState<boolean | null>(null) // null = unknown/loading
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  // Load onboarding status on mount so the caller can decide whether to gate.
  useEffect(() => {
    let alive = true
    apiFetch<{ onboarded: boolean }>('/profile/status')
      .then((s) => {
        if (alive) setOnboarded(s.onboarded)
      })
      .catch(() => {
        // Treat a status failure as "not onboarded, don't block": the shopper
        // can still use the stylist, and onboarding can be offered later. Never
        // let a personalization lookup stand between a shopper and shopping.
        if (alive) setOnboarded(false)
      })
      .finally(() => {
        if (alive) setLoading(false)
      })
    return () => {
      alive = false
    }
  }, [])

  // loadProfile fetches the full profile (for pre-filling the edit screen).
  const loadProfile = useCallback(async (): Promise<UserProfile | null> => {
    try {
      const p = await apiFetch<UserProfile>('/profile/preferences')
      setProfile(p)
      setOnboarded(p.onboarded)
      return p
    } catch {
      return null
    }
  }, [])

  // save persists the profile. Onboarding passes onboarded: true; the returned
  // profile becomes the new local truth.
  const save = useCallback(async (req: UpdateProfileRequest): Promise<UserProfile | null> => {
    setSaving(true)
    try {
      const saved = await apiFetch<UserProfile>('/profile/preferences', {
        method: 'PUT',
        body: JSON.stringify(req),
      })
      setProfile(saved)
      setOnboarded(saved.onboarded)
      return saved
    } catch {
      return null
    } finally {
      setSaving(false)
    }
  }, [])

  return { onboarded, profile, loading, saving, loadProfile, save }
}
