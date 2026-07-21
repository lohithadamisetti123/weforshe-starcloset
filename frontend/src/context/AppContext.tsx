import { createContext, useContext, type ReactNode } from 'react'
import { useCart } from '@/hooks/useCart'
import { useFavorites } from '@/hooks/useFavorites'
import { useChatHistory } from '@/hooks/useChatHistory'
import { useSavedCollections } from '@/hooks/useSavedCollections'
import { useCollectionCart } from '@/hooks/useCollectionCart'
import type { Product, CollectionResponse } from '@/types'

interface AppContextValue {
  // Product-level cart (localStorage)
  cartItems: ReturnType<typeof useCart>['items']
  isInCart: (productId: string) => boolean
  addToCart: (product: Product) => void
  removeFromCart: (productId: string) => void
  clearCart: () => void
  cartCount: number
  cartTotal: number
  // Product favorites (backend)
  isFavorite: (productId: string) => boolean
  toggleFavorite: (productId: string) => void
  favoritesCount: number
  // AI Chat history
  chatHistory: ReturnType<typeof useChatHistory>['entries']
  addChatHistory: ReturnType<typeof useChatHistory>['addEntry']
  removeChatHistory: ReturnType<typeof useChatHistory>['remove']
  clearChatHistory: ReturnType<typeof useChatHistory>['clear']
  // Saved collections
  savedCollections: ReturnType<typeof useSavedCollections>['saved']
  isCollectionSaved: (title: string, dept: string) => boolean
  saveCollection: (c: CollectionResponse, prompt: string) => boolean
  removeSavedCollection: (id: string) => void
  savedCount: number
  // Collection cart
  collectionCart: ReturnType<typeof useCollectionCart>['items']
  isCollectionInCart: (title: string, dept: string) => boolean
  addCollectionToCart: (c: CollectionResponse) => boolean
  removeCollectionFromCart: (id: string) => void
  removeItemFromCollection: (cartId: string, slot: string) => void
  clearCollectionCart: () => void
  collectionCartCount: number
  collectionGrandTotal: number
}

const AppContext = createContext<AppContextValue | null>(null)

export function ShopProvider({ children }: { children: ReactNode }) {
  const cart = useCart()
  const favorites = useFavorites()
  const history = useChatHistory()
  const savedColls = useSavedCollections()
  const collCart = useCollectionCart()

  const value: AppContextValue = {
    cartItems: cart.items,
    isInCart: cart.isInCart,
    addToCart: cart.add,
    removeFromCart: cart.remove,
    clearCart: cart.clear,
    cartCount: cart.count,
    cartTotal: cart.total,
    isFavorite: favorites.isFavorite,
    toggleFavorite: favorites.toggle,
    favoritesCount: favorites.count,
    chatHistory: history.entries,
    addChatHistory: history.addEntry,
    removeChatHistory: history.remove,
    clearChatHistory: history.clear,
    savedCollections: savedColls.saved,
    isCollectionSaved: savedColls.isSaved,
    saveCollection: savedColls.save,
    removeSavedCollection: savedColls.remove,
    savedCount: savedColls.count,
    collectionCart: collCart.items,
    isCollectionInCart: collCart.isInCart,
    addCollectionToCart: collCart.add,
    removeCollectionFromCart: collCart.remove,
    removeItemFromCollection: collCart.removeItemFromCollection,
    clearCollectionCart: collCart.clear,
    collectionCartCount: collCart.count,
    collectionGrandTotal: collCart.grandTotal,
  }

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>
}

export function useShop() {
  const ctx = useContext(AppContext)
  if (!ctx) throw new Error('useShop must be used inside <ShopProvider>')
  return ctx
}
