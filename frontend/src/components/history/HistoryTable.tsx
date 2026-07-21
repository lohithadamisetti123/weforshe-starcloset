import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronLeft, ChevronRight, Search } from 'lucide-react'
import { apiFetch } from '@/lib/api'
import { formatDate } from '@/lib/utils'
import { Badge } from '@/components/ui/Badge'
import { cn } from '@/lib/utils'
import type { HistoryEntry } from '@/types'

const PAGE_SIZE = 6
const statuses: Array<HistoryEntry['status'] | 'All'> = ['All', 'Completed', 'Saved', 'Purchased']

const statusVariant: Record<HistoryEntry['status'], 'success' | 'brand' | 'outline'> = {
  Purchased: 'success',
  Completed: 'brand',
  Saved: 'outline',
}

export function HistoryTable() {
  const [query, setQuery] = useState('')
  const [status, setStatus] = useState<(typeof statuses)[number]>('All')
  const [page, setPage] = useState(1)

  // pageSize=200 pulls the full history in one call — the backend paginates,
  // but at this demo's scale that's simpler than wiring server-side search
  // and status filtering through as separate query params, and this keeps
  // every bit of the existing client-side filter/paginate logic below
  // untouched.
  const { data: historyEntries } = useQuery({
    queryKey: ['history-all'],
    queryFn: () => apiFetch<HistoryEntry[]>('/history?pageSize=200'),
  })

  const filtered = useMemo(() => {
    return (historyEntries ?? []).filter((entry) => {
      const matchesQuery = entry.prompt.toLowerCase().includes(query.toLowerCase())
      const matchesStatus = status === 'All' || entry.status === status
      return matchesQuery && matchesStatus
    })
  }, [historyEntries, query, status])

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE))
  const paginated = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  const handleFilterChange = (fn: () => void) => {
    fn()
    setPage(1)
  }

  return (
    <div>
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative flex-1 sm:max-w-xs">
          <Search className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-ink-300" />
          <input
            value={query}
            onChange={(e) => handleFilterChange(() => setQuery(e.target.value))}
            placeholder="Search prompts..."
            className="w-full rounded-full border border-line bg-white py-2.5 pl-10 pr-4 text-sm text-ink-900 placeholder:text-ink-300 focus:border-brand-500 focus:outline-none"
          />
        </div>
        <div className="no-scrollbar flex gap-2 overflow-x-auto">
          {statuses.map((s) => (
            <button
              key={s}
              onClick={() => handleFilterChange(() => setStatus(s))}
              className={cn(
                'shrink-0 rounded-full border px-4 py-2 text-xs font-semibold transition-colors',
                status === s ? 'border-ink-900 bg-ink-900 text-white' : 'border-line bg-white text-ink-700'
              )}
            >
              {s}
            </button>
          ))}
        </div>
      </div>

      <div className="mt-5 overflow-hidden rounded-xl border border-line bg-white">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-line bg-surface text-[11px] font-semibold uppercase tracking-wide text-ink-500">
                <th className="px-5 py-3.5">Search Prompt</th>
                <th className="px-5 py-3.5">Date</th>
                <th className="px-5 py-3.5">Budget</th>
                <th className="px-5 py-3.5">Occasion</th>
                <th className="px-5 py-3.5">Products</th>
                <th className="px-5 py-3.5">Status</th>
              </tr>
            </thead>
            <tbody>
              {paginated.map((entry) => (
                <tr key={entry.id} className="border-b border-line last:border-0 hover:bg-surface">
                  <td className="max-w-xs px-5 py-4 font-medium text-ink-900">{entry.prompt}</td>
                  <td className="px-5 py-4 text-ink-500">{formatDate(entry.date)}</td>
                  <td className="px-5 py-4 text-ink-700">{entry.budget}</td>
                  <td className="px-5 py-4 text-ink-700">{entry.occasion}</td>
                  <td className="px-5 py-4 text-ink-700">{entry.productsRecommended}</td>
                  <td className="px-5 py-4">
                    <Badge variant={statusVariant[entry.status]} size="sm">
                      {entry.status}
                    </Badge>
                  </td>
                </tr>
              ))}
              {paginated.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-5 py-14 text-center text-sm text-ink-500">
                    No results match your search.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <div className="flex items-center justify-between border-t border-line px-5 py-3.5">
          <p className="text-xs text-ink-500">
            Page {page} of {totalPages} · {filtered.length} results
          </p>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="flex h-8 w-8 items-center justify-center rounded-full border border-line text-ink-700 disabled:opacity-30"
              aria-label="Previous page"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="flex h-8 w-8 items-center justify-center rounded-full border border-line text-ink-700 disabled:opacity-30"
              aria-label="Next page"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}