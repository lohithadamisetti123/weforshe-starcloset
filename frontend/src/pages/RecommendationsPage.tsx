import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Clock3, IndianRupee, LayoutGrid, Sparkles, TrendingUp, Wand2 } from 'lucide-react'
import { apiFetch } from '@/lib/api'
import { formatINR } from '@/lib/utils'
import { DashboardCard } from '@/components/recommendations/DashboardCard'
import { CategoryChart } from '@/components/recommendations/CategoryChart'
import { BudgetChart } from '@/components/recommendations/BudgetChart'
import { RecommendationCard } from '@/components/recommendations/RecommendationCard'
import { CollectionCard } from '@/components/assistant/CollectionCard'
import { Badge } from '@/components/ui/Badge'
import { cn } from '@/lib/utils'
import { useForYou } from '@/hooks/useForYou'
import type { DashboardStats, HistoryEntry, Product } from '@/types'

const filterOptions = [
  'All', 'Wedding', 'Festival', 'Office', 'Party',
  'College', 'Casual', 'Travel', 'Gym', 'Date Night',
  'Kids Party', 'Self Care',
] as const

export function RecommendationsPage() {
  const [filter, setFilter] = useState<(typeof filterOptions)[number]>('All')
  const forYou = useForYou()

  const { data: stats } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: () => apiFetch<DashboardStats>('/dashboard/stats'),
  })

  const { data: recentHistory } = useQuery({
    queryKey: ['history-recent'],
    queryFn: () => apiFetch<HistoryEntry[]>('/history?pageSize=6'),
  })

  const { data: products } = useQuery({
    queryKey: ['products-search', filter],
    queryFn: () =>
      apiFetch<Product[]>(
        `/products/search${filter === 'All' ? '' : `?occasion=${encodeURIComponent(filter)}`}`
      ),
  })

  const filtered = useMemo(() => (products ?? []).slice(0, 12), [products])

  return (
    <div className="mx-auto max-w-7xl px-5 py-10 lg:px-8">
      <div className="flex flex-col gap-2">
        <p className="text-xs font-semibold uppercase tracking-widest text-brand-500">Dashboard</p>
        <h1 className="text-3xl font-bold tracking-tight text-ink-900 sm:text-4xl">Recommendation Dashboard</h1>
        <p className="max-w-xl text-ink-500">A live snapshot of how Bharat AI is styling shoppers right now.</p>
      </div>

      {/* ── Personalized "For You" feed, generated from the saved profile ── */}
      {forYou.loading ? (
        <div className="mt-8 h-56 animate-pulse rounded-2xl border border-line bg-white" />
      ) : forYou.hasProfile && (forYou.products.length > 0 || forYou.collections.length > 0) ? (
        <section className="mt-8">
          <div className="flex items-center gap-2">
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-brand-50 to-brand-100">
              <Wand2 className="h-4 w-4 text-brand-500" />
            </span>
            <h2 className="font-display text-xl font-bold text-ink-900">
              {forYou.profileName ? `Styled for ${forYou.profileName}` : 'Styled for you'}
            </h2>
            {forYou.personalized && (
              <span className="rounded-full bg-brand-50 px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide text-brand-600">
                Personalized
              </span>
            )}
          </div>
          <p className="mt-1 text-sm text-ink-500">
            Picked from your saved style, budget, favourites, and purchase history — no need to ask.
          </p>

          {forYou.products.length > 0 && (
            <div className="mt-5 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
              {forYou.products.slice(0, 8).map((product, i) => (
                <RecommendationCard key={product.id} product={product} index={i} />
              ))}
            </div>
          )}

          {forYou.collections.length > 0 && (
            <div className="mt-6 space-y-4">
              {forYou.collections.slice(0, 2).map((collection, i) => (
                <CollectionCard key={`${collection.title}-${i}`} collection={collection} />
              ))}
            </div>
          )}
        </section>
      ) : !forYou.hasProfile ? (
        <Link
          to="/profile"
          className="mt-8 flex items-center justify-between gap-4 rounded-2xl border border-dashed border-brand-300 bg-brand-50/40 px-5 py-4 transition-colors hover:bg-brand-50"
        >
          <div className="flex items-center gap-3">
            <span className="flex h-10 w-10 items-center justify-center rounded-full bg-white text-brand-500 shadow-sm">
              <Sparkles className="h-5 w-5" />
            </span>
            <div>
              <p className="text-sm font-semibold text-ink-900">Get personalized recommendations</p>
              <p className="text-sm text-ink-500">Set up your style preferences to unlock a feed made for you.</p>
            </div>
          </div>
          <span className="shrink-0 rounded-full bg-brand-500 px-4 py-2 text-sm font-semibold text-white">Set up</span>
        </Link>
      ) : null}
      
    </div>
  )
}