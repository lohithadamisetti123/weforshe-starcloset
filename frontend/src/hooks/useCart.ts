import { useCallback, useEffect, useState } from 'react'
import type { Product } from '@/types'

interface CartItem {
  product: Product
  quantity: number
  addedAt: string
}

const CART_KEY = 'myntra-cart'

function loadCart(): CartItem[] {
  try {
    const raw = localStorage.getItem(CART_KEY)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

function saveCart(items: CartItem[]) {
  try {
    localStorage.setItem(CART_KEY, JSON.stringify(items))
  } catch {}
}

// useCart manages a client-side cart with localStorage persistence.
// The backend doesn't have a cart/orders endpoint (by design — this is a
// shopping assistant demo, not a checkout flow), so cart state lives on the
// client and survives page refreshes via localStorage.
export function useCart() {
  const [items, setItems] = useState<CartItem[]>(loadCart)

  // Persist every change to localStorage
  useEffect(() => {
    saveCart(items)
  }, [items])

  const isInCart = useCallback(
    (productId: string) => items.some((i) => i.product.id === productId),
    [items],
  )

  const add = useCallback((product: Product) => {
    setItems((prev) => {
      const existing = prev.find((i) => i.product.id === product.id)
      if (existing) {
        return prev.map((i) =>
          i.product.id === product.id ? { ...i, quantity: i.quantity + 1 } : i,
        )
      }
      return [...prev, { product, quantity: 1, addedAt: new Date().toISOString() }]
    })
  }, [])

  const remove = useCallback((productId: string) => {
    setItems((prev) => prev.filter((i) => i.product.id !== productId))
  }, [])

  const clear = useCallback(() => setItems([]), [])

  const total = items.reduce((sum, i) => sum + i.product.price * i.quantity, 0)
  const count = items.reduce((sum, i) => sum + i.quantity, 0)

  return { items, isInCart, add, remove, clear, total, count }
}
