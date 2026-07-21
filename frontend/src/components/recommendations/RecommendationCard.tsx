import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Heart, ShoppingBag, Sparkles, Star, X } from 'lucide-react'
import type { Product } from '@/types'
import { formatINR, onImgError} from '@/lib/utils'
import { Badge } from '@/components/ui/Badge'
import { cn } from '@/lib/utils'
import { useShop } from '@/context/ShopContext'

interface RecommendationCardProps {
  product: Product
  index?: number
}

export function RecommendationCard({ product, index = 0 }: RecommendationCardProps) {
  const { isInCart, addToCart, removeFromCart, isFavorite, toggleFavorite } = useShop()
  const [showToast, setShowToast] = useState<string | null>(null)
  const [showDetail, setShowDetail] = useState(false)

  const inCart = isInCart(product.id)
  const saved = isFavorite(product.id)

  const handleCart = () => {
    if (inCart) {
      removeFromCart(product.id)
      setShowToast('Removed from cart')
    } else {
      addToCart(product)
      setShowToast('Added to cart!')
    }
    setTimeout(() => setShowToast(null), 2000)
  }

  const handleFavorite = () => {
    toggleFavorite(product.id)
    setShowToast(saved ? 'Removed from saved' : 'Saved!')
    setTimeout(() => setShowToast(null), 2000)
  }

  return (
    <>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, margin: '-40px' }}
        transition={{ duration: 0.4, delay: Math.min(index * 0.06, 0.4) }}
        whileHover={{ y: -6 }}
        className="group relative overflow-hidden rounded-xl border border-line bg-card shadow-[0_2px_16px_-4px_rgba(17,24,39,0.06)] transition-shadow hover:shadow-[0_20px_44px_-16px_rgba(17,24,39,0.22)]"
      >
        {/* Toast notification */}
        <AnimatePresence>
          {showToast && (
            <motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              className="absolute inset-x-0 top-0 z-20 flex items-center justify-center bg-ink-900/90 py-1.5 text-[11px] font-semibold text-white"
            >
              {showToast}
            </motion.div>
          )}
        </AnimatePresence>

        <div className="relative aspect-[3/4] overflow-hidden bg-ink-900/5">
          <img
            src={product.image}
            alt={product.title}
            loading="lazy"
            className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105" onError={onImgError} />

          {product.matchPercent && (
            <span className="absolute left-2.5 top-2.5 flex items-center gap-1 rounded-full bg-white/95 px-2.5 py-1 text-[10px] font-bold text-brand-600 shadow">
              <Sparkles className="h-3 w-3" />
              {product.matchPercent}% Match
            </span>
          )}

          {/* Heart — wired to real favorites API */}
          <button
            onClick={handleFavorite}
            aria-label={saved ? 'Remove from saved' : 'Save product'}
            aria-pressed={saved}
            className="absolute right-2.5 top-2.5 flex h-8 w-8 items-center justify-center rounded-full bg-white/95 shadow transition-transform hover:scale-110"
          >
            <Heart
              className={cn(
                'h-4 w-4 transition-colors',
                saved ? 'fill-brand-500 text-brand-500' : 'text-ink-500',
              )}
            />
          </button>

          {product.discount > 0 && (
            <span className="absolute bottom-2.5 left-2.5 rounded-md bg-ink-900 px-2 py-1 text-[10px] font-bold text-white">
              {product.discount}% OFF
            </span>
          )}
        </div>

        <div className="p-4">
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0">
              <p className="truncate text-sm font-bold text-ink-900">{product.brand}</p>
              <p className="truncate text-xs text-ink-500">{product.title}</p>
            </div>
            <Badge variant="outline" size="sm" className="shrink-0">
              {product.category}
            </Badge>
          </div>

          <div className="mt-2.5 flex items-center gap-1.5">
            <div className="flex items-center gap-0.5 rounded bg-emerald-50 px-1.5 py-0.5 text-[11px] font-bold text-emerald-700">
              {product.rating.toFixed(1)}
              <Star className="h-2.5 w-2.5 fill-emerald-700" />
            </div>
            <span className="text-[11px] text-ink-500">({product.ratingCount})</span>
          </div>

          <div className="mt-2.5 flex items-baseline gap-2">
            <span className="text-base font-bold text-ink-900">{formatINR(product.price)}</span>
            {product.mrp > product.price && (
              <span className="text-xs text-ink-300 line-through">{formatINR(product.mrp)}</span>
            )}
          </div>

          <div className="mt-4 flex gap-2">
            {/* Add to Cart — localStorage backed */}
            <button
              onClick={handleCart}
              className={cn(
                'flex flex-1 items-center justify-center gap-1.5 rounded-full py-2 text-xs font-semibold transition-colors',
                inCart
                  ? 'bg-emerald-50 text-emerald-700'
                  : 'bg-ink-900 text-white hover:bg-brand-600',
              )}
            >
              <ShoppingBag className="h-3.5 w-3.5" />
              {inCart ? 'Remove from Cart' : 'Add to Cart'}
            </button>

            {/* View Details */}
            <button
              onClick={() => setShowDetail(true)}
              className="flex-1 rounded-full border border-line py-2 text-xs font-semibold text-ink-700 transition-colors hover:border-ink-900"
            >
              View Details
            </button>
          </div>
        </div>
      </motion.div>

      {/* Product detail modal */}
      <AnimatePresence>
        {showDetail && (
          <ProductDetailModal product={product} onClose={() => setShowDetail(false)} />
        )}
      </AnimatePresence>
    </>
  )
}

// Inline detail modal — keeps the component self-contained without a new file
function ProductDetailModal({ product, onClose }: { product: Product; onClose: () => void }) {
  const { isInCart, addToCart, removeFromCart, isFavorite, toggleFavorite } = useShop()
  const inCart = isInCart(product.id)
  const saved = isFavorite(product.id)

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 z-50 flex items-end justify-center bg-ink-900/50 sm:items-center"
      onClick={onClose}
    >
      <motion.div
        initial={{ y: 60, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        exit={{ y: 60, opacity: 0 }}
        transition={{ type: 'spring', damping: 26 }}
        onClick={(e) => e.stopPropagation()}
        className="relative w-full max-w-sm overflow-hidden rounded-t-2xl bg-white sm:rounded-2xl"
      >
        {/* Close */}
        <button
          onClick={onClose}
          className="absolute right-3 top-3 z-10 flex h-8 w-8 items-center justify-center rounded-full bg-white/90 shadow"
        >
          <X className="h-4 w-4 text-ink-700" />
        </button>

        <div className="relative aspect-[4/5] overflow-hidden">
          <img src={product.image} alt={product.title} className="h-full w-full object-cover" onError={onImgError} />
          {product.discount > 0 && (
            <span className="absolute left-3 top-3 rounded-md bg-ink-900 px-2 py-1 text-[11px] font-bold text-white">
              {product.discount}% OFF
            </span>
          )}
        </div>

        <div className="p-5">
          <div className="flex items-start justify-between gap-2">
            <div>
              <p className="text-sm font-bold text-ink-900">{product.brand}</p>
              <p className="mt-0.5 text-xs text-ink-500">{product.title}</p>
            </div>
            <button onClick={() => toggleFavorite(product.id)}>
              <Heart
                className={cn(
                  'h-5 w-5 transition-colors',
                  saved ? 'fill-brand-500 text-brand-500' : 'text-ink-400',
                )}
              />
            </button>
          </div>

          <div className="mt-3 flex items-center gap-2">
            <span className="text-xl font-bold text-ink-900">{formatINR(product.price)}</span>
            {product.mrp > product.price && (
              <span className="text-sm text-ink-300 line-through">{formatINR(product.mrp)}</span>
            )}
            <span className="rounded bg-emerald-50 px-1.5 py-0.5 text-[11px] font-bold text-emerald-700">
              {product.rating.toFixed(1)} ★
            </span>
          </div>

          {product.description && (
            <p className="mt-3 line-clamp-3 text-xs leading-relaxed text-ink-500">
              {product.description}
            </p>
          )}

          <div className="mt-4 grid grid-cols-2 gap-2 text-[11px] text-ink-600">
            <span className="rounded-lg bg-surface px-3 py-2">🎨 {product.color}</span>
            <span className="rounded-lg bg-surface px-3 py-2">👗 {product.occasion}</span>
            <span className="rounded-lg bg-surface px-3 py-2">✨ {product.style}</span>
            <span className="rounded-lg bg-surface px-3 py-2">📦 {product.category}</span>
          </div>

          <div className="mt-4 flex gap-2">
            <button
              onClick={() => { inCart ? removeFromCart(product.id) : addToCart(product); onClose() }}
              className={cn(
                'flex flex-1 items-center justify-center gap-2 rounded-full py-3 text-sm font-semibold transition-colors',
                inCart
                  ? 'bg-emerald-50 text-emerald-700'
                  : 'bg-ink-900 text-white hover:bg-brand-600',
              )}
            >
              <ShoppingBag className="h-4 w-4" />
              {inCart ? 'In Cart ✓' : 'Add to Cart'}
            </button>
            <button
              onClick={() => toggleFavorite(product.id)}
              className={cn(
                'flex h-12 w-12 shrink-0 items-center justify-center rounded-full border transition-colors',
                saved
                  ? 'border-brand-500 bg-brand-50 text-brand-500'
                  : 'border-line text-ink-500 hover:border-brand-500',
              )}
            >
              <Heart className={cn('h-5 w-5', saved && 'fill-brand-500')} />
            </button>
          </div>
        </div>
      </motion.div>
    </motion.div>
  )
}
