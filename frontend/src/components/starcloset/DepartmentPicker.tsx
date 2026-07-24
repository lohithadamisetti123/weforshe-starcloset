import { motion } from 'framer-motion'
import { ArrowRight } from 'lucide-react'
import { useStarCloset } from '@/context/StarClosetContext'
import { Skeleton } from '@/components/ui/Skeleton'

// DepartmentPicker is StarCloset V2's first step — "Who are you shopping for?".
// Choosing a department opens its own independent stylist journey.
const DEPT_STYLE: Record<string, { from: string; to: string; blurb: string }> = {
  Women: { from: '#FF7EB3', to: '#FF6C9C', blurb: 'Ethnic, western & everything in between' },
  Men: { from: '#2C3E50', to: '#4CA1AF', blurb: 'Sharp formals to easy streetwear' },
  Kids: { from: '#F6D365', to: '#FDA085', blurb: 'Playful, comfy & celebration-ready' },
}

export function DepartmentPicker({ afterSelect }: { afterSelect?: () => void } = {}) {
  const { departments, loading, selectDepartment } = useStarCloset()

  const choose = (key: string) => {
    selectDepartment(key)
    afterSelect?.()
  }

  return (
    <div className="mx-auto max-w-4xl px-5 py-10 lg:px-8">
      <div className="text-center">
        <h1 className="font-display text-2xl font-bold text-ink-900 sm:text-3xl">
          Who are you shopping for?
        </h1>
        <p className="mt-2 text-sm text-ink-500">
          Each department has its own celebrity stylists — and its own unlock journey.
        </p>
      </div>

      <div className="mt-10 grid grid-cols-1 gap-5 sm:grid-cols-3">
        {loading
          ? Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-64 rounded-3xl" />)
          : departments.map((dept, i) => {
              const style = DEPT_STYLE[dept.key] ?? { from: '#FF3F6C', to: '#1F1C18', blurb: '' }
              return (
                <motion.button
                  key={dept.key}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.08 }}
                  whileHover={{ y: -6 }}
                  onClick={() => choose(dept.key)}
                  className="group relative flex h-64 flex-col justify-between overflow-hidden rounded-3xl border border-line p-6 text-left shadow-sm transition-shadow hover:shadow-xl"
                >
                  <div
                    className="pointer-events-none absolute inset-0 opacity-90 transition-opacity group-hover:opacity-100"
                    style={{ backgroundImage: `linear-gradient(150deg, ${style.from}, ${style.to})` }}
                  />
                  <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(rgba(255,255,255,0.25)_1px,transparent_1px)] [background-size:20px_20px] opacity-30" />

                  <div className="relative">
                    <span className="text-5xl drop-shadow-sm">{dept.emoji}</span>
                  </div>
                  <div className="relative text-white">
                    <p className="font-display text-2xl font-bold">{dept.label}</p>
                    <p className="mt-1 text-sm text-white/85">{style.blurb}</p>
                    <span className="mt-4 inline-flex items-center gap-1.5 rounded-full bg-white/20 px-3 py-1.5 text-xs font-semibold backdrop-blur-sm">
                      {dept.stylists.length} stylists <ArrowRight className="h-3.5 w-3.5" />
                    </span>
                  </div>
                </motion.button>
              )
            })}
      </div>
    </div>
  )
}
