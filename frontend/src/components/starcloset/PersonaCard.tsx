import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Lock, Check, Sparkles } from 'lucide-react'
import { usePersona } from '@/context/PersonaContext'
import { StylistAvatar } from '@/components/starcloset/StylistAvatar'
import { cn } from '@/lib/utils'
import type { StylistPersona } from '@/types/persona'

interface PersonaCardProps {
  persona: StylistPersona
  // onSelected fires after a stylist is chosen (e.g. to close a swap panel).
  // When omitted, selecting an unlocked stylist navigates to the studio.
  onSelected?: () => void
}

export function PersonaCard({ persona, onSelected }: PersonaCardProps) {
  const { isUnlocked, itemsToUnlock, selectPersona, selectedSlug } = usePersona()
  const navigate = useNavigate()

  const unlocked = isUnlocked(persona)
  const remaining = itemsToUnlock(persona)
  const active = selectedSlug === persona.slug

  const handleSelect = () => {
    if (!unlocked) return
    selectPersona(persona.slug)
    if (onSelected) onSelected()
    else navigate('/assistant')
  }

  return (
    <motion.div
      whileHover={unlocked ? { y: -4 } : undefined}
      className={cn(
        'group relative flex flex-col overflow-hidden rounded-3xl border bg-white p-5 text-left transition-shadow',
        active ? 'border-brand-500 shadow-[0_12px_32px_-12px_rgba(255,63,108,0.45)]' : 'border-line',
        unlocked ? 'hover:shadow-lg' : 'opacity-95',
      )}
    >
      {/* Accent wash header */}
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-24 opacity-15"
        style={{ backgroundImage: `linear-gradient(135deg, ${persona.accentFrom}, ${persona.accentTo})` }}
      />

      <div className="relative flex items-start justify-between">
        <StylistAvatar persona={persona} size="lg" locked={!unlocked} />
        {active ? (
          <span className="flex items-center gap-1 rounded-full bg-brand-500 px-2.5 py-1 text-[11px] font-bold text-white">
            <Check className="h-3 w-3" /> Active
          </span>
        ) : unlocked ? (
          <span className="rounded-full bg-emerald-50 px-2.5 py-1 text-[11px] font-bold text-emerald-600">
            Unlocked
          </span>
        ) : (
          <span className="flex items-center gap-1 rounded-full bg-ink-100 px-2.5 py-1 text-[11px] font-bold text-ink-500">
            <Lock className="h-3 w-3" /> Locked
          </span>
        )}
      </div>

      <h3 className="relative mt-4 font-display text-lg font-bold text-ink-900">{persona.displayName}</h3>
      <p className="relative mt-1 line-clamp-2 text-sm text-ink-500">{persona.tagline}</p>

      <div className="relative mt-4 flex-1" />

      {unlocked ? (
        <button
          onClick={handleSelect}
          className={cn(
            'relative mt-2 flex items-center justify-center gap-1.5 rounded-full px-4 py-2.5 text-sm font-semibold transition-colors',
            active
              ? 'bg-ink-900 text-white hover:bg-brand-600'
              : 'bg-brand-500 text-white hover:bg-brand-600',
          )}
        >
          <Sparkles className="h-4 w-4" />
          {active ? 'Continue styling' : `Style with ${persona.displayName.split(' ')[0]}`}
        </button>
      ) : (
        <div className="relative mt-2 rounded-full border border-dashed border-line px-4 py-2.5 text-center text-xs font-semibold text-ink-500">
          Add {remaining} more item{remaining === 1 ? '' : 's'} to unlock
        </div>
      )}
    </motion.div>
  )
}
