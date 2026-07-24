import { motion } from 'framer-motion'
import { CheckCircle2, IndianRupee, Sparkles, Star, Heart, ShoppingBag, ImageOff } from 'lucide-react'
import { useState } from 'react'
import type { CollectionResponse, Product } from '@/types'
import { formatINR } from '@/lib/utils'
import { cn } from '@/lib/utils'
import { useShop } from '@/context/ShopContext'

function SlotCard({ product, slot, onCartAdd }: { product: Product; slot: string; onCartAdd?: () => void }) {
  const { isFavorite, toggleFavorite, isInCart, addToCart, removeFromCart } = useShop()
  const saved = isFavorite(product.id)
  const inCart = isInCart(product.id)
  const [toasted, setToasted] = useState(false)
  const [imgError, setImgError] = useState(false)

  const handleCart = () => {
    if (inCart) {
      removeFromCart(product.id)
    } else {
      addToCart(product)
      setToasted(true)
      setTimeout(() => setToasted(false), 1500)
      // Only an add advances the department journey — never a remove.
      onCartAdd?.()
    }
  }

  return (
    <div className="relative w-44 shrink-0 overflow-hidden rounded-xl border border-line bg-card shadow-sm">
      <span className="absolute left-2 top-2 z-10 rounded-full bg-ink-900/80 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide text-white">
        {slot}
      </span>
      <button
        onClick={() => toggleFavorite(product.id)}
        className="absolute right-2 top-2 z-10 flex h-7 w-7 items-center justify-center rounded-full bg-white/90 shadow"
      >
        <Heart
          className={cn('h-3.5 w-3.5', saved ? 'fill-brand-500 text-brand-500' : 'text-ink-400')}
        />
      </button>

      <div className="relative aspect-[3/4] overflow-hidden bg-ink-900/5">
        {imgError || !product.image ? (
          // Broken/missing image → clean placeholder instead of the browser
          // rendering overflowing alt text over the card.
          <div className="flex h-full w-full flex-col items-center justify-center gap-1 bg-ink-900/[0.04] text-ink-400">
            <ImageOff className="h-6 w-6" />
            <span className="px-2 text-center text-[9px] font-medium leading-tight">
              {product.title}
            </span>
          </div>
        ) : (
          <img
            src={product.image}
            alt={product.title}
            loading="lazy"
            onError={() => setImgError(true)}
            className="h-full w-full object-cover"
          />
        )}
        {product.matchPercent && (
          <span className="absolute bottom-2 left-2 flex items-center gap-0.5 rounded-full bg-white/95 px-2 py-0.5 text-[9px] font-bold text-brand-600 shadow">
            <Sparkles className="h-2.5 w-2.5" />
            {product.matchPercent}%
          </span>
        )}
        {product.discount > 0 && (
          <span className="absolute right-2 bottom-2 rounded bg-ink-900 px-1.5 py-0.5 text-[9px] font-bold text-white">
            {product.discount}% OFF
          </span>
        )}
      </div>

      <div className="p-2.5">
        <p className="truncate text-[11px] font-bold text-ink-900">{product.brand}</p>
        <p className="truncate text-[10px] text-ink-500">{product.title}</p>
        <div className="mt-1 flex items-center gap-1">
          <span className="flex items-center gap-0.5 rounded bg-emerald-50 px-1 py-0.5 text-[10px] font-bold text-emerald-700">
            {product.rating.toFixed(1)}
            <Star className="h-2 w-2 fill-emerald-700" />
          </span>
          <span className="text-[10px] text-ink-400">({product.ratingCount})</span>
        </div>
        <div className="mt-1 flex items-baseline gap-1">
          <span className="text-xs font-bold text-ink-900">{formatINR(product.price)}</span>
          {product.mrp > product.price && (
            <span className="text-[10px] text-ink-300 line-through">{formatINR(product.mrp)}</span>
          )}
        </div>
        <button
          onClick={handleCart}
          className={cn(
            'mt-2 flex w-full items-center justify-center gap-1 rounded-full py-1.5 text-[10px] font-semibold transition-colors',
            inCart || toasted ? 'bg-emerald-50 text-emerald-700' : 'bg-ink-900 text-white hover:bg-brand-600',
          )}
        >
          <ShoppingBag className="h-2.5 w-2.5" />
          {inCart ? 'Remove ✕' : toasted ? 'Added!' : 'Add to Cart'}
        </button>
      </div>
    </div>
  )
}

interface CollectionCardProps {
  collection: CollectionResponse
  // When true, the title/total header is omitted — used inside CollectionsView
  // which renders its own accordion header row instead.
  hideHeader?: boolean
  // onCartAdd fires when an individual slot is added to the cart. Wired only in
  // the studio (to advance the department journey); other usages leave it unset.
  onCartAdd?: () => void
}

export function CollectionCard({ collection, hideHeader = false, onCartAdd }: CollectionCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Header — shown when used standalone, hidden inside CollectionsView */}
      {!hideHeader && (
        <div className="mb-4 flex items-center justify-between gap-3">
          <h3 className="font-display text-base font-bold text-ink-900">{collection.title}</h3>
          <span className="flex shrink-0 items-center gap-1 rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-600">
            <IndianRupee className="h-3 w-3" />
            {formatINR(collection.totalPrice)}
          </span>
        </div>
      )}

      {/* Budget label — always shown */}
      {collection.budgetLimit != null && (
        <p className="mb-3 text-xs text-ink-500">
          Budget: {formatINR(collection.budgetLimit)} ·{' '}
          {collection.totalPrice <= collection.budgetLimit ? (
            <span className="font-medium text-emerald-600">Fits your budget ✓</span>
          ) : (
            <span className="font-medium text-amber-600">Slightly over</span>
          )}
        </p>
      )}

      {/* Horizontal scroll rail */}
      <div className="no-scrollbar flex gap-3 overflow-x-auto pb-1">
        {collection.items.map((item) => (
          <SlotCard key={item.slot} product={item.product} slot={item.slot} onCartAdd={onCartAdd} />
        ))}
      </div>

      {/* Skipped slots */}
      {(collection.skippedSlots ?? []).length > 0 && (
        <p className="mt-3 text-xs text-ink-400">
          Not included (over budget): {collection.skippedSlots!.join(', ')}
        </p>
      )}

      {/* Why this collection */}
      {collection.explanation.length > 0 && (
        <div className="mt-4 rounded-lg bg-brand-50 p-3">
          <div className="flex items-center gap-2">
            <Sparkles className="h-3.5 w-3.5 shrink-0 text-brand-500" />
            <p className="text-xs font-bold text-ink-900">Why this collection?</p>
          </div>
          <ul className="mt-2 space-y-1.5">
            {collection.explanation.map((reason, i) => (
              <li key={i} className="flex items-start gap-2 text-xs text-ink-700">
                <CheckCircle2 className="mt-0.5 h-3 w-3 shrink-0 text-brand-500" />
                {reason}
              </li>
            ))}
          </ul>
        </div>
      )}
    </motion.div>
  )
}
