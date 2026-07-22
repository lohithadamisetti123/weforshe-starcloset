import { useCallback, useEffect, useState } from 'react'
import { apiFetch } from '@/lib/api'
import { useAuth } from '@/context/AuthContext'
import type { CollectionResponse, SavedCollection } from '@/types'

// useSavedCollections manages the signed-in shopper's saved collections (the
// "Saved" page).
//
// Stored per user on the backend (saved_collections.user_id) rather than under
// a browser-global localStorage key, which is what previously made every
// account on one machine share a single wishlist.
//
// The public API is unchanged (saved / isSaved / save / remove / count). `save`
// still returns a boolean synchronously — the UI uses it to show a "saved" vs
// "already saved" toast — decided from the local duplicate check, with the
// network write following in the background.
export function useSavedCollections() {
  const { isAuthenticated } = useAuth()
  const [saved, setSaved] = useState<SavedCollection[]>([])

  useEffect(() => {
    if (!isAuthenticated) {
      setSaved([])
      return
    }
    let alive = true
    apiFetch<SavedCollection[]>('/saved-collections')
      .then((data) => {
        if (alive) setSaved(data)
      })
      .catch(() => {
        if (alive) setSaved([])
      })
    return () => {
      alive = false
    }
  }, [isAuthenticated])

  const isSaved = useCallback(
    (collectionTitle: string, department: string) =>
      saved.some((s) => s.name === collectionTitle && s.department === department),
    [saved],
  )

  const save = useCallback(
    (collection: CollectionResponse, prompt: string): boolean => {
      const duplicate = saved.some(
        (s) => s.name === collection.title && s.department === collection.department,
      )
      if (duplicate) return false

      const tempId = crypto.randomUUID()
      const entry: SavedCollection = {
        id: tempId,
        name: collection.title,
        prompt,
        department: collection.department,
        collection,
        savedAt: new Date().toISOString(),
      }
      setSaved((prev) => [entry, ...prev])

      apiFetch<SavedCollection>('/saved-collections', {
        method: 'POST',
        body: JSON.stringify({
          name: collection.title,
          department: collection.department,
          prompt,
          collection,
        }),
      })
        .then((stored) => {
          // Swap the optimistic row for the stored one so `remove` uses the
          // real server id.
          setSaved((prev) => prev.map((s) => (s.id === tempId ? stored : s)))
        })
        .catch(() => {
          setSaved((prev) => prev.filter((s) => s.id !== tempId))
        })

      return true
    },
    [saved],
  )

  const remove = useCallback((id: string) => {
    const removed = saved.find((s) => s.id === id)
    setSaved((prev) => prev.filter((s) => s.id !== id))
    apiFetch(`/saved-collections/${id}`, { method: 'DELETE' }).catch(() => {
      if (!removed) return
      setSaved((prev) => (prev.some((s) => s.id === id) ? prev : [removed, ...prev]))
    })
  }, [saved])

  return { saved, isSaved, save, remove, count: saved.length }
}
