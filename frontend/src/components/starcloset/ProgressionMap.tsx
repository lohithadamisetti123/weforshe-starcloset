import { motion } from 'framer-motion'
import { Lock, Check, Sparkles, ArrowLeftRight, Star } from 'lucide-react'
import { useStarCloset } from '@/context/StarClosetContext'
import { StylistAvatar } from '@/components/starcloset/StylistAvatar'
import { cn } from '@/lib/utils'
import type { StarStylist } from '@/types/starcloset'

// ProgressionMap is the gamified stylist journey for one department: a vertical
// path of milestone nodes (Duolingo / Candy-Crush style). Unlocked stylists glow
// and are pickable; locked ones show how many more cart items unlock them; the
// path fills as the shopper progresses. Selecting an unlocked node opens the
// studio with that stylist.
export function ProgressionMap({ onPicked }: { onPicked?: () => void }) {
  const {
    currentDept,
    unlockScore,
    isUnlocked,
    itemsToUnlock,
    selectedSlug,
    selectPersona,
    clearDepartment,
    nextLocked,
  } = useStarCloset()

  if (!currentDept) return null
  const stylists = currentDept.stylists

  const handlePick = (s: StarStylist) => {
    if (!isUnlocked(s)) return
    selectPersona(s.slug)
    onPicked?.()
  }

  return (
    <div className="mx-auto max-w-2xl px-5 py-8 lg:px-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-3xl">{currentDept.emoji}</span>
          <div>
            <h1 className="font-display text-xl font-bold text-ink-900">{currentDept.label} Journey</h1>
            <p className="text-xs font-semibold text-brand-600">
              {unlockScore} item{unlockScore === 1 ? '' : 's'} added
              {nextLocked ? ` · ${itemsToUnlock(nextLocked)} more to unlock ${nextLocked.displayName}` : ' · all unlocked!'}
            </p>
          </div>
        </div>
        <button
          onClick={clearDepartment}
          className="flex items-center gap-1.5 rounded-full border border-line px-3 py-1.5 text-xs font-medium text-ink-500 transition-colors hover:border-ink-900 hover:text-ink-900"
        >
          <ArrowLeftRight className="h-3 w-3" /> Switch
        </button>
      </div>

      {/* The path */}
      <div className="mt-8">
        {stylists.map((s, i) => {
          const unlocked = isUnlocked(s)
          const active = selectedSlug === s.slug
          const reachedHere = unlocked // path is "walked" up to unlocked nodes
          const firstName = s.displayName.split(' ')[0]

          return (
            <div key={s.slug} className="flex gap-4">
              {/* Rail + node */}
              <div className="flex w-16 shrink-0 flex-col items-center">
                {i > 0 && (
                  <div
                    className={cn('w-1.5 flex-1 rounded-full', reachedHere ? '' : 'bg-ink-100')}
                    style={
                      reachedHere
                        ? { minHeight: 28, backgroundImage: `linear-gradient(${s.accentFrom}, ${s.accentTo})` }
                        : { minHeight: 28 }
                    }
                  />
                )}
                <motion.button
                  initial={{ scale: 0.6, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  transition={{ delay: i * 0.06, type: 'spring', stiffness: 260, damping: 18 }}
                  whileHover={unlocked ? { scale: 1.08 } : undefined}
                  onClick={() => handlePick(s)}
                  disabled={!unlocked}
                  className={cn(
                    'relative rounded-full',
                    unlocked ? 'cursor-pointer' : 'cursor-not-allowed',
                    active && 'ring-4 ring-brand-400 ring-offset-2',
                    unlocked && !active && 'ring-2 ring-white',
                  )}
                  style={
                    unlocked
                      ? { boxShadow: `0 0 0 4px ${s.accentFrom}22, 0 8px 24px -6px ${s.accentTo}88` }
                      : undefined
                  }
                >
                  <StylistAvatar persona={s} size="lg" locked={!unlocked} />
                  {/* status pip */}
                  <span
                    className={cn(
                      'absolute -bottom-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full border-2 border-white text-white',
                      active ? 'bg-brand-500' : unlocked ? 'bg-emerald-500' : 'bg-ink-400',
                    )}
                  >
                    {active ? <Star className="h-3 w-3 fill-white" /> : unlocked ? <Check className="h-3 w-3" /> : <Lock className="h-3 w-3" />}
                  </span>
                </motion.button>
                {i < stylists.length - 1 && (
                  <div
                    className={cn('w-1.5 flex-1 rounded-full', isUnlocked(stylists[i + 1]) ? '' : 'bg-ink-100')}
                    style={
                      isUnlocked(stylists[i + 1])
                        ? { minHeight: 28, backgroundImage: `linear-gradient(${s.accentFrom}, ${s.accentTo})` }
                        : { minHeight: 28 }
                    }
                  />
                )}
              </div>

              {/* Card */}
              <div className="flex-1 py-2">
                <motion.div
                  initial={{ opacity: 0, x: 12 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: i * 0.06 + 0.05 }}
                  className={cn(
                    'rounded-2xl border p-4 transition-shadow',
                    active
                      ? 'border-brand-400 bg-brand-50/40'
                      : unlocked
                        ? 'border-line bg-white hover:shadow-md'
                        : 'border-dashed border-line bg-ink-50/40',
                  )}
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="min-w-0">
                      <p className={cn('font-display text-base font-bold', unlocked ? 'text-ink-900' : 'text-ink-500')}>
                        {s.displayName}
                      </p>
                      <p className="mt-0.5 line-clamp-1 text-xs text-ink-500">{s.tagline}</p>
                    </div>
                    {unlocked ? (
                      <button
                        onClick={() => handlePick(s)}
                        className={cn(
                          'shrink-0 rounded-full px-4 py-2 text-xs font-semibold transition-colors',
                          active ? 'bg-ink-900 text-white hover:bg-brand-600' : 'bg-brand-500 text-white hover:bg-brand-600',
                        )}
                      >
                        {active ? (
                          'Continue'
                        ) : (
                          <span className="flex items-center gap-1">
                            <Sparkles className="h-3 w-3" /> Style with {firstName}
                          </span>
                        )}
                      </button>
                    ) : (
                      <span className="shrink-0 rounded-full bg-ink-100 px-3 py-2 text-[11px] font-bold text-ink-500">
                        {itemsToUnlock(s)} more item{itemsToUnlock(s) === 1 ? '' : 's'}
                      </span>
                    )}
                  </div>
                </motion.div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
