import type { LucideIcon } from 'lucide-react'
import { motion } from 'framer-motion'
import { cn } from '@/lib/utils'

interface DashboardCardProps {
  icon: LucideIcon
  label: string
  value: string
  trend?: string
  accent?: boolean
}

export function DashboardCard({ icon: Icon, label, value, trend, accent }: DashboardCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.4 }}
      className={cn(
        'rounded-xl border p-5 shadow-sm',
        accent ? 'border-transparent bg-gradient-to-br from-ink-900 to-brand-700 text-white' : 'border-line bg-card'
      )}
    >
      <div className="flex items-center justify-between">
        <span
          className={cn(
            'flex h-10 w-10 items-center justify-center rounded-lg',
            accent ? 'bg-white/15' : 'bg-brand-50 text-brand-600'
          )}
        >
          <Icon className="h-5 w-5" />
        </span>
        {trend && (
          <span className={cn('text-[11px] font-semibold', accent ? 'text-white/80' : 'text-emerald-600')}>
            {trend}
          </span>
        )}
      </div>
      <p className={cn('mt-4 text-2xl font-bold tracking-tight', accent ? 'text-white' : 'text-ink-900')}>{value}</p>
      <p className={cn('mt-1 text-xs font-medium', accent ? 'text-white/70' : 'text-ink-500')}>{label}</p>
    </motion.div>
  )
}
