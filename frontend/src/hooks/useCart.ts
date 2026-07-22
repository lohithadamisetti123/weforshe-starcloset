import { useCallback, useEffect, useState } from 'react'
import { apiFetch } from '@/lib/api'
import { useAuth } from '@/context/AuthContext'
import type { Product } from '@/types'

export interface CartItem {
  id?: string
  product: Product
  quantity: number
  addedAt: string
}

// useCart manages the signed-in shopper's cart.
//
// The cart is stored per user on the backend (cart_items.user_id), not in
// localStorage. That is the whole point: a browser-global localStorage key made
// every account on the same machine share one cart, so signing in as a
// different user showed the previous user's items. Now the cart is fetched for
// the authenticated identity and every mutation is scoped server-side to that
// user id.
//
// The public API is unchanged (items / isInCart / add / remove / clear / total /
// count) so no UI component needed to change. Mutations apply optimistically for
// the same instant feel the local version had, then sync in the background and
// roll back if the server rejects them.
export function useCart() {
  const { isAuthenticated } = useAuth()
  const [items, setItems] = useState<CartItem[]>([])

  // Load this user's cart, and drop it from memory the moment they are no
  // longer authenticated — logging out must not leave the next user looking at
  // the previous one's items.
  useEffect(() => {
    if (!isAuthenticated) {
      setItems([])
      return
    }
    let alive = true
    apiFetch<CartItem[]>('/cart')
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
    (productId: string) => items.some((i) => i.product.id === productId),
    [items],
  )

  const add = useCallback((product: Product) => {
    // Optimistic: add or increment locally, then persist.
    setItems((prev) => {
      const existing = prev.find((i) => i.product.id === product.id)
      if (existing) {
        return prev.map((i) =>
          i.product.id === product.id ? { ...i, quantity: i.quantity + 1 } : i,
        )
      }
      return [...prev, { product, quantity: 1, addedAt: new Date().toISOString() }]
    })

    apiFetch<CartItem>('/cart', {
      method: 'POST',
      body: JSON.stringify({ productId: product.id }),
    })
      .then((saved) => {
        // Adopt the server's row (real id and authoritative quantity).
        setItems((prev) => prev.map((i) => (i.product.id === product.id ? saved : i)))
      })
      .catch(() => {
        // Roll back the optimistic change.
        setItems((prev) => {
          const existing = prev.find((i) => i.product.id === product.id)
          if (!existing) return prev
          if (existing.quantity <= 1) return prev.filter((i) => i.product.id !== product.id)
          return prev.map((i) =>
            i.product.id === product.id ? { ...i, quantity: i.quantity - 1 } : i,
          )
        })
      })
  }, [])

  const remove = useCallback((productId: string) => {
    const removed = items.find((i) => i.product.id === productId)
    setItems((prev) => prev.filter((i) => i.product.id !== productId))

    apiFetch(`/cart/${productId}`, { method: 'DELETE' }).catch(() => {
      if (!removed) return
      setItems((prev) => (prev.some((i) => i.product.id === productId) ? prev : [...prev, removed]))
    })
  }, [items])

  const clear = useCallback(() => {
    const snapshot = items
    setItems([])
    apiFetch('/cart', { method: 'DELETE' }).catch(() => setItems(snapshot))
  }, [items])

  const total = items.reduce((sum, i) => sum + i.product.price * i.quantity, 0)
  const count = items.reduce((sum, i) => sum + i.quantity, 0)

  return { items, isInCart, add, remove, clear, total, count }
}
