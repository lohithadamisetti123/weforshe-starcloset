import { useCallback, useEffect, useState } from 'react'
import type { CartCollection, CollectionResponse } from '@/types'

const KEY = 'myntra-collection-cart'

function load(): CartCollection[] {
  try { return JSON.parse(localStorage.getItem(KEY) ?? '[]') }
  catch { return [] }
}

function persist(items: CartCollection[]) {
  try { localStorage.setItem(KEY, JSON.stringify(items)) }
  catch {}
}

export function useCollectionCart() {
  const [items, setItems] = useState<CartCollection[]>(load)

  useEffect(() => { persist(items) }, [items])

  const isInCart = useCallback(
    (collectionTitle: string, department: string) =>
      items.some(
        (i) => i.collection.title === collectionTitle && i.collection.department === department,
      ),
    [items],
  )

  const add = useCallback((collection: CollectionResponse): boolean => {
    const duplicate = items.some(
      (i) => i.collection.title === collection.title && i.collection.department === collection.department,
    )
    if (duplicate) return false
    setItems((prev) => [
      ...prev,
      { id: crypto.randomUUID(), collection, addedAt: new Date().toISOString() },
    ])
    return true
  }, [items])

  const remove = useCallback((id: string) => {
    setItems((prev) => prev.filter((i) => i.id !== id))
  }, [])

  // Remove a single slot item from a collection without removing the whole
  // collection. If removing the item leaves the collection empty, the entire
  // collection entry is dropped automatically.
  const removeItemFromCollection = useCallback((cartId: string, slot: string) => {
    setItems((prev) =>
      prev
        .map((entry) => {
          if (entry.id !== cartId) return entry
          const remaining = entry.collection.items.filter((i) => i.slot !== slot)
          if (remaining.length === 0) return null
          return {
            ...entry,
            collection: {
              ...entry.collection,
              items: remaining,
              totalPrice: remaining.reduce((s, i) => s + i.product.price, 0),
            },
          }
        })
        .filter((entry): entry is NonNullable<typeof entry> => entry !== null),
    )
  }, [])

  const clear = useCallback(() => setItems([]), [])

  const grandTotal = items.reduce((sum, i) => sum + i.collection.totalPrice, 0)

  return { items, isInCart, add, remove, removeItemFromCollection, clear, grandTotal, count: items.length }
}
