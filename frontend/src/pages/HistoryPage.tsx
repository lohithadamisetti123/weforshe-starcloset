import { HistoryTable } from '@/components/history/HistoryTable'

export function HistoryPage() {
  return (
    <div className="mx-auto max-w-7xl px-5 py-10 lg:px-8">
      <div className="flex flex-col gap-2">
        <p className="text-xs font-semibold uppercase tracking-widest text-brand-500">Your activity</p>
        <h1 className="text-3xl font-bold tracking-tight text-ink-900 sm:text-4xl">Chat History</h1>
        <p className="max-w-xl text-ink-500">Every prompt you've sent Bharat AI, and what it recommended.</p>
      </div>

      <div className="mt-8">
        <HistoryTable />
      </div>
    </div>
  )
}
