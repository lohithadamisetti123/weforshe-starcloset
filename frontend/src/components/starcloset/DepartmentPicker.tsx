import { useState } from 'react'
import { motion } from 'framer-motion'
import { ArrowRight } from 'lucide-react'
import { useStarCloset } from '@/context/StarClosetContext'
import { Skeleton } from '@/components/ui/Skeleton'
import type { StarDepartment } from '@/types/starcloset'

// DepartmentPicker is StarCloset V2's first step — "Who are you shopping for?".
// Choosing a department opens its own independent stylist journey. Each card is
// fronted by a premium department image; the gradient behind it is both the
// brand accent and the graceful fallback when the image is missing.
const DEPT_STYLE: Record<string, { from: string; to: string; blurb: string; img: string }> = {
  Women: { from: '#FF7EB3', to: '#C64B7B', blurb: 'Ethnic, western & everything in between', img: '/images/departments/women.png' },
  Men: { from: '#2C3E50', to: '#1B2A38', blurb: 'Sharp formals to easy streetwear', img: '/images/departments/men.png' },
  Kids: { from: '#F6A365', to: '#E4713F', blurb: 'Playful, comfy & celebration-ready', img: '/images/departments/kids.png' },
}

function DeptCard({ dept, index, onChoose }: { dept: StarDepartment; index: number; onChoose: () => void }) {
  const style = DEPT_STYLE[dept.key] ?? { from: '#FF3F6C', to: '#1F1C18', blurb: '', img: '' }
  const [imgOk, setImgOk] = useState(true)

  return (
    <motion.button
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.08 }}
      whileHover={{ y: -6 }}
      onClick={onChoose}
      className="group relative flex aspect-[4/5] w-full flex-col justify-end overflow-hidden rounded-3xl border border-line text-left shadow-sm transition-shadow hover:shadow-xl sm:aspect-[3/4]"
    >
      {/* Gradient base — also the fallback when the image is missing */}
      <div
        className="pointer-events-none absolute inset-0"
        style={{ backgroundImage: `linear-gradient(150deg, ${style.from}, ${style.to})` }}
      />
      {/* Department image, full-bleed */}
      {imgOk && style.img && (
        <img
          src={style.img}
          alt={`${dept.label} department`}
          loading="lazy"
          decoding="async"
          onError={() => setImgOk(false)}
          className="absolute inset-0 h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
        />
      )}
      {/* Legibility scrim so the white text reads over any image */}
      <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent" />

      {/* Content */}
      <div className="relative z-10 p-6 text-white">
        <span className="mb-2 block text-4xl drop-shadow-md">{dept.emoji}</span>
        <p className="font-display text-2xl font-bold drop-shadow-sm">{dept.label}</p>
        <p className="mt-1 text-sm text-white/90 drop-shadow-sm">{style.blurb}</p>
        <span className="mt-4 inline-flex items-center gap-1.5 rounded-full bg-white/25 px-3 py-1.5 text-xs font-semibold backdrop-blur-sm">
          {dept.stylists.length} stylists <ArrowRight className="h-3.5 w-3.5" />
        </span>
      </div>
    </motion.button>
  )
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
          ? Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="aspect-[3/4] rounded-3xl" />)
          : departments.map((dept, i) => (
              <DeptCard key={dept.key} dept={dept} index={i} onChoose={() => choose(dept.key)} />
            ))}
      </div>
    </div>
  )
}
