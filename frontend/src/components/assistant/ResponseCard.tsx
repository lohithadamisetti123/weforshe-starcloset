import { motion } from 'framer-motion'
import { IndianRupee, MapPin, Palette, Sparkles } from 'lucide-react'
import type { AIRecommendationResponse } from '@/types'
import { RecommendationCard } from '@/components/recommendations/RecommendationCard'

export function ResponseCard({ response }: { response: AIRecommendationResponse }) {
  const meta = [
    { icon: IndianRupee, label: 'Budget', value: response.budget },
    { icon: MapPin, label: 'Occasion', value: response.occasion },
    { icon: Palette, label: 'Style', value: response.style },
  ]

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      className="rounded-xl border border-line bg-white p-5 shadow-sm"
    >
      <p className="text-[15px] leading-relaxed text-ink-900">{response.summary}</p>

      <div className="mt-4 grid grid-cols-3 gap-2">
        {meta.map((m) => (
          <div key={m.label} className="rounded-lg border border-line bg-surface px-3 py-2.5 text-center">
            <m.icon className="mx-auto h-3.5 w-3.5 text-brand-500" />
            <p className="mt-1 text-[10px] font-medium uppercase tracking-wide text-ink-500">{m.label}</p>
            <p className="truncate text-xs font-bold text-ink-900">{m.value}</p>
          </div>
        ))}
      </div>

      <div className="mt-4 flex gap-2 rounded-lg bg-brand-50 p-3">
        <Sparkles className="h-4 w-4 shrink-0 text-brand-500" />
        <p className="text-xs leading-relaxed text-ink-700">{response.reason}</p>
      </div>

      <div className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-3">
        {response.products.map((product, i) => (
          <RecommendationCard key={product.id} product={product} index={i} />
        ))}
      </div>
    </motion.div>
  )
}
