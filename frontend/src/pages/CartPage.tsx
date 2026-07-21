import { useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import {
  CheckCircle2,
  ChevronDown,
  Layers,
  ShoppingBag,
  Sparkles,
  Star,
  Trash2,
} from 'lucide-react'
import { useShop } from '@/context/AppContext'
import { formatINR, onImgError} from '@/lib/utils'
import { cn } from '@/lib/utils'
import type { CartCollection } from '@/types'

// ─── Checkout confirmation ──────────────────────────────────────────────────
function CheckoutSuccess({ onReset }: { onReset: () => void }) {
  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center px-5 py-20 text-center">
      <motion.div
        initial={{ scale: 0.8, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ type: 'spring', damping: 18 }}
        className="flex h-20 w-20 items-center justify-center rounded-full bg-gradient-to-br from-brand-100 to-brand-50"
      >
        <CheckCircle2 className="h-10 w-10 text-brand-500" />
      </motion.div>
      <h2 className="mt-6 font-display text-2xl font-bold text-ink-900">
        Your curated collection is ready.
      </h2>
      <p className="mt-3 max-w-sm text-sm text-ink-500">
        Bharat AI has curated the perfect picks for you. Head to Myntra to complete your order.
      </p>
      <button
        onClick={onReset}
        className="mt-8 rounded-full bg-ink-900 px-8 py-3 text-sm font-semibold text-white hover:bg-brand-600"
      >
        Start New Shopping Session
      </button>
    </div>
  )
}

// ─── Individual product row (within accordion) ──────────────────────────────
function ProductRow({
  product,
  onRemove,
}: {
  product: { product: { id: string; brand: string; title: string; image: string; price: number; mrp: number; rating: number; ratingCount: number; category: string }; quantity: number }
  onRemove: () => void
}) {
  return (
    <div className="flex items-center gap-3 px-5 py-3">
      <img
        src={product.product.image}
        alt={product.product.title}
        className="h-14 w-14 shrink-0 rounded-lg object-cover" onError={onImgError} />
      <div className="min-w-0 flex-1">
        <p className="text-[10px] font-bold uppercase tracking-wide text-ink-400">
          {product.product.category}
        </p>
        <p className="truncate text-xs font-semibold text-ink-900">{product.product.brand}</p>
        <p className="truncate text-[11px] text-ink-500">{product.product.title}</p>
        <div className="mt-0.5 flex items-center gap-1.5">
          <span className="flex items-center gap-0.5 text-[10px] font-bold text-emerald-700">
            {product.product.rating.toFixed(1)}
            <Star className="h-2 w-2 fill-emerald-700" />
          </span>
          <span className="text-[10px] text-ink-400">({product.product.ratingCount})</span>
        </div>
      </div>
      <div className="shrink-0 text-right">
        <p className="text-xs font-bold text-ink-900">{formatINR(product.product.price)}</p>
        {product.product.mrp > product.product.price && (
          <p className="text-[10px] text-ink-300 line-through">
            {formatINR(product.product.mrp)}
          </p>
        )}
        {product.quantity > 1 && (
          <p className="text-[10px] text-ink-400">× {product.quantity}</p>
        )}
      </div>
      <button
        onClick={onRemove}
        className="ml-1 flex h-7 w-7 shrink-0 items-center justify-center rounded-full border border-line text-ink-300 transition-colors hover:border-red-300 hover:text-red-500"
      >
        <Trash2 className="h-3 w-3" />
      </button>
    </div>
  )
}

// ─── Individual products accordion card ─────────────────────────────────────
function IndividualProductsCard() {
  const { cartItems, removeFromCart, clearCart, cartTotal } = useShop()
  const [expanded, setExpanded] = useState(true)

  if (cartItems.length === 0) return null

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      className="overflow-hidden rounded-xl border border-line bg-white shadow-sm"
    >
      {/* Header */}
      <div className="flex items-center gap-4 p-5">
        {/* Mini image strip of first 3 products */}
        <div className="flex h-16 w-24 shrink-0 overflow-hidden rounded-lg">
          {cartItems.slice(0, 3).map((ci, j) => (
            <div key={j} className="flex-1 overflow-hidden">
              <img
                src={ci.product.image}
                alt={ci.product.title}
                className="h-full w-full object-cover" onError={onImgError} />
            </div>
          ))}
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <ShoppingBag className="h-3.5 w-3.5 text-brand-500" />
            <p className="font-display text-sm font-bold text-ink-900">Individual Products</p>
          </div>
          <p className="text-xs text-ink-400 mt-0.5">
            {cartItems.length} item{cartItems.length !== 1 ? 's' : ''} ·{' '}
            {cartItems.reduce((s, i) => s + i.quantity, 0)} piece
            {cartItems.reduce((s, i) => s + i.quantity, 0) !== 1 ? 's' : ''}
          </p>
          <p className="text-sm font-bold text-brand-600 mt-1">{formatINR(cartTotal)}</p>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <button
            onClick={() => setExpanded((v) => !v)}
            className="flex items-center gap-1 text-xs font-semibold text-ink-500 hover:text-ink-900"
          >
            View Details
            <ChevronDown
              className={`h-3.5 w-3.5 transition-transform ${expanded ? 'rotate-180' : ''}`}
            />
          </button>
          <button
            onClick={clearCart}
            className="flex h-8 w-8 items-center justify-center rounded-full border border-line text-ink-400 hover:border-red-300 hover:text-red-500"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>

      {/* Expanded rows */}
      <AnimatePresence initial={false}>
        {expanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            className="overflow-hidden border-t border-line"
          >
            <div className="divide-y divide-line">
              {cartItems.map((item) => (
                <ProductRow
                  key={item.product.id}
                  product={item}
                  onRemove={() => removeFromCart(item.product.id)}
                />
              ))}
            </div>
            <div className="flex justify-between border-t border-line px-5 py-3 text-xs font-bold text-ink-900">
              <span>Products Total</span>
              <span className="text-brand-600">{formatINR(cartTotal)}</span>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  )
}

// ─── Collection accordion card ───────────────────────────────────────────────
function CollectionCartCard({
  item,
  isExpanded,
  onToggle,
  onRemove,
  onRemoveItem,
}: {
  item: CartCollection
  isExpanded: boolean
  onToggle: () => void
  onRemove: () => void
  onRemoveItem: (slot: string) => void
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.97 }}
      className="overflow-hidden rounded-xl border border-line bg-white shadow-sm"
    >
      <div className="flex items-center gap-4 p-5">
        <div className="flex h-16 w-24 shrink-0 overflow-hidden rounded-lg">
          {item.collection.items.slice(0, 3).map((ci, j) => (
            <div key={j} className="flex-1 overflow-hidden">
              <img
                src={ci.product.image}
                alt={ci.product.title}
                className="h-full w-full object-cover" onError={onImgError} />
            </div>
          ))}
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <Layers className="h-3.5 w-3.5 text-brand-500" />
            <p className="font-display text-sm font-bold text-ink-900 truncate">
              {item.collection.title}
            </p>
          </div>
          <p className="text-xs text-ink-400 mt-0.5">
            {item.collection.department} · {item.collection.items.length} pieces
          </p>
          <p className="text-sm font-bold text-brand-600 mt-1">
            {formatINR(item.collection.totalPrice)}
          </p>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <button
            onClick={onToggle}
            className="flex items-center gap-1 text-xs font-semibold text-ink-500 hover:text-ink-900"
          >
            View Details
            <ChevronDown
              className={`h-3.5 w-3.5 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
            />
          </button>
          <button
            onClick={onRemove}
            className="flex h-8 w-8 items-center justify-center rounded-full border border-line text-ink-400 hover:border-red-300 hover:text-red-500"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>

      <AnimatePresence initial={false}>
        {isExpanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            className="overflow-hidden border-t border-line"
          >
            <div className="divide-y divide-line">
              {item.collection.items.map((ci) => (
                <div key={ci.slot} className="flex items-center gap-3 px-5 py-3">
                  <img
                    src={ci.product.image}
                    alt={ci.product.title}
                    className="h-12 w-12 shrink-0 rounded-lg object-cover" onError={onImgError} />
                  <div className="min-w-0 flex-1">
                    <p className="text-[10px] font-bold uppercase tracking-wide text-ink-400">
                      {ci.slot}
                    </p>
                    <p className="truncate text-xs font-semibold text-ink-900">
                      {ci.product.brand}
                    </p>
                    <p className="truncate text-[11px] text-ink-500">{ci.product.title}</p>
                  </div>
                  <p className="shrink-0 text-xs font-bold text-ink-900">
                    {formatINR(ci.product.price)}
                  </p>
                  <button
                    onClick={() => onRemoveItem(ci.slot)}
                    title={`Remove ${ci.slot}`}
                    className="ml-1 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-line text-ink-300 transition-colors hover:border-red-300 hover:text-red-500"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </div>
              ))}
            </div>
            <div className="flex justify-between border-t border-line px-5 py-3 text-xs font-bold text-ink-900">
              <span>Collection Total</span>
              <span className="text-brand-600">
                {formatINR(item.collection.totalPrice)}
              </span>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  )
}

// ─── Main CartPage ───────────────────────────────────────────────────────────
export function CartPage() {
  const {
    cartItems,
    cartTotal,
    clearCart,
    collectionCart,
    removeCollectionFromCart,
    removeItemFromCollection,
    clearCollectionCart,
    collectionGrandTotal,
  } = useShop()

  const [checkedOut, setCheckedOut] = useState(false)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())

  const toggleCollection = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const isEmpty = cartItems.length === 0 && collectionCart.length === 0
  const grandTotal = cartTotal + collectionGrandTotal
  const totalPieces =
    cartItems.reduce((s, i) => s + i.quantity, 0) +
    collectionCart.reduce((s, i) => s + i.collection.items.length, 0)

  const handleCheckout = () => setCheckedOut(true)

  const handleReset = () => {
    setCheckedOut(false)
    clearCart()
    clearCollectionCart()
  }

  if (checkedOut) return <CheckoutSuccess onReset={handleReset} />

  return (
    <div className="mx-auto max-w-4xl px-5 py-10 lg:px-8">
      <div className="flex flex-col gap-2">
        <p className="text-xs font-semibold uppercase tracking-widest text-brand-500">
          Your cart
        </p>
        <h1 className="text-3xl font-bold tracking-tight text-ink-900 sm:text-4xl">Cart</h1>
        <p className="max-w-xl text-ink-500">
          Individual picks and curated collections — all in one place.
        </p>
      </div>

      {isEmpty ? (
        <div className="mt-16 flex flex-col items-center justify-center rounded-2xl border border-dashed border-line py-24 text-center">
          <ShoppingBag className="h-12 w-12 text-ink-200" />
          <p className="mt-4 font-display text-lg font-semibold text-ink-900">Your cart is empty</p>
          <p className="mt-2 text-sm text-ink-500">
            Add individual products or complete collections from the AI assistant.
          </p>
        </div>
      ) : (
        <div className="mt-8 grid grid-cols-1 gap-8 lg:grid-cols-3">
          {/* Left — cart items */}
          <div className="space-y-4 lg:col-span-2">
            {/* Individual products as one accordion card */}
            <IndividualProductsCard />

            {/* Collection cards */}
            <AnimatePresence>
              {collectionCart.map((item) => (
                <CollectionCartCard
                  key={item.id}
                  item={item}
                  isExpanded={expanded.has(item.id)}
                  onToggle={() => toggleCollection(item.id)}
                  onRemove={() => removeCollectionFromCart(item.id)}
                  onRemoveItem={(slot) => removeItemFromCollection(item.id, slot)}
                />
              ))}
            </AnimatePresence>
          </div>

          {/* Right — order summary */}
          <div className="lg:col-span-1">
            <div className="sticky top-24 rounded-2xl border border-line bg-white p-6 shadow-sm">
              <h3 className="font-display text-base font-bold text-ink-900">Order Summary</h3>

              <div className="mt-4 space-y-2.5">
                {cartItems.length > 0 && (
                  <div className="flex justify-between text-sm text-ink-600">
                    <span className="flex items-center gap-1.5">
                      <ShoppingBag className="h-3.5 w-3.5" />
                      Individual Products ({cartItems.length})
                    </span>
                    <span className="font-semibold">{formatINR(cartTotal)}</span>
                  </div>
                )}
                {collectionCart.map((item) => (
                  <div key={item.id} className="flex justify-between text-sm text-ink-600">
                    <span className="flex items-center gap-1.5 min-w-0">
                      <Layers className="h-3.5 w-3.5 shrink-0" />
                      <span className="truncate">{item.collection.title}</span>
                    </span>
                    <span className="ml-2 shrink-0 font-semibold">
                      {formatINR(item.collection.totalPrice)}
                    </span>
                  </div>
                ))}

                <div className="flex justify-between text-sm text-ink-400">
                  <span>Total Pieces</span>
                  <span>{totalPieces}</span>
                </div>

                <div className="mt-2 border-t border-line pt-3 flex justify-between text-base font-bold text-ink-900">
                  <span>Grand Total</span>
                  <span className="text-brand-600">{formatINR(grandTotal)}</span>
                </div>
              </div>

              <button
                onClick={handleCheckout}
                className="mt-6 flex w-full items-center justify-center gap-2 rounded-full bg-brand-500 py-3 text-sm font-bold text-white transition-colors hover:bg-brand-600"
              >
                <Sparkles className="h-4 w-4" />
                Checkout
              </button>

              <button
                onClick={() => { clearCart(); clearCollectionCart() }}
                className="mt-3 w-full py-2 text-xs font-semibold text-ink-400 hover:text-red-500"
              >
                Clear Everything
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
