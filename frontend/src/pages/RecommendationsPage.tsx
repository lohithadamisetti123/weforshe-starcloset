import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { Clock3, IndianRupee, LayoutGrid, Sparkles, TrendingUp } from 'lucide-react'
import { apiFetch } from '@/lib/api'
import { formatINR } from '@/lib/utils'
import { DashboardCard } from '@/components/recommendations/DashboardCard'
import { CategoryChart } from '@/components/recommendations/CategoryChart'
import { BudgetChart } from '@/components/recommendations/BudgetChart'
import { RecommendationCard } from '@/components/recommendations/RecommendationCard'
import { Badge } from '@/components/ui/Badge'
import { cn } from '@/lib/utils'
import type { DashboardStats, HistoryEntry, Product } from '@/types'

const filterOptions = [
  'All', 'Wedding', 'Festival', 'Office', 'Party',
  'College', 'Casual', 'Travel', 'Gym', 'Date Night',
  'Kids Party', 'Self Care',
] as const

export function RecommendationsPage() {
  const [filter, setFilter] = useState<(typeof filterOptions)[number]>('All')

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

      <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <DashboardCard icon={Sparkles} label="Recommendations Today" value={String(stats?.recommendationsToday ?? '—')} trend="+18% vs yesterday" accent />
        <DashboardCard icon={IndianRupee} label="Average Budget" value={stats ? formatINR(stats.averageBudget) : '—'} trend="Across all products" />
        <DashboardCard icon={LayoutGrid} label="Most Selected Category" value={stats?.mostSelectedCategory ?? '—'} trend="Top pick this week" />
        <DashboardCard icon={TrendingUp} label="Average AI Match" value={stats ? `${stats.averageAIMatch}%` : '—'} trend="High confidence" />
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
        <CategoryChart products={products ?? []} />
        <BudgetChart products={products ?? []} />
      </div>

      <div className="mt-10 flex items-center gap-2">
        <Clock3 className="h-4 w-4 text-brand-500" />
        <p className="font-display text-base font-semibold text-ink-900">Recent Searches</p>
      </div>
      <div className="mt-4 flex flex-wrap gap-2">
        {(recentHistory ?? []).map((entry) => (
          <Badge key={entry.id} variant="outline" size="md">
            {entry.prompt}
          </Badge>
        ))}
      </div>

      <div className="mt-10 flex items-center justify-between">
        <p className="font-display text-lg font-semibold text-ink-900">Recommendation History</p>
      </div>
      <div className="no-scrollbar mt-4 flex gap-2 overflow-x-auto pb-2">
        {filterOptions.map((option) => (
          <button
            key={option}
            onClick={() => setFilter(option)}
            className={cn(
              'shrink-0 rounded-full border px-4 py-2 text-sm font-medium transition-colors',
              filter === option
                ? 'border-brand-500 bg-brand-500 text-white'
                : 'border-line bg-white text-ink-700 hover:border-brand-500 hover:text-brand-600'
            )}
          >
            {option}
          </button>
        ))}
      </div>

      {filtered.length === 0 ? (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="mt-10 flex flex-col items-center justify-center rounded-xl border border-dashed border-line py-20 text-center"
        >
          <Sparkles className="h-8 w-8 text-ink-300" />
          <p className="mt-4 text-sm font-medium text-ink-500">No recommendations yet for this occasion.</p>
        </motion.div>
      ) : (
        <div className="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
          {filtered.map((product, i) => (
            <RecommendationCard key={product.id} product={product} index={i} />
          ))}
        </div>
      )}
    </div>
  )
}