import { motion } from 'framer-motion'
import { Heart, Sparkles, TrendingUp } from 'lucide-react'
import { products } from '@/data/products'
import { formatINR, onImgError} from '@/lib/utils'

const featured = [products[0], products[9], products[19], products[35]]

export function FloatingCards() {
  return (
    <div className="relative mx-auto h-[420px] w-full max-w-md sm:h-[520px]">
      {/* Ambient glow */}
      <div className="absolute left-1/2 top-1/2 h-[420px] w-[420px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-gradient-to-br from-brand-500/30 via-brand-300/20 to-transparent blur-3xl animate-glow" />

      {/* AI particles */}
      {Array.from({ length: 10 }).map((_, i) => (
        <motion.span
          key={i}
          className="absolute h-1.5 w-1.5 rounded-full bg-brand-500/70"
          style={{
            top: `${(i * 37) % 100}%`,
            left: `${(i * 53) % 100}%`,
          }}
          animate={{ opacity: [0.2, 0.9, 0.2], scale: [1, 1.6, 1] }}
          transition={{ duration: 3 + (i % 4), repeat: Infinity, delay: i * 0.3 }}
        />
      ))}

      {/* Main product card */}
      <motion.div
        className="absolute left-1/2 top-1/2 w-52 -translate-x-1/2 -translate-y-1/2 rounded-xl border border-line bg-card p-3 shadow-[0_24px_60px_-16px_rgba(17,24,39,0.28)] sm:w-60"
        animate={{ y: [0, -14, 0] }}
        transition={{ duration: 5, repeat: Infinity, ease: 'easeInOut' }}
      >
        <div className="relative overflow-hidden rounded-lg">
          <img src={featured[0].image} alt={featured[0].title} className="h-56 w-full object-cover sm:h-64" onError={onImgError} />
          <span className="absolute right-2 top-2 flex items-center gap-1 rounded-full bg-white/90 px-2 py-1 text-[10px] font-bold text-brand-600 shadow">
            <Sparkles className="h-3 w-3" /> {featured[0].matchPercent ?? 96}% Match
          </span>
        </div>
        <p className="mt-2.5 truncate text-xs font-semibold text-ink-900">{featured[0].brand}</p>
        <p className="truncate text-[11px] text-ink-500">{featured[0].title}</p>
        <p className="mt-1 text-sm font-bold text-ink-900">{formatINR(featured[0].price)}</p>
      </motion.div>

      {/* Secondary card top-left */}
      <motion.div
        className="glass absolute left-0 top-2 w-32 rounded-lg border border-white/50 p-2 shadow-xl sm:w-36"
        animate={{ y: [0, 12, 0], rotate: [-6, -3, -6] }}
        transition={{ duration: 6, repeat: Infinity, ease: 'easeInOut' }}
      >
        <img src={featured[1].image} alt={featured[1].title} className="h-28 w-full rounded-md object-cover sm:h-32" onError={onImgError} />
        <p className="mt-1.5 truncate text-[10px] font-semibold text-ink-900">{featured[1].brand}</p>
        <p className="text-[10px] font-bold text-brand-600">{formatINR(featured[1].price)}</p>
      </motion.div>

      {/* Secondary card bottom-right */}
      <motion.div
        className="glass absolute bottom-4 right-0 w-32 rounded-lg border border-white/50 p-2 shadow-xl sm:w-36"
        animate={{ y: [0, -10, 0], rotate: [6, 3, 6] }}
        transition={{ duration: 5.5, repeat: Infinity, ease: 'easeInOut', delay: 0.5 }}
      >
        <img src={featured[2].image} alt={featured[2].title} className="h-28 w-full rounded-md object-cover sm:h-32" onError={onImgError} />
        <p className="mt-1.5 truncate text-[10px] font-semibold text-ink-900">{featured[2].brand}</p>
        <p className="text-[10px] font-bold text-brand-600">{formatINR(featured[2].price)}</p>
      </motion.div>

      {/* Floating stat chip */}
      <motion.div
        className="glass absolute right-2 top-16 flex items-center gap-2 rounded-full border border-white/50 px-3 py-2 shadow-lg sm:right-4"
        animate={{ y: [0, -8, 0] }}
        transition={{ duration: 4, repeat: Infinity, ease: 'easeInOut', delay: 1 }}
      >
        <TrendingUp className="h-3.5 w-3.5 text-brand-500" />
        <span className="text-[11px] font-semibold text-ink-900">98% Style Match</span>
      </motion.div>

      <motion.div
        className="glass absolute bottom-16 left-2 flex items-center gap-2 rounded-full border border-white/50 px-3 py-2 shadow-lg"
        animate={{ y: [0, 10, 0] }}
        transition={{ duration: 4.5, repeat: Infinity, ease: 'easeInOut', delay: 0.7 }}
      >
        <Heart className="h-3.5 w-3.5 fill-brand-500 text-brand-500" />
        <span className="text-[11px] font-semibold text-ink-900">Curated for you</span>
      </motion.div>
    </div>
  )
}
