import { useCallback, useEffect, useState } from 'react'
import type { CollectionResponse, SavedCollection } from '@/types'

const KEY = 'myntra-saved-collections'

function load(): SavedCollection[] {
  try { return JSON.parse(localStorage.getItem(KEY) ?? '[]') }
  catch { return [] }
}

function persist(items: SavedCollection[]) {
  try { localStorage.setItem(KEY, JSON.stringify(items)) }
  catch {}
}

export function useSavedCollections() {
  const [saved, setSaved] = useState<SavedCollection[]>(load)

  useEffect(() => { persist(saved) }, [saved])

  const isSaved = useCallback(
    (collectionTitle: string, department: string) =>
      saved.some((s) => s.name === collectionTitle && s.department === department),
    [saved],
  )

  const save = useCallback((
    collection: CollectionResponse,
    prompt: string,
  ): boolean => {
    // Prevent exact duplicates (same title + department)
    const duplicate = saved.some(
      (s) => s.name === collection.title && s.department === collection.department,
    )
    if (duplicate) return false

    const entry: SavedCollection = {
      id: crypto.randomUUID(),
      name: collection.title,
      prompt,
      department: collection.department,
      collection,
      savedAt: new Date().toISOString(),
    }
    setSaved((prev) => [entry, ...prev])
    return true
  }, [saved])

  const remove = useCallback((id: string) => {
    setSaved((prev) => prev.filter((s) => s.id !== id))
  }, [])

  return { saved, isSaved, save, remove, count: saved.length }
}
