import { useCallback, useEffect, useState } from 'react'
import { apiFetch } from '@/lib/api'

interface FavoriteResponse {
  id: string
  productId: string
  product: { id: string }
}

// useFavorites manages the full lifecycle of saved products:
// - loads existing favorites on mount (so the heart icon is correctly filled
//   across page refreshes and sessions — the backend is the source of truth)
// - optimistically updates local state on toggle so the UI feels instant
// - syncs adds/removes to the backend in the background
export function useFavorites() {
  // Set of product IDs the current session has saved
  const [favoriteIds, setFavoriteIds] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(true)

  // Load existing favorites on mount
  useEffect(() => {
    apiFetch<FavoriteResponse[]>('/favorites')
      .then((favs) => {
        setFavoriteIds(new Set(favs.map((f) => f.product?.id ?? f.productId)))
      })
      .catch(() => {
        // Network failure — start with empty set, UI still works locally
      })
      .finally(() => setLoading(false))
  }, [])

  const isFavorite = useCallback(
    (productId: string) => favoriteIds.has(productId),
    [favoriteIds],
  )

  const toggle = useCallback(async (productId: string) => {
    const wasFavorite = favoriteIds.has(productId)

    // Optimistic update — flip immediately so button feels responsive
    setFavoriteIds((prev) => {
      const next = new Set(prev)
      wasFavorite ? next.delete(productId) : next.add(productId)
      return next
    })

    try {
      if (wasFavorite) {
        await apiFetch(`/favorites/${productId}`, { method: 'DELETE' })
      } else {
        await apiFetch('/favorites', {
          method: 'POST',
          body: JSON.stringify({ productId }),
        })
      }
    } catch {
      // Revert optimistic update on failure
      setFavoriteIds((prev) => {
        const next = new Set(prev)
        wasFavorite ? next.add(productId) : next.delete(productId)
        return next
      })
    }
  }, [favoriteIds])

  return { isFavorite, toggle, loading, count: favoriteIds.size }
}
