import { useCallback, useEffect, useState } from 'react'
import { apiFetch } from '@/lib/api'
import { useAuth } from '@/context/AuthContext'
import type { CartCollection, CollectionResponse } from '@/types'

// useCollectionCart manages whole collections the signed-in shopper has added
// to their cart.
//
// Stored per user on the backend (collection_cart_items.user_id) instead of a
// browser-global localStorage key, so two accounts on the same machine no
// longer share one collection cart.
//
// The public API is unchanged (items / isInCart / add / remove /
// removeItemFromCollection / clear / grandTotal / count).
export function useCollectionCart() {
  const { isAuthenticated } = useAuth()
  const [items, setItems] = useState<CartCollection[]>([])

  useEffect(() => {
    if (!isAuthenticated) {
      setItems([])
      return
    }
    let alive = true
    apiFetch<CartCollection[]>('/collection-cart')
      .then((data) => {
        if (alive) setItems(data)
      })
      .catch(() => {
        if (alive) setItems([])
      })
    return () => {
      alive = false
    }
  }, [isAuthenticated])

  const isInCart = useCallback(
    (collectionTitle: string, department: string) =>
      items.some(
        (i) => i.collection.title === collectionTitle && i.collection.department === department,
      ),
    [items],
  )

  const add = useCallback(
    (collection: CollectionResponse): boolean => {
      const duplicate = items.some(
        (i) =>
          i.collection.title === collection.title &&
          i.collection.department === collection.department,
      )
      if (duplicate) return false

      const tempId = crypto.randomUUID()
      setItems((prev) => [...prev, { id: tempId, collection, addedAt: new Date().toISOString() }])

      apiFetch<CartCollection>('/collection-cart', {
        method: 'POST',
        body: JSON.stringify({
          name: collection.title,
          department: collection.department,
          collection,
        }),
      })
        .then((stored) => {
          setItems((prev) => prev.map((i) => (i.id === tempId ? stored : i)))
        })
        .catch(() => {
          setItems((prev) => prev.filter((i) => i.id !== tempId))
        })

      return true
    },
    [items],
  )

  const remove = useCallback((id: string) => {
    const removed = items.find((i) => i.id === id)
    setItems((prev) => prev.filter((i) => i.id !== id))
    apiFetch(`/collection-cart/${id}`, { method: 'DELETE' }).catch(() => {
      if (!removed) return
      setItems((prev) => (prev.some((i) => i.id === id) ? prev : [...prev, removed]))
    })
  }, [items])

  // Remove a single slot item from a collection without removing the whole
  // collection. If removing the item leaves the collection empty, the entire
  // collection entry is dropped automatically.
  const removeItemFromCollection = useCallback((cartId: string, slot: string) => {
    const entry = items.find((i) => i.id === cartId)
    if (!entry) return

    const remaining = entry.collection.items.filter((i) => i.slot !== slot)

    if (remaining.length === 0) {
      // Removing the last slot removes the whole collection entry.
      setItems((prev) => prev.filter((i) => i.id !== cartId))
      apiFetch(`/collection-cart/${cartId}`, { method: 'DELETE' }).catch(() => {})
      return
    }

    const updatedCollection = {
      ...entry.collection,
      items: remaining,
      totalPrice: remaining.reduce((s, i) => s + i.product.price, 0),
    }
    setItems((prev) =>
      prev.map((i) => (i.id === cartId ? { ...i, collection: updatedCollection } : i)),
    )

    // Persist the trimmed snapshot so the change survives a refresh and stays
    // scoped to this user.
    apiFetch(`/collection-cart/${cartId}`, {
      method: 'PUT',
      body: JSON.stringify({ collection: updatedCollection }),
    }).catch(() => {})
  }, [items])

  const clear = useCallback(() => {
    const snapshot = items
    setItems([])
    apiFetch('/collection-cart', { method: 'DELETE' }).catch(() => setItems(snapshot))
  }, [items])

  const grandTotal = items.reduce((sum, i) => sum + i.collection.totalPrice, 0)

  return { items, isInCart, add, remove, removeItemFromCollection, clear, grandTotal, count: items.length }
}
