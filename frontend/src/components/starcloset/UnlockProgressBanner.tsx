import { motion } from 'framer-motion'
import { Lock, Sparkles } from 'lucide-react'
import { usePersona } from '@/context/PersonaContext'
import { StylistAvatar } from '@/components/starcloset/StylistAvatar'

// UnlockProgressBanner nudges the shopper toward the next locked stylist:
// "Add N more items to unlock <stylist>", with a progress bar from the previous
// threshold to the next. It renders nothing once every stylist is unlocked — the
// loop is complete, so there is nothing to chase.
export function UnlockProgressBanner() {
  const { nextLocked, unlockScore, itemsToUnlock, unlockedPersonas } = usePersona()

  if (!nextLocked) {
    if (unlockedPersonas.length === 0) return null
    return (
      <div className="flex items-center gap-2 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-2.5 text-sm font-semibold text-emerald-700">
        <Sparkles className="h-4 w-4" />
        Every stylist unlocked — your full StarCloset is open.
      </div>
    )
  }

  const remaining = itemsToUnlock(nextLocked)
  // Progress from the previous unlock threshold to this one.
  const prevThreshold = Math.max(
    0,
    ...unlockedPersonas.map((p) => p.unlockAtItems),
  )
  const span = Math.max(1, nextLocked.unlockAtItems - prevThreshold)
  const done = Math.min(span, unlockScore - prevThreshold)
  const pct = Math.max(6, Math.round((done / span) * 100))

  return (
    <div className="rounded-2xl border border-line bg-white px-4 py-3">
      <div className="flex items-center gap-3">
        <StylistAvatar persona={nextLocked} size="sm" locked />
        <div className="flex-1">
          <p className="text-sm font-semibold text-ink-900">
            <Lock className="mr-1 inline h-3 w-3 text-ink-400" />
            Add {remaining} more item{remaining === 1 ? '' : 's'} to unlock{' '}
            <span className="text-brand-600">{nextLocked.displayName}</span>
          </p>
          <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-ink-100">
            <motion.div
              className="h-full rounded-full"
              style={{ backgroundImage: `linear-gradient(90deg, ${nextLocked.accentFrom}, ${nextLocked.accentTo})` }}
              initial={{ width: 0 }}
              animate={{ width: `${pct}%` }}
              transition={{ type: 'spring', stiffness: 120, damping: 20 }}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
