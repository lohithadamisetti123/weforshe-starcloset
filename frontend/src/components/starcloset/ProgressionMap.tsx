import { useEffect, useMemo, useRef, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Lock, Check, Sparkles, Crown, ArrowLeftRight, Star, Gift } from 'lucide-react'
import { useStarCloset } from '@/context/StarClosetContext'
import { Confetti } from '@/components/starcloset/Confetti'
import { cn } from '@/lib/utils'
import type { StarStylist } from '@/types/starcloset'

// ProgressionMap — a Candy-Crush / Royal-Match style stylist journey. A long,
// curving luxury runway threads huge circular stylist nodes; unlocked ones glow
// and are pickable, the current one breathes with orbiting stars, and locked
// ones sit behind frosted glass with a purchase countdown. Floating collectibles
// and milestone islands make the scroll itself feel rewarding. This is pure
// presentation over the existing unlock data — no logic, API, or persona change.

// ── Layout geometry (a narrow centred column, like the mobile games it apes) ──
const W = 360 // logical map width in px
const GAP = 214 // vertical distance between nodes
const TOP = 56 // padding above the first node
const AMP = 92 // horizontal sway of the curve
const R = 52 // node radius

const cx = (i: number) => W / 2 + AMP * Math.sin(i * 0.9 + 0.7)

// Decorative collectibles that float alongside the path.
const COLLECTIBLES = ['👠', '💄', '👜', '💎', '👑', '✨', '🌸', '⭐']
// Decorative milestone island labels (purely cosmetic, not tied to unlocks).
const MILESTONES = ['Rising Star', 'Style Maven', 'Luxury Icon', 'Red Carpet Royalty']

function initials(name: string) {
  const p = name.trim().split(/\s+/)
  return ((p[0]?.[0] ?? '') + (p.length > 1 ? p[p.length - 1][0] : p[0]?.[1] ?? '')).toUpperCase()
}

// A smooth S-curve between two points going downward.
function segPath(p0: { x: number; y: number }, p1: { x: number; y: number }) {
  const c = (p1.y - p0.y) * 0.5
  return `M ${p0.x} ${p0.y} C ${p0.x} ${p0.y + c}, ${p1.x} ${p1.y - c}, ${p1.x} ${p1.y}`
}

type NodeState = 'current' | 'completed' | 'locked'

function NodePortrait({ stylist, state }: { stylist: StarStylist; state: NodeState }) {
  const [imgOk, setImgOk] = useState(true)
  const locked = state === 'locked'
  return (
    <div
      className="relative h-full w-full overflow-hidden rounded-full"
      style={{ backgroundImage: `linear-gradient(135deg, ${stylist.accentFrom}, ${stylist.accentTo})` }}
    >
      {imgOk && stylist.portraitUrl ? (
        <img
          src={stylist.portraitUrl}
          alt={stylist.displayName}
          loading="lazy"
          decoding="async"
          onError={() => setImgOk(false)}
          className={cn('h-full w-full object-cover transition-all', locked && 'grayscale')}
          style={locked ? { filter: 'grayscale(1) blur(2px) brightness(0.9)' } : undefined}
        />
      ) : (
        <div className="flex h-full w-full items-center justify-center font-display text-xl font-bold text-white">
          {initials(stylist.displayName)}
        </div>
      )}
      {/* Frosted glass over locked portraits */}
      {locked && <div className="absolute inset-0 bg-white/25 backdrop-blur-[2px]" />}
    </div>
  )
}

export function ProgressionMap() {
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

  const [confetti, setConfetti] = useState(false)
  const [celebrating, setCelebrating] = useState<string[]>([])
  const nodeRefs = useRef<Record<string, HTMLDivElement | null>>({})
  const containerRef = useRef<HTMLDivElement>(null)

  const stylists = useMemo(() => currentDept?.stylists ?? [], [currentDept])

  // Geometry for every node + the final palace.
  const nodes = useMemo(
    () => stylists.map((s, i) => ({ s, i, x: cx(i), y: TOP + i * GAP })),
    [stylists],
  )
  const finalPos = { x: W / 2, y: TOP + stylists.length * GAP }
  const totalH = finalPos.y + 190

  const unlockedList = stylists.filter((s) => isUnlocked(s))
  const allUnlocked = unlockedList.length === stylists.length && stylists.length > 0

  // The "current" node — the stylist you're at: your selection if it's unlocked,
  // otherwise the most recently earned one.
  const featured =
    (selectedSlug && stylists.find((s) => s.slug === selectedSlug && isUnlocked(s))) ||
    unlockedList[unlockedList.length - 1] ||
    stylists[0]

  const nodeState = (s: StarStylist): NodeState => {
    if (!isUnlocked(s)) return 'locked'
    if (featured && s.slug === featured.slug) return 'current'
    return 'completed'
  }

  // HUD numbers (derived, read-only).
  const total = stylists.length
  const unlockedCount = unlockedList.length
  const maxThreshold = Math.max(1, ...stylists.map((s) => s.unlockAtItems))
  const pct = allUnlocked ? 100 : Math.min(100, Math.round((unlockScore / maxThreshold) * 100))
  const level = unlockedCount

  // Auto-scroll to the current stylist on mount / when it changes.
  useEffect(() => {
    const el = featured ? nodeRefs.current[featured.slug] : null
    if (el) {
      const t = setTimeout(() => el.scrollIntoView({ behavior: 'smooth', block: 'center' }), 250)
      return () => clearTimeout(t)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [featured?.slug, currentDept?.key])

  // Celebrate stylists unlocked since the last time this map was viewed — so
  // returning from the chat after a purchase pops confetti on the new node.
  useEffect(() => {
    if (!currentDept) return
    const key = `sc-seen-${currentDept.key}`
    let seen: string[] = []
    try {
      seen = JSON.parse(sessionStorage.getItem(key) || '[]')
    } catch {
      seen = []
    }
    const now = unlockedList.map((s) => s.slug)
    const fresh = now.filter((slug) => !seen.includes(slug))
    if (seen.length > 0 && fresh.length > 0) {
      setCelebrating(fresh)
      setConfetti(true)
      const target = nodeRefs.current[fresh[fresh.length - 1]]
      setTimeout(() => target?.scrollIntoView({ behavior: 'smooth', block: 'center' }), 300)
      const t = setTimeout(() => {
        setConfetti(false)
        setCelebrating([])
      }, 3200)
      sessionStorage.setItem(key, JSON.stringify(now))
      return () => clearTimeout(t)
    }
    sessionStorage.setItem(key, JSON.stringify(now))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [unlockScore, currentDept?.key])

  if (!currentDept) return null

  return (
    <div className="relative min-h-[calc(100vh-5rem)] overflow-hidden">
      {/* Keyframes for the ambient, non-transform-heavy motion. */}
      <style>{`
        @keyframes sc-twinkle {0%,100%{opacity:.15;transform:scale(.7)}50%{opacity:.9;transform:scale(1.15)}}
        @keyframes sc-cloud {0%{transform:translateX(-2%)}50%{transform:translateX(2%)}100%{transform:translateX(-2%)}}
        @keyframes sc-sheen {0%{background-position:-180% 0}100%{background-position:180% 0}}
      `}</style>

      {confetti && <Confetti count={60} />}

      {/* ── Luxury background: clouds, bokeh, sparkles ──────────────────────── */}
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-gradient-to-b from-rose-50 via-pink-50/60 to-white" />
        <div
          className="absolute -top-24 left-1/2 h-[520px] w-[820px] -translate-x-1/2 rounded-full opacity-60 blur-3xl"
          style={{ background: 'radial-gradient(closest-side, rgba(255,158,197,0.5), transparent)', animation: 'sc-cloud 16s ease-in-out infinite' }}
        />
        <div
          className="absolute top-1/3 -left-20 h-72 w-72 rounded-full opacity-40 blur-3xl"
          style={{ background: 'radial-gradient(closest-side, rgba(214,178,255,0.5), transparent)', animation: 'sc-cloud 22s ease-in-out infinite' }}
        />
        <div
          className="absolute bottom-24 -right-16 h-80 w-80 rounded-full opacity-40 blur-3xl"
          style={{ background: 'radial-gradient(closest-side, rgba(255,215,180,0.5), transparent)', animation: 'sc-cloud 19s ease-in-out infinite' }}
        />
        {Array.from({ length: 26 }).map((_, i) => (
          <span
            key={i}
            className="absolute rounded-full bg-white"
            style={{
              left: `${(i * 37) % 100}%`,
              top: `${(i * 53) % 100}%`,
              width: 3 + (i % 3),
              height: 3 + (i % 3),
              boxShadow: '0 0 6px rgba(255,255,255,0.9)',
              animation: `sc-twinkle ${2.4 + (i % 5) * 0.6}s ease-in-out ${(i % 7) * 0.3}s infinite`,
            }}
          />
        ))}
      </div>

      {/* ── Journey HUD (sticky top) ────────────────────────────────────────── */}
      <div className="sticky top-0 z-30 px-4 pb-2 pt-4">
        <div className="mx-auto max-w-md rounded-3xl border border-white/60 bg-white/70 p-4 shadow-[0_10px_40px_-12px_rgba(255,63,108,0.35)] backdrop-blur-xl">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-display text-sm font-bold text-ink-900">
                {currentDept.label}'s Fashion Journey
              </p>
              <div className="mt-1 flex items-center gap-2">
                <span className="rounded-full bg-gradient-to-r from-brand-500 to-rose-400 px-2 py-0.5 text-[10px] font-black uppercase tracking-wide text-white">
                  Level {level}
                </span>
                <span className="flex items-center gap-0.5">
                  {Array.from({ length: total }).map((_, i) => (
                    <Star
                      key={i}
                      className={cn('h-3.5 w-3.5', i < unlockedCount ? 'fill-amber-400 text-amber-400' : 'fill-ink-200 text-ink-200')}
                    />
                  ))}
                </span>
              </div>
            </div>
            <button
              onClick={clearDepartment}
              className="flex items-center gap-1.5 rounded-full border border-line bg-white px-3 py-1.5 text-xs font-semibold text-ink-500 transition-colors hover:border-ink-900 hover:text-ink-900"
            >
              <ArrowLeftRight className="h-3 w-3" /> Switch
            </button>
          </div>

          {/* Animated fill bar */}
          <div className="mt-3 h-2.5 w-full overflow-hidden rounded-full bg-ink-100">
            <motion.div
              className="h-full rounded-full bg-gradient-to-r from-brand-500 via-rose-400 to-amber-300"
              style={{ backgroundSize: '200% 100%', animation: 'sc-sheen 2.4s linear infinite' }}
              initial={{ width: 0 }}
              animate={{ width: `${pct}%` }}
              transition={{ type: 'spring', stiffness: 90, damping: 20 }}
            />
          </div>
          <div className="mt-1.5 flex items-center justify-between text-[11px]">
            <span className="font-bold text-brand-600">{pct}% Complete</span>
            {nextLocked ? (
              <span className="text-ink-500">
                Next: <span className="font-semibold text-ink-800">{nextLocked.displayName}</span> ·{' '}
                {itemsToUnlock(nextLocked)} away
              </span>
            ) : (
              <span className="font-semibold text-emerald-600">All stylists unlocked ✨</span>
            )}
          </div>
        </div>
      </div>

      {/* ── The map ─────────────────────────────────────────────────────────── */}
      <div ref={containerRef} className="relative mx-auto" style={{ width: W, height: totalH }}>
        {/* Runway ribbon (SVG) behind the nodes */}
        <svg className="absolute inset-0" width={W} height={totalH} viewBox={`0 0 ${W} ${totalH}`} fill="none">
          <defs>
            <linearGradient id="sc-ribbon-on" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#FF9EC5" />
              <stop offset="50%" stopColor="#FF3F6C" />
              <stop offset="100%" stopColor="#FFC46B" />
            </linearGradient>
            <filter id="sc-glow" x="-40%" y="-40%" width="180%" height="180%">
              <feGaussianBlur stdDeviation="4" result="b" />
              <feMerge>
                <feMergeNode in="b" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
          </defs>

          {nodes.map((n, i) => {
            const to = i < nodes.length - 1 ? nodes[i + 1] : finalPos
            const reached = i < nodes.length - 1 ? isUnlocked(nodes[i + 1].s) : allUnlocked
            const d = segPath({ x: n.x, y: n.y }, { x: to.x, y: to.y })
            return (
              <g key={`seg-${i}`}>
                {/* soft base */}
                <path d={d} stroke="#F3D9E4" strokeWidth={14} strokeLinecap="round" />
                {reached ? (
                  <path d={d} stroke="url(#sc-ribbon-on)" strokeWidth={9} strokeLinecap="round" filter="url(#sc-glow)" />
                ) : (
                  <path d={d} stroke="#E7DCE3" strokeWidth={7} strokeLinecap="round" strokeDasharray="2 12" />
                )}
              </g>
            )
          })}
        </svg>

        {/* Floating collectibles at segment midpoints */}
        {nodes.map((n, i) => {
          if (i >= nodes.length - 1) return null
          const to = nodes[i + 1]
          const midX = (n.x + to.x) / 2 + (i % 2 === 0 ? 84 : -84)
          const midY = (n.y + to.y) / 2
          const emoji = COLLECTIBLES[i % COLLECTIBLES.length]
          return (
            <motion.div
              key={`col-${i}`}
              className="absolute z-10 text-2xl drop-shadow"
              style={{ left: midX, top: midY }}
              animate={{ y: [0, -12, 0], rotate: [0, 6, -6, 0] }}
              transition={{ duration: 4 + (i % 3), repeat: Infinity, ease: 'easeInOut' }}
            >
              {emoji}
            </motion.div>
          )
        })}

        {/* Milestone islands (decorative) between selected nodes */}
        {nodes.map((n, i) => {
          if (i === 0 || i % 2 !== 0 || i >= nodes.length - 1) return null
          const to = nodes[i + 1]
          const midY = (n.y + to.y) / 2
          const side = i % 4 === 0 ? -1 : 1
          const label = MILESTONES[Math.floor(i / 2) % MILESTONES.length]
          const reached = isUnlocked(n.s)
          return (
            <motion.div
              key={`ms-${i}`}
              className="absolute z-10"
              style={{ left: W / 2 + side * 120, top: midY - 14, transform: 'translateX(-50%)' }}
              animate={{ y: [0, -6, 0] }}
              transition={{ duration: 5, repeat: Infinity, ease: 'easeInOut' }}
            >
              <div
                className={cn(
                  'flex items-center gap-1 rounded-full border px-2.5 py-1 text-[10px] font-bold shadow-sm backdrop-blur',
                  reached
                    ? 'border-amber-300 bg-amber-50/90 text-amber-700'
                    : 'border-line bg-white/70 text-ink-400',
                )}
              >
                <Crown className="h-3 w-3" /> {label}
              </div>
            </motion.div>
          )
        })}

        {/* Stylist nodes */}
        {nodes.map((n) => {
          const state = nodeState(n.s)
          const locked = state === 'locked'
          const current = state === 'current'
          const isFresh = celebrating.includes(n.s.slug)
          const first = n.s.displayName.split(' ')[0]

          return (
            <div
              key={n.s.slug}
              ref={(el) => { nodeRefs.current[n.s.slug] = el }}
              className="absolute z-20 flex flex-col items-center"
              style={{ left: n.x, top: n.y, transform: 'translate(-50%, -50%)' }}
            >
              {/* Portrait zone — aura + orbiting stars are centred on the node */}
              <div className="relative flex items-center justify-center" style={{ width: R * 2, height: R * 2 }}>
                {/* Breathing aura for the current stylist */}
                {current && (
                  <div
                    className="pointer-events-none absolute -z-10"
                    style={{ left: '50%', top: '50%', width: R * 3.1, height: R * 3.1, transform: 'translate(-50%,-50%)' }}
                  >
                    <motion.div
                      className="h-full w-full rounded-full"
                      style={{ background: `radial-gradient(closest-side, ${n.s.accentFrom}55, transparent)` }}
                      animate={{ scale: [1, 1.12, 1], opacity: [0.7, 1, 0.7] }}
                      transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
                    />
                  </div>
                )}

                {/* Orbiting stars for the current stylist */}
                {current &&
                  Array.from({ length: 3 }).map((_, k) => (
                    <motion.div
                      key={k}
                      className="pointer-events-none absolute left-1/2 top-1/2 h-0 w-0"
                      initial={{ rotate: k * 120 }}
                      animate={{ rotate: k * 120 + 360 }}
                      transition={{ duration: 8, repeat: Infinity, ease: 'linear' }}
                    >
                      <span
                        className="absolute text-amber-400"
                        style={{ transform: `translate(-50%,-50%) translateY(-${R + 20}px)` }}
                      >
                        <Sparkles className="h-3.5 w-3.5 drop-shadow" />
                      </span>
                    </motion.div>
                  ))}

                {/* The node itself */}
                <motion.button
                  onClick={() => !locked && selectPersona(n.s.slug)}
                  disabled={locked}
                  initial={isFresh ? { scale: 0.4, rotate: -12 } : false}
                  animate={
                    isFresh
                      ? { scale: [0.4, 1.25, 1], rotate: [-12, 0, 0] }
                      : current
                        ? { y: [0, -6, 0] }
                        : {}
                  }
                  transition={
                    isFresh
                      ? { duration: 0.9, ease: 'easeOut' }
                      : current
                        ? { duration: 3.2, repeat: Infinity, ease: 'easeInOut' }
                        : {}
                  }
                  whileHover={!locked ? { scale: 1.08, y: -8 } : {}}
                  whileTap={!locked ? { scale: 0.96 } : {}}
                  className={cn('group rounded-full', locked ? 'cursor-not-allowed' : 'cursor-pointer')}
                  style={{
                    width: R * 2,
                    height: R * 2,
                    padding: 4,
                    background: locked
                      ? 'linear-gradient(135deg,#e9e2ea,#d8cdd6)'
                      : `linear-gradient(135deg, ${n.s.accentFrom}, ${n.s.accentTo})`,
                    boxShadow: locked
                      ? '0 8px 20px -8px rgba(80,60,80,0.4)'
                      : `0 0 0 5px #ffffff, 0 14px 34px -10px ${n.s.accentTo}cc`,
                  }}
                >
                  <div className="relative h-full w-full rounded-full ring-2 ring-white/70">
                    <NodePortrait stylist={n.s} state={state} />

                    {/* Status pip */}
                    <span
                      className={cn(
                        'absolute -bottom-1 -right-1 flex h-8 w-8 items-center justify-center rounded-full border-[3px] border-white text-white shadow',
                        current ? 'bg-brand-500' : locked ? 'bg-ink-400' : 'bg-emerald-500',
                      )}
                    >
                      {current ? <Star className="h-4 w-4 fill-white" /> : locked ? <Lock className="h-4 w-4" /> : <Check className="h-4 w-4" />}
                    </span>

                    {/* Crown for completed (not current) */}
                    {state === 'completed' && (
                      <span className="absolute -top-4 left-1/2 -translate-x-1/2 text-lg drop-shadow">👑</span>
                    )}

                    {/* Hover sparkle */}
                    {!locked && (
                      <span className="pointer-events-none absolute -right-2 -top-2 text-amber-400 opacity-0 transition-opacity duration-200 group-hover:opacity-100">
                        <Sparkles className="h-4 w-4" />
                      </span>
                    )}
                  </div>
                </motion.button>
              </div>

              {/* Name + CTA / lock label */}
              <div className="mt-3 w-40 text-center">
                <p className={cn('font-display text-sm font-bold', locked ? 'text-ink-400' : 'text-ink-900')}>
                  {n.s.displayName}
                </p>
                {locked ? (
                  <p className="mt-0.5 text-[11px] font-semibold text-ink-400">
                    {itemsToUnlock(n.s) === 1 ? 'Almost there — 1 outfit' : `${itemsToUnlock(n.s)} outfits remaining`}
                  </p>
                ) : current ? (
                  <motion.button
                    onClick={() => selectPersona(n.s.slug)}
                    animate={{ scale: [1, 1.05, 1] }}
                    transition={{ duration: 1.8, repeat: Infinity, ease: 'easeInOut' }}
                    className="mt-1.5 inline-flex items-center gap-1 rounded-full bg-ink-900 px-4 py-2 text-xs font-bold text-white shadow-lg hover:bg-brand-600"
                  >
                    <Sparkles className="h-3.5 w-3.5" /> Style with {first}
                  </motion.button>
                ) : (
                  <button
                    onClick={() => selectPersona(n.s.slug)}
                    className="mt-1.5 inline-flex items-center gap-1 rounded-full border border-brand-200 bg-white px-3.5 py-1.5 text-xs font-semibold text-brand-600 hover:border-brand-500"
                  >
                    Style with {first}
                  </button>
                )}
              </div>
            </div>
          )
        })}

        {/* ── Final reward: the Ultimate Stylist palace ─────────────────────── */}
        <div
          className="absolute z-20 flex flex-col items-center"
          style={{ left: finalPos.x, top: finalPos.y, transform: 'translate(-50%, -50%)' }}
        >
          <motion.div
            animate={allUnlocked ? { y: [0, -8, 0], scale: [1, 1.04, 1] } : {}}
            transition={{ duration: 3.5, repeat: Infinity, ease: 'easeInOut' }}
            className={cn(
              'relative flex h-28 w-28 items-center justify-center rounded-3xl border-4 text-5xl shadow-xl',
              allUnlocked ? 'border-amber-300' : 'border-line',
            )}
            style={{
              background: allUnlocked
                ? 'linear-gradient(135deg,#FFE7A3,#FFC46B)'
                : 'linear-gradient(135deg,#ece6ee,#ddd3dc)',
              boxShadow: allUnlocked ? '0 0 40px -6px rgba(255,196,107,0.8)' : undefined,
            }}
          >
            <span className={cn(!allUnlocked && 'opacity-40 grayscale')}>🏰</span>
            {!allUnlocked && (
              <span className="absolute -bottom-2 -right-2 flex h-8 w-8 items-center justify-center rounded-full border-2 border-white bg-ink-400 text-white">
                <Lock className="h-4 w-4" />
              </span>
            )}
          </motion.div>
          <p className="mt-3 flex items-center gap-1 font-display text-sm font-black tracking-wide text-ink-900">
            <Gift className="h-4 w-4 text-brand-500" /> Ultimate Stylist
          </p>
          <p className={cn('text-[11px] font-semibold', allUnlocked ? 'text-amber-600' : 'text-ink-400')}>
            {allUnlocked ? 'Unlocked — the crown is yours 👑' : 'Unlock every stylist to reveal'}
          </p>
        </div>

        {/* One-shot golden halo on freshly-unlocked nodes */}
        <AnimatePresence>
          {celebrating.map((slug) => {
            const n = nodes.find((x) => x.s.slug === slug)
            if (!n) return null
            return (
              <motion.div
                key={`halo-${slug}`}
                className="pointer-events-none absolute z-10 rounded-full"
                style={{ left: n.x, top: n.y, transform: 'translate(-50%,-50%)', width: R * 4, height: R * 4, background: 'radial-gradient(closest-side, rgba(255,196,107,0.55), transparent)' }}
                initial={{ scale: 0.3, opacity: 0.9 }}
                animate={{ scale: 1.4, opacity: 0 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 1.4, ease: 'easeOut' }}
              />
            )
          })}
        </AnimatePresence>
      </div>
    </div>
  )
}
