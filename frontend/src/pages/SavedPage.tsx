import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Bookmark, Heart, ShoppingBag, Sparkles, Trash2 } from 'lucide-react'
import { useShop } from '@/context/AppContext'
import { formatINR, onImgError} from '@/lib/utils'
import { Toast } from '@/components/ui/Toast'
import type { SavedCollection } from '@/types'

export function SavedPage() {
  const { savedCollections, removeSavedCollection, addCollectionToCart, isCollectionInCart } = useShop()
  const [toast, setToast] = useState<string | null>(null)
  const [detail, setDetail] = useState<SavedCollection | null>(null)

  const showToast = (msg: string) => { setToast(null); setTimeout(() => setToast(msg), 10) }

  const handleMoveToCart = (s: SavedCollection) => {
    const added = addCollectionToCart(s.collection)
    showToast(added ? 'Collection moved to cart!' : 'Already in cart.')
  }

  return (
    <div className="mx-auto max-w-7xl px-5 py-10 lg:px-8">
      <div className="flex flex-col gap-2">
        <p className="text-xs font-semibold uppercase tracking-widest text-brand-500">Your saves</p>
        <h1 className="text-3xl font-bold tracking-tight text-ink-900 sm:text-4xl">Saved Collections</h1>
        <p className="max-w-xl text-ink-500">
          Collections you've loved. Move them to cart whenever you're ready.
        </p>
      </div>

      <div className="mt-8">
        {savedCollections.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-line py-24 text-center">
            <Heart className="h-12 w-12 text-ink-200" />
            <p className="mt-4 font-display text-lg font-semibold text-ink-900">Nothing saved yet</p>
            <p className="mt-2 text-sm text-ink-500">
              Tap the heart on any collection in the assistant to save it here.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            <AnimatePresence>
              {savedCollections.map((s, i) => (
                <motion.div
                  key={s.id}
                  initial={{ opacity: 0, y: 16 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  transition={{ delay: i * 0.05 }}
                  className="overflow-hidden rounded-xl border border-line bg-white shadow-sm"
                >
                  {/* Product image strip */}
                  <div className="flex h-24 overflow-hidden bg-surface">
                    {s.collection.items.slice(0, 4).map((item, j) => (
                      <div key={j} className="flex-1 overflow-hidden">
                        <img
                          src={item.product.image}
                          alt={item.product.title}
                          className="h-full w-full object-cover" onError={onImgError} />
                      </div>
                    ))}
                  </div>

                  <div className="p-4">
                    {/* Name + category badge */}
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="font-display text-sm font-bold text-ink-900">{s.name}</p>
                        <p className="mt-0.5 text-[11px] text-ink-400">{s.department}</p>
                      </div>
                      <span className="shrink-0 rounded-full bg-brand-50 px-2 py-0.5 text-[10px] font-bold text-brand-600">
                        {formatINR(s.collection.totalPrice)}
                      </span>
                    </div>

                    {/* Prompt */}
                    <p className="mt-2 line-clamp-2 text-[11px] italic text-ink-400">"{s.prompt}"</p>

                    {/* Item slots */}
                    <div className="mt-3 flex flex-wrap gap-1">
                      {s.collection.items.map((it) => (
                        <span
                          key={it.slot}
                          className="rounded bg-surface px-2 py-0.5 text-[10px] font-medium text-ink-600"
                        >
                          {it.slot}
                        </span>
                      ))}
                    </div>

                    {/* Saved time */}
                    <p className="mt-3 text-[10px] text-ink-300">
                      Saved {new Date(s.savedAt).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })}
                    </p>

                    {/* Buttons */}
                    <div className="mt-4 flex gap-2">
                      <button
                        onClick={() => setDetail(s)}
                        className="flex flex-1 items-center justify-center gap-1.5 rounded-full border border-line py-2 text-xs font-semibold text-ink-700 hover:border-ink-900"
                      >
                        <Bookmark className="h-3.5 w-3.5" />
                        View
                      </button>
                      <button
                        onClick={() => handleMoveToCart(s)}
                        className={`flex flex-1 items-center justify-center gap-1.5 rounded-full py-2 text-xs font-semibold transition-colors ${
                          isCollectionInCart(s.collection.title, s.collection.department)
                            ? 'bg-emerald-50 text-emerald-700'
                            : 'bg-ink-900 text-white hover:bg-brand-600'
                        }`}
                      >
                        <ShoppingBag className="h-3.5 w-3.5" />
                        {isCollectionInCart(s.collection.title, s.collection.department) ? 'In Cart ✓' : 'Move to Cart'}
                      </button>
                      <button
                        onClick={() => removeSavedCollection(s.id)}
                        className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-line text-ink-400 transition-colors hover:border-red-300 hover:text-red-500"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}
      </div>

      {/* Detail modal */}
      <AnimatePresence>
        {detail && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setDetail(null)}
            className="fixed inset-0 z-50 flex items-center justify-center bg-ink-900/50 p-4"
          >
            <motion.div
              initial={{ scale: 0.96, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.96, opacity: 0 }}
              onClick={(e) => e.stopPropagation()}
              className="no-scrollbar max-h-[85vh] w-full max-w-md overflow-y-auto rounded-2xl bg-white p-6 shadow-2xl"
            >
              <div className="flex items-start justify-between gap-2">
                <div>
                  <h3 className="font-display text-lg font-bold text-ink-900">{detail.name}</h3>
                  <p className="mt-0.5 text-xs text-ink-400">{detail.department} · {formatINR(detail.collection.totalPrice)}</p>
                </div>
                <button onClick={() => setDetail(null)} className="text-ink-400 hover:text-ink-900">✕</button>
              </div>
              <div className="mt-5 space-y-4">
                {detail.collection.items.map((item) => (
                  <div key={item.slot} className="flex items-center gap-3">
                    <img src={item.product.image} alt={item.product.title} className="h-14 w-14 rounded-lg object-cover" onError={onImgError} />
                    <div className="min-w-0">
                      <p className="text-[10px] font-bold uppercase tracking-wide text-ink-400">{item.slot}</p>
                      <p className="truncate text-xs font-semibold text-ink-900">{item.product.brand}</p>
                      <p className="truncate text-[11px] text-ink-500">{item.product.title}</p>
                      <p className="text-xs font-bold text-brand-600">{formatINR(item.product.price)}</p>
                    </div>
                  </div>
                ))}
              </div>
              {detail.collection.explanation?.length > 0 && (
                <div className="mt-5 rounded-lg bg-brand-50 p-3">
                  <div className="flex items-center gap-1.5 mb-2">
                    <Sparkles className="h-3.5 w-3.5 text-brand-500" />
                    <p className="text-xs font-bold text-ink-900">Why this collection?</p>
                  </div>
                  <ul className="space-y-1">
                    {detail.collection.explanation.map((r, i) => (
                      <li key={i} className="text-xs text-ink-700">· {r}</li>
                    ))}
                  </ul>
                </div>
              )}
              <button
                onClick={() => { handleMoveToCart(detail); setDetail(null) }}
                className="mt-5 flex w-full items-center justify-center gap-2 rounded-full bg-ink-900 py-3 text-sm font-semibold text-white hover:bg-brand-600"
              >
                <ShoppingBag className="h-4 w-4" />
                Move to Cart
              </button>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      <Toast message={toast} onDismiss={() => setToast(null)} />
    </div>
  )
}
