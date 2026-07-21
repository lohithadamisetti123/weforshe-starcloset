import { motion } from 'framer-motion'
import {
  Gem,
  ShoppingBag,
  Shirt,
  Sparkles,
  User,
  Users,
  Watch,
} from 'lucide-react'
import type { DepartmentOption } from '@/types'

interface DepartmentSelectorProps {
  departments: DepartmentOption[]
  onSelect: (name: string) => void
}

const DEPT_ICON: Record<string, React.ElementType> = {
  Women: Users,
  Men: User,
  'Kids Girls': Sparkles,
  'Kids Boys': Sparkles,
  Beauty: Gem,
  Footwear: ShoppingBag,
  Accessories: Watch,
  // Legacy fallback
  'Kids - Girls': Sparkles,
  'Kids - Boys': Sparkles,
}

const DEPT_GRADIENT: Record<string, string> = {
  Women: 'from-rose-400 to-pink-600',
  Men: 'from-blue-500 to-indigo-700',
  'Kids Girls': 'from-purple-400 to-fuchsia-600',
  'Kids Boys': 'from-sky-400 to-cyan-600',
  Beauty: 'from-pink-400 to-rose-600',
  Footwear: 'from-emerald-400 to-teal-600',
  Accessories: 'from-violet-400 to-purple-600',
  // Legacy fallback
  'Kids - Girls': 'from-purple-400 to-fuchsia-600',
  'Kids - Boys': 'from-sky-400 to-cyan-600',
}

export function DepartmentSelector({ departments, onSelect }: DepartmentSelectorProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      className="flex h-full flex-col items-center justify-center py-8 text-center"
    >
      <div className="flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-brand-50 to-brand-100">
        <Sparkles className="h-7 w-7 text-brand-500" />
      </div>
      <h2 className="mt-5 font-display text-xl font-bold text-ink-900">
        Hi! I'm your AI Shopping Assistant.
      </h2>
      <p className="mt-2 max-w-sm text-sm text-ink-500">
        What would you like to shop for today?
      </p>

      <div className="mt-8 grid w-full max-w-xl grid-cols-2 gap-3 sm:grid-cols-4">
        {departments.map((dept, i) => {
          const Icon = DEPT_ICON[dept.name] ?? Shirt
          const gradient = DEPT_GRADIENT[dept.name] ?? 'from-brand-400 to-brand-600'
          return (
            <motion.button
              key={dept.name}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3, delay: i * 0.05 }}
              onClick={() => dept.available && onSelect(dept.name)}
              disabled={!dept.available}
              className={[
                'relative flex flex-col items-center gap-3 rounded-xl border p-4 text-center transition-all duration-200',
                dept.available
                  ? 'cursor-pointer border-line bg-white hover:-translate-y-0.5 hover:border-brand-300 hover:shadow-md'
                  : 'cursor-not-allowed border-line bg-surface opacity-50',
              ].join(' ')}
            >
              <span
                className={`flex h-11 w-11 items-center justify-center rounded-full bg-gradient-to-br ${gradient}`}
              >
                <Icon className="h-5 w-5 text-white" />
              </span>
              <span className="text-xs font-semibold leading-tight text-ink-900">
                {dept.name}
              </span>
              {!dept.available && (
                <span className="absolute -top-1.5 right-2 rounded-full bg-ink-900/10 px-2 py-0.5 text-[9px] font-medium uppercase tracking-wide text-ink-500">
                  Soon
                </span>
              )}
            </motion.button>
          )
        })}
      </div>
    </motion.div>
  )
}
