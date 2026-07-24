import { useState, useCallback } from 'react'
import { motion } from 'framer-motion'
import { ChevronDown, Heart, LayoutGrid, Loader2, Plus, ShoppingBag } from 'lucide-react'
import type { AIRecommendationResponse, CollectionResponse } from '@/types'
import { CollectionCard } from '@/components/assistant/CollectionCard'
import { ResponseCard } from '@/components/assistant/ResponseCard'
import { Toast } from '@/components/ui/Toast'
import { useShop } from '@/context/AppContext'
import { cn } from '@/lib/utils'

interface CollectionsViewProps {
  collections: CollectionResponse[]
  response?: AIRecommendationResponse
  messageId: string
  prompt: string
  isThinking: boolean
  onLoadMore: (messageId: string, prompt: string) => void
  onBrowseAll: (messageId: string, prompt: string) => void
  // onCartAdd fires on each cart addition (a slot product or a whole
  // collection), so the studio can advance the department's unlock journey.
  onCartAdd?: () => void
}

export function CollectionsView({
  collections,
  response,
  messageId,
  prompt,
  isThinking,
  onLoadMore,
  onBrowseAll,
  onCartAdd,
}: CollectionsViewProps) {
  const { saveCollection, isCollectionSaved, addCollectionToCart, isCollectionInCart } = useShop()
  const [expanded, setExpanded] = useState<Set<number>>(new Set([0]))
  const [showBrowse, setShowBrowse] = useState(false)
  const [toast, setToast] = useState<string | null>(null)

  const showToast = useCallback((msg: string) => {
    setToast(null)
    setTimeout(() => setToast(msg), 10)
  }, [])

  const toggleExpand = (i: number) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      next.has(i) ? next.delete(i) : next.add(i)
      return next
    })
  }

  const handleSave = (collection: CollectionResponse) => {
    const added = saveCollection(collection, prompt)
    showToast(added ? 'Collection saved successfully.' : 'Already in your saved collections.')
  }

  const handleAddToCart = (collection: CollectionResponse) => {
    const added = addCollectionToCart(collection)
    showToast(added ? 'Collection added to cart!' : 'Already in cart.')
    if (added) onCartAdd?.()
  }

  const handleBrowseAll = () => {
    setShowBrowse(true)
    onBrowseAll(messageId, prompt)
  }

  return (
    <>
      <div className="space-y-3">
        {collections.map((collection, i) => {
          const saved = isCollectionSaved(collection.title, collection.department)
          const inCart = isCollectionInCart(collection.title, collection.department)

          return (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.08 }}
              className="overflow-hidden rounded-xl border border-line bg-white shadow-sm"
            >
              {/* Accordion header */}
              <button
                onClick={() => toggleExpand(i)}
                className="flex w-full items-center justify-between px-5 py-4 text-left"
              >
                <div className="flex items-center gap-3">
                  <span className="flex h-7 w-7 items-center justify-center rounded-full bg-brand-500 text-xs font-bold text-white">
                    {i + 1}
                  </span>
                  <div>
                    <p className="font-display text-sm font-bold text-ink-900">{collection.title}</p>
                    <p className="text-[11px] text-ink-500">
                      {collection.items.length} pieces ·{' '}
                      <span className="font-semibold text-brand-600">
                        ₹{collection.totalPrice.toLocaleString('en-IN')}
                      </span>
                    </p>
                  </div>
                </div>
                <ChevronDown
                  className={`h-4 w-4 text-ink-400 transition-transform duration-200 ${
                    expanded.has(i) ? 'rotate-180' : ''
                  }`}
                />
              </button>

              {/* Expanded body */}
              {expanded.has(i) && (
                <div className="border-t border-line px-5 pb-5 pt-4">
                  <CollectionCard collection={collection} hideHeader onCartAdd={onCartAdd} />

                  {/* ── Per-collection action buttons ─────────────────── */}
                  <div className="mt-4 flex flex-wrap gap-2">
                    <button
                      onClick={() => handleSave(collection)}
                      className={cn(
                        'flex items-center gap-1.5 rounded-full border px-4 py-2 text-xs font-semibold transition-colors',
                        saved
                          ? 'border-brand-500 bg-brand-50 text-brand-600'
                          : 'border-line text-ink-600 hover:border-brand-500 hover:text-brand-600',
                      )}
                    >
                      <Heart className={cn('h-3.5 w-3.5', saved && 'fill-brand-500 text-brand-500')} />
                      {saved ? 'Saved' : 'Save Collection'}
                    </button>

                    <button
                      onClick={() => handleAddToCart(collection)}
                      className={cn(
                        'flex items-center gap-1.5 rounded-full border px-4 py-2 text-xs font-semibold transition-colors',
                        inCart
                          ? 'border-emerald-500 bg-emerald-50 text-emerald-700'
                          : 'border-ink-900 bg-ink-900 text-white hover:bg-brand-600',
                      )}
                    >
                      <ShoppingBag className="h-3.5 w-3.5" />
                      {inCart ? 'In Cart ✓' : 'Add Collection to Cart'}
                    </button>

                    <button
                      onClick={() => {
                        setExpanded((prev) => {
                          const next = new Set(prev); next.delete(i); return next
                        })
                        onLoadMore(messageId, prompt)
                      }}
                      disabled={isThinking}
                      className="flex items-center gap-1.5 rounded-full border border-line px-4 py-2 text-xs font-semibold text-ink-600 hover:border-ink-900 hover:text-ink-900 disabled:opacity-50"
                    >
                      {isThinking ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
                      Generate Another
                    </button>
                  </div>
                </div>
              )}
            </motion.div>
          )
        })}

        {/* Bottom action row */}
        <div className="flex flex-wrap gap-2 pt-1">
          <button
            onClick={() => onLoadMore(messageId, prompt)}
            disabled={isThinking}
            className="flex items-center gap-1.5 rounded-full border border-brand-500 px-4 py-2 text-xs font-semibold text-brand-600 transition-colors hover:bg-brand-50 disabled:opacity-50"
          >
            {isThinking ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
            More Collections
          </button>

          {!showBrowse && (
            <button
              onClick={handleBrowseAll}
              disabled={isThinking}
              className="flex items-center gap-1.5 rounded-full border border-line px-4 py-2 text-xs font-semibold text-ink-600 transition-colors hover:border-ink-900 hover:text-ink-900 disabled:opacity-50"
            >
              <LayoutGrid className="h-3.5 w-3.5" />
              Browse All Products
            </button>
          )}
        </div>

        {showBrowse && (
          <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }}>
            {response ? (
              <ResponseCard response={response} />
            ) : isThinking ? (
              <div className="flex items-center gap-2 rounded-xl border border-line bg-white px-5 py-4 text-sm text-ink-500">
                <Loader2 className="h-4 w-4 animate-spin text-brand-500" />
                Fetching all products…
              </div>
            ) : null}
          </motion.div>
        )}
      </div>

      <Toast message={toast} onDismiss={() => setToast(null)} />
    </>
  )
}
