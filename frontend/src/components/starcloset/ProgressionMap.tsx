import { useEffect, useMemo, useRef, useState } from 'react'
import { motion, AnimatePresence, useScroll, useTransform } from 'framer-motion'
import { Lock, Check, Sparkles, Crown, ArrowLeftRight, Star, Gift } from 'lucide-react'
import { useStarCloset } from '@/context/StarClosetContext'
import { Confetti } from '@/components/starcloset/Confetti'
import { cn } from '@/lib/utils'
import type { StarStylist } from '@/types/starcloset'

// ProgressionMap — a full-viewport, Candy-Crush / Royal-Match style fashion
// adventure. A wide winding luxury ribbon (with a glow that flows continuously
// through it) threads huge circular stylist nodes down an animated premium
// world: parallax clouds & bokeh, floating petals, scattered collectible reward
// badges and milestone islands. Purely presentational over the existing unlock
// data — no logic, API, persona or engine change.

// ── Reward badges (decorative collectibles) ──────────────────────────────────
type Rarity = 'bronze' | 'silver' | 'gold' | 'diamond'
const RARITY: Record<Rarity, { border: string; glow: string; chip: string }> = {
  bronze: { border: '#C79A5B', glow: 'rgba(199,154,91,0.45)', chip: '🥉' },
  silver: { border: '#B9C2CF', glow: 'rgba(200,210,225,0.55)', chip: '🥈' },
  gold: { border: '#E7B84B', glow: 'rgba(231,184,75,0.6)', chip: '🥇' },
  diamond: { border: '#8AE7FF', glow: 'rgba(140,231,255,0.65)', chip: '💎' },
}
const REWARDS: { icon: string; label: string; rarity: Rarity; tip: string }[] = [
  { icon: '👜', label: 'Style Maven', rarity: 'gold', tip: 'Unlocked after styling 10 outfits' },
  { icon: '👑', label: 'Fashion Queen', rarity: 'diamond', tip: 'Top Fashion Journey Reward' },
  { icon: '💎', label: 'Luxury Pick', rarity: 'diamond', tip: 'Reached Level 5' },
  { icon: '👠', label: 'Shoe Lover', rarity: 'silver', tip: 'Added 5 footwear picks' },
  { icon: '✨', label: 'Trend Setter', rarity: 'gold', tip: 'Followed 3 trends' },
  { icon: '💄', label: 'Beauty Icon', rarity: 'silver', tip: 'A true beauty maven' },
  { icon: '🕶', label: 'Street Star', rarity: 'bronze', tip: 'Streetwear unlocked' },
  { icon: '🧥', label: 'Layering Pro', rarity: 'bronze', tip: 'Master of layers' },
  { icon: '🎀', label: 'Ethnic Expert', rarity: 'gold', tip: 'Ethnic wear pro' },
  { icon: '🛍', label: 'Shopping Streak', rarity: 'silver', tip: '3-day styling streak' },
  { icon: '⭐', label: 'Style Points +50', rarity: 'bronze', tip: '+50 style points earned' },
  { icon: '💖', label: 'Fashion XP', rarity: 'bronze', tip: 'Fashion XP boost' },
  { icon: '🏆', label: 'Collection Master', rarity: 'diamond', tip: 'Completed a full collection' },
  { icon: '👗', label: 'Outfit Complete', rarity: 'silver', tip: 'A full outfit built' },
  { icon: '🔥', label: 'Trending Look', rarity: 'gold', tip: 'This look is trending now' },
]
const COLLECTIBLES = ['👠', '💄', '👜', '💎', '👑', '✨', '🌸', '🪞']
const MILESTONES = ['Wedding Expert', 'Office Fashion', 'Luxury Queen', 'Festival Master', 'Streetwear Icon']

const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v))

function initials(name: string) {
  const p = name.trim().split(/\s+/)
  return ((p[0]?.[0] ?? '') + (p.length > 1 ? p[p.length - 1][0] : p[0]?.[1] ?? '')).toUpperCase()
}

function segPath(a: { x: number; y: number }, b: { x: number; y: number }) {
  const c = (b.y - a.y) * 0.5
  return `M ${a.x} ${a.y} C ${a.x} ${a.y + c}, ${b.x} ${b.y - c}, ${b.x} ${b.y}`
}

function NodePortrait({ stylist, locked }: { stylist: StarStylist; locked: boolean }) {
  const [imgOk, setImgOk] = useState(true)
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
          className="h-full w-full object-cover"
          style={locked ? { filter: 'grayscale(1) blur(2px) brightness(0.92)' } : undefined}
        />
      ) : (
        <div className="flex h-full w-full items-center justify-center font-display text-2xl font-bold text-white">
          {initials(stylist.displayName)}
        </div>
      )}
      {locked && <div className="absolute inset-0 bg-white/25 backdrop-blur-[2px]" />}
    </div>
  )
}

function RewardBadge({
  reward,
  open,
  onToggle,
}: {
  reward: { icon: string; label: string; rarity: Rarity; tip: string }
  open: boolean
  onToggle: () => void
}) {
  const r = RARITY[reward.rarity]
  const diamond = reward.rarity === 'diamond'
  return (
    <motion.button
      onClick={(e) => {
        e.stopPropagation()
        onToggle()
      }}
      className="pointer-events-auto relative"
      animate={{ y: [0, -9, 0], rotate: [-1.5, 1.5, -1.5] }}
      transition={{ duration: 6 + (reward.label.length % 4), repeat: Infinity, ease: 'easeInOut' }}
      whileHover={{ scale: 1.12 }}
    >
      <div
        className="relative flex items-center gap-1.5 overflow-hidden rounded-full border-2 bg-white/80 px-3 py-1.5 shadow-lg backdrop-blur-md"
        style={{ borderColor: r.border, boxShadow: `0 6px 20px -6px ${r.glow}, 0 0 0 3px rgba(255,255,255,0.5)` }}
      >
        <span className="text-base leading-none">{reward.icon}</span>
        <span className="whitespace-nowrap font-display text-[11px] font-bold text-ink-800">{reward.label}</span>
        <span className="text-[9px]">{r.chip}</span>
        {/* light sweep */}
        <span
          className="pointer-events-none absolute inset-y-0 -left-1/3 w-1/3 skew-x-[-20deg] bg-white/60"
          style={{ animation: 'sc-shimmer 3.4s ease-in-out infinite' }}
        />
        {diamond && (
          <span
            className="pointer-events-none absolute inset-0 rounded-full"
            style={{ boxShadow: `inset 0 0 12px ${r.glow}`, animation: 'sc-pulse 2.2s ease-in-out infinite' }}
          />
        )}
      </div>
      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, y: 6, scale: 0.9 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 6, scale: 0.9 }}
            className="absolute left-1/2 top-full z-40 mt-2 w-max max-w-[180px] -translate-x-1/2 rounded-xl border border-white/70 bg-ink-900/90 px-3 py-1.5 text-center text-[10px] font-semibold text-white shadow-xl backdrop-blur"
          >
            {reward.tip}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.button>
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
  const [openBadge, setOpenBadge] = useState<number | null>(null)
  const [mapW, setMapW] = useState(() =>
    typeof window !== 'undefined' ? clamp(window.innerWidth - 24, 340, 760) : 380,
  )
  const nodeRefs = useRef<Record<string, HTMLDivElement | null>>({})

  useEffect(() => {
    const onResize = () => setMapW(clamp(window.innerWidth - 24, 340, 760))
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  // Parallax: background layers drift slower than the page for depth.
  const { scrollY } = useScroll()
  const yDeep = useTransform(scrollY, (v) => v * 0.14)
  const yMid = useTransform(scrollY, (v) => v * 0.07)
  const yNear = useTransform(scrollY, (v) => v * -0.05)

  const stylists = useMemo(() => currentDept?.stylists ?? [], [currentDept])
  const isMobile = mapW < 480

  // ── Geometry ────────────────────────────────────────────────────────────────
  const geo = useMemo(() => {
    const R = isMobile ? 58 : 66
    const GAP = R * 2 + 132
    const AMP = Math.min(mapW * 0.3, 200)
    const TOP = 150 // ≥120px of air below the pinned progress card
    const nodes = stylists.map((s, i) => ({
      s,
      i,
      x: mapW / 2 + AMP * Math.sin(i * 0.85 + 0.6),
      y: TOP + i * GAP,
    }))
    const finalPos = { x: mapW / 2, y: TOP + stylists.length * GAP }
    const totalH = finalPos.y + 240
    const points = [...nodes.map((n) => ({ x: n.x, y: n.y })), finalPos]
    const fullPath = points.reduce((acc, p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `${acc} ${segPath(points[i - 1], p).slice(2)}`), '')
    const pathXAt = (y: number) => {
      for (let i = 0; i < points.length - 1; i++) {
        const a = points[i]
        const b = points[i + 1]
        if (y >= a.y && y <= b.y) return a.x + (b.x - a.x) * ((y - a.y) / (b.y - a.y || 1))
      }
      return mapW / 2
    }
    return { R, GAP, AMP, TOP, nodes, finalPos, totalH, fullPath, pathXAt }
  }, [stylists, mapW, isMobile])

  const { R, nodes, finalPos, totalH, fullPath, pathXAt } = geo

  const unlockedList = stylists.filter((s) => isUnlocked(s))
  const allUnlocked = unlockedList.length === stylists.length && stylists.length > 0
  const featured =
    (selectedSlug && stylists.find((s) => s.slug === selectedSlug && isUnlocked(s))) ||
    unlockedList[unlockedList.length - 1] ||
    stylists[0]

  // Scattered reward-badge placements (kept clear of nodes: offset from the path
  // and, being between nodes, vertically separated from them).
  const badges = useMemo(() => {
    const list = isMobile ? REWARDS.filter((_, i) => i % 2 === 0) : REWARDS
    return list.map((reward, i) => {
      const t = (i + 0.6) / list.length
      const y = geo.TOP + t * (finalPos.y - geo.TOP)
      const side = i % 2 === 0 ? -1 : 1
      const px = pathXAt(y)
      let x = px + side * (R + (isMobile ? 62 : 120))
      if (x < 60 || x > mapW - 60 || Math.abs(x - px) < R + 44) x = clamp(px - side * (R + 80), 60, mapW - 60)
      return { reward, x: clamp(x, 56, mapW - 56), y }
    })
  }, [geo, finalPos.y, pathXAt, R, mapW, isMobile])

  // HUD numbers (derived, read-only).
  const total = stylists.length
  const unlockedCount = unlockedList.length
  const maxThreshold = Math.max(1, ...stylists.map((s) => s.unlockAtItems))
  const pct = allUnlocked ? 100 : Math.min(100, Math.round((unlockScore / maxThreshold) * 100))

  // Camera: centre the current stylist on open / change.
  useEffect(() => {
    const el = featured ? nodeRefs.current[featured.slug] : null
    if (el) {
      const t = setTimeout(() => el.scrollIntoView({ behavior: 'smooth', block: 'center' }), 260)
      return () => clearTimeout(t)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [featured?.slug, currentDept?.key])

  // Celebrate stylists unlocked since the last visit (confetti + halo + pan).
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
      setTimeout(() => nodeRefs.current[fresh[fresh.length - 1]]?.scrollIntoView({ behavior: 'smooth', block: 'center' }), 320)
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
    <div className="relative min-h-[calc(100vh-5rem)] overflow-hidden" onClick={() => setOpenBadge(null)}>
      <style>{`
        @keyframes sc-twinkle {0%,100%{opacity:.15;transform:scale(.7)}50%{opacity:.9;transform:scale(1.2)}}
        @keyframes sc-sheen {0%{background-position:-180% 0}100%{background-position:180% 0}}
        @keyframes sc-flow {to{stroke-dashoffset:-320}}
        @keyframes sc-shimmer {0%{transform:translateX(-30%) skewX(-20deg)}60%,100%{transform:translateX(560%) skewX(-20deg)}}
        @keyframes sc-pulse {0%,100%{opacity:.4}50%{opacity:1}}
      `}</style>

      {confetti && <Confetti count={70} />}

      {/* ── Parallax fashion world ─────────────────────────────────────────── */}
      <div className="pointer-events-none absolute inset-0 -z-10 overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-rose-50 via-pink-50/70 to-white" />
        {/* deep gradient waves */}
        <motion.div style={{ y: yDeep }} className="absolute inset-x-0 top-0" >
          <div className="absolute -top-32 left-1/2 h-[640px] w-[980px] -translate-x-1/2 rounded-full opacity-60 blur-3xl" style={{ background: 'radial-gradient(closest-side, rgba(255,158,197,0.5), transparent)' }} />
          <div className="absolute top-[900px] -left-24 h-96 w-96 rounded-full opacity-40 blur-3xl" style={{ background: 'radial-gradient(closest-side, rgba(214,178,255,0.55), transparent)' }} />
          <div className="absolute top-[1700px] -right-20 h-[420px] w-[420px] rounded-full opacity-40 blur-3xl" style={{ background: 'radial-gradient(closest-side, rgba(255,215,180,0.55), transparent)' }} />
          <div className="absolute top-[2500px] left-1/3 h-96 w-96 rounded-full opacity-40 blur-3xl" style={{ background: 'radial-gradient(closest-side, rgba(255,182,214,0.5), transparent)' }} />
        </motion.div>
        {/* mid bokeh */}
        <motion.div style={{ y: yMid }} className="absolute inset-0">
          {Array.from({ length: 14 }).map((_, i) => (
            <span key={i} className="absolute rounded-full bg-white/50 blur-md" style={{ left: `${(i * 61) % 100}%`, top: `${(i * 43 + 8) % 100}%`, width: 20 + (i % 4) * 14, height: 20 + (i % 4) * 14 }} />
          ))}
        </motion.div>
        {/* near sparkles + petals */}
        <motion.div style={{ y: yNear }} className="absolute inset-0">
          {Array.from({ length: 22 }).map((_, i) => (
            <span key={i} className="absolute rounded-full bg-white" style={{ left: `${(i * 37) % 100}%`, top: `${(i * 53) % 100}%`, width: 3 + (i % 3), height: 3 + (i % 3), boxShadow: '0 0 6px rgba(255,255,255,0.9)', animation: `sc-twinkle ${2.4 + (i % 5) * 0.6}s ease-in-out ${(i % 7) * 0.3}s infinite` }} />
          ))}
          {['🌸', '✨', '🌸', '💮', '✨', '🌸'].map((p, i) => (
            <motion.span key={i} className="absolute text-lg opacity-70" style={{ left: `${(i * 71 + 10) % 96}%`, top: `${(i * 47 + 12) % 96}%` }} animate={{ y: [0, -18, 0], rotate: [0, 12, -12, 0] }} transition={{ duration: 8 + i, repeat: Infinity, ease: 'easeInOut' }}>
              {p}
            </motion.span>
          ))}
        </motion.div>
      </div>

      {/* ── Progress card, pinned below the navbar ─────────────────────────── */}
      <div className="sticky top-[84px] z-30 px-4">
        <div className="mx-auto max-w-md rounded-3xl border border-white/60 bg-white/70 p-4 shadow-[0_16px_50px_-16px_rgba(255,63,108,0.45)] backdrop-blur-2xl">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <span className="flex h-9 w-9 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-500 to-rose-400 text-lg shadow">
                {currentDept.emoji}
              </span>
              <div>
                <p className="font-display text-sm font-bold text-ink-900">{currentDept.label}'s Fashion Journey</p>
                <div className="mt-1 flex items-center gap-2">
                  <span className="rounded-full bg-gradient-to-r from-brand-500 to-rose-400 px-2 py-0.5 text-[10px] font-black uppercase tracking-wide text-white">
                    Level {unlockedCount}
                  </span>
                  <span className="flex items-center gap-0.5">
                    {Array.from({ length: total }).map((_, i) => (
                      <Star key={i} className={cn('h-3.5 w-3.5', i < unlockedCount ? 'fill-amber-400 text-amber-400' : 'fill-ink-200 text-ink-200')} />
                    ))}
                  </span>
                </div>
              </div>
            </div>
            <button onClick={clearDepartment} className="flex items-center gap-1.5 rounded-full border border-line bg-white px-3 py-1.5 text-xs font-semibold text-ink-500 transition-colors hover:border-ink-900 hover:text-ink-900">
              <ArrowLeftRight className="h-3 w-3" /> Switch
            </button>
          </div>
          <div className="mt-3 h-2.5 w-full overflow-hidden rounded-full bg-ink-100">
            <motion.div className="h-full rounded-full bg-gradient-to-r from-brand-500 via-rose-400 to-amber-300" style={{ backgroundSize: '200% 100%', animation: 'sc-sheen 2.4s linear infinite' }} initial={{ width: 0 }} animate={{ width: `${pct}%` }} transition={{ type: 'spring', stiffness: 90, damping: 20 }} />
          </div>
          <div className="mt-1.5 flex items-center justify-between text-[11px]">
            <span className="font-bold text-brand-600">{pct}% · XP</span>
            {nextLocked ? (
              <span className="text-ink-500">Next: <span className="font-semibold text-ink-800">{nextLocked.displayName}</span> · {itemsToUnlock(nextLocked)} away</span>
            ) : (
              <span className="font-semibold text-emerald-600">All stylists unlocked ✨</span>
            )}
          </div>
        </div>
      </div>

      {/* ── The map ────────────────────────────────────────────────────────── */}
      <div className="relative mx-auto mt-6" style={{ width: mapW, height: totalH }}>
        {/* Glowing luxury ribbon */}
        <svg className="absolute inset-0" width={mapW} height={totalH} viewBox={`0 0 ${mapW} ${totalH}`} fill="none">
          <defs>
            <linearGradient id="sc-ribbon" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#FF9EC5" />
              <stop offset="50%" stopColor="#FF3F6C" />
              <stop offset="100%" stopColor="#FFC46B" />
            </linearGradient>
            <filter id="sc-glow" x="-40%" y="-40%" width="180%" height="180%">
              <feGaussianBlur stdDeviation="5" result="b" />
              <feMerge>
                <feMergeNode in="b" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
          </defs>

          {/* soft base ribbon */}
          <path d={fullPath} stroke="#F3D9E4" strokeWidth={18} strokeLinecap="round" />
          {/* reached segments light up */}
          {nodes.map((n, i) => {
            const to = i < nodes.length - 1 ? nodes[i + 1] : finalPos
            const reached = i < nodes.length - 1 ? isUnlocked(nodes[i + 1].s) : allUnlocked
            const d = segPath({ x: n.x, y: n.y }, { x: to.x, y: to.y })
            return reached ? (
              <path key={i} d={d} stroke="url(#sc-ribbon)" strokeWidth={11} strokeLinecap="round" filter="url(#sc-glow)" />
            ) : (
              <path key={i} d={d} stroke="#E7DCE3" strokeWidth={8} strokeLinecap="round" strokeDasharray="2 14" />
            )
          })}
          {/* continuous glow flowing through the ribbon */}
          <path d={fullPath} stroke="#ffffff" strokeWidth={5} strokeLinecap="round" opacity={0.85} filter="url(#sc-glow)" strokeDasharray="22 300" style={{ animation: 'sc-flow 3.2s linear infinite' }} />
        </svg>

        {/* Floating collectibles at segment midpoints */}
        {nodes.map((n, i) => {
          if (i >= nodes.length - 1) return null
          const to = nodes[i + 1]
          const midX = (n.x + to.x) / 2 + (i % 2 === 0 ? R + 20 : -(R + 20))
          const midY = (n.y + to.y) / 2
          return (
            <motion.div key={`col-${i}`} className="absolute z-10 text-2xl drop-shadow" style={{ left: clamp(midX, 20, mapW - 20), top: midY }} animate={{ y: [0, -12, 0], rotate: [0, 8, -8, 0] }} transition={{ duration: 4 + (i % 3), repeat: Infinity, ease: 'easeInOut' }}>
              {COLLECTIBLES[i % COLLECTIBLES.length]}
            </motion.div>
          )
        })}

        {/* Reward badges scattered around the world */}
        {badges.map((b, i) => (
          <div key={`badge-${i}`} className="absolute z-20" style={{ left: b.x, top: b.y, transform: 'translate(-50%,-50%)' }}>
            <RewardBadge reward={b.reward} open={openBadge === i} onToggle={() => setOpenBadge((v) => (v === i ? null : i))} />
          </div>
        ))}

        {/* Milestone islands (illustrated platforms) */}
        {nodes.map((n, i) => {
          if (i === 0 || i % 2 !== 0 || i >= nodes.length - 1) return null
          const to = nodes[i + 1]
          const midY = (n.y + to.y) / 2
          const side = i % 4 === 0 ? -1 : 1
          const label = MILESTONES[Math.floor(i / 2) % MILESTONES.length]
          const reached = isUnlocked(n.s)
          return (
            <motion.div key={`ms-${i}`} className="absolute z-10" style={{ left: clamp(mapW / 2 + side * (R + 74), 70, mapW - 70), top: midY, transform: 'translate(-50%,-50%)' }} animate={{ y: [0, -7, 0] }} transition={{ duration: 5.5, repeat: Infinity, ease: 'easeInOut' }}>
              <div className={cn('flex flex-col items-center rounded-2xl border px-3 py-2 shadow-lg backdrop-blur', reached ? 'border-amber-300 bg-amber-50/90' : 'border-line bg-white/70')}>
                <span className="text-xl drop-shadow">{reached ? '🏝️' : '🌫️'}</span>
                <span className={cn('mt-0.5 flex items-center gap-1 text-[10px] font-black', reached ? 'text-amber-700' : 'text-ink-400')}>
                  <Crown className="h-3 w-3" /> {label}
                </span>
              </div>
            </motion.div>
          )
        })}

        {/* Stylist nodes */}
        {nodes.map((n) => {
          const locked = !isUnlocked(n.s)
          const current = !locked && featured && n.s.slug === featured.slug
          const completed = !locked && !current
          const isFresh = celebrating.includes(n.s.slug)
          const first = n.s.displayName.split(' ')[0]

          return (
            <div
              key={n.s.slug}
              ref={(el) => {
                nodeRefs.current[n.s.slug] = el
              }}
              className="absolute z-20 flex flex-col items-center"
              style={{ left: n.x, top: n.y, transform: 'translate(-50%, -50%)' }}
            >
              <div className="relative flex items-center justify-center" style={{ width: R * 2, height: R * 2 }}>
                {current && (
                  <div className="pointer-events-none absolute -z-10" style={{ left: '50%', top: '50%', width: R * 3.2, height: R * 3.2, transform: 'translate(-50%,-50%)' }}>
                    <motion.div className="h-full w-full rounded-full" style={{ background: `radial-gradient(closest-side, ${n.s.accentFrom}55, transparent)` }} animate={{ scale: [1, 1.14, 1], opacity: [0.7, 1, 0.7] }} transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }} />
                  </div>
                )}
                {current &&
                  Array.from({ length: 3 }).map((_, k) => (
                    <motion.div key={k} className="pointer-events-none absolute left-1/2 top-1/2 h-0 w-0" initial={{ rotate: k * 120 }} animate={{ rotate: k * 120 + 360 }} transition={{ duration: 8, repeat: Infinity, ease: 'linear' }}>
                      <span className="absolute text-amber-400" style={{ transform: `translate(-50%,-50%) translateY(-${R + 22}px)` }}>
                        <Sparkles className="h-4 w-4 drop-shadow" />
                      </span>
                    </motion.div>
                  ))}

                <motion.button
                  onClick={() => !locked && selectPersona(n.s.slug)}
                  disabled={locked}
                  initial={isFresh ? { scale: 0.4, rotate: -12 } : false}
                  animate={isFresh ? { scale: [0.4, 1.25, 1], rotate: [-12, 0, 0] } : current ? { y: [0, -7, 0] } : {}}
                  transition={isFresh ? { duration: 0.9, ease: 'easeOut' } : current ? { duration: 3.2, repeat: Infinity, ease: 'easeInOut' } : {}}
                  whileHover={!locked ? { scale: 1.09, y: -8 } : {}}
                  whileTap={!locked ? { scale: 0.96 } : {}}
                  className={cn('group rounded-full', locked ? 'cursor-not-allowed' : 'cursor-pointer')}
                  style={{
                    width: R * 2,
                    height: R * 2,
                    padding: 5,
                    background: locked ? 'linear-gradient(135deg,#e9e2ea,#d8cdd6)' : `linear-gradient(135deg, ${n.s.accentFrom}, ${n.s.accentTo})`,
                    boxShadow: locked ? '0 10px 24px -8px rgba(80,60,80,0.45)' : `0 0 0 6px #ffffff, 0 18px 40px -10px ${n.s.accentTo}cc`,
                  }}
                >
                  <div className="relative h-full w-full rounded-full ring-2 ring-white/70">
                    <NodePortrait stylist={n.s} locked={locked} />
                    <span className={cn('absolute -bottom-1 -right-1 flex h-9 w-9 items-center justify-center rounded-full border-[3px] border-white text-white shadow', current ? 'bg-brand-500' : locked ? 'bg-ink-400' : 'bg-emerald-500')}>
                      {current ? <Star className="h-4 w-4 fill-white" /> : locked ? <Lock className="h-4 w-4" /> : (
                        <motion.span initial={{ scale: 0 }} animate={{ scale: 1 }} transition={{ type: 'spring', stiffness: 300, damping: 12 }}>
                          <Check className="h-4 w-4" />
                        </motion.span>
                      )}
                    </span>
                    {completed && <span className="absolute -top-4 left-1/2 -translate-x-1/2 text-xl drop-shadow">👑</span>}
                    {!locked && (
                      <span className="pointer-events-none absolute -right-2 -top-2 text-amber-400 opacity-0 transition-opacity duration-200 group-hover:opacity-100">
                        <Sparkles className="h-4 w-4" />
                      </span>
                    )}
                  </div>
                </motion.button>
              </div>

              <div className="mt-3 w-44 text-center">
                <p className={cn('font-display text-sm font-bold', locked ? 'text-ink-400' : 'text-ink-900')}>{n.s.displayName}</p>
                {locked ? (
                  <p className="mt-0.5 text-[11px] font-semibold text-ink-400">{itemsToUnlock(n.s) === 1 ? 'Almost there — 1 outfit' : `${itemsToUnlock(n.s)} outfits remaining`}</p>
                ) : current ? (
                  <motion.button onClick={() => selectPersona(n.s.slug)} animate={{ scale: [1, 1.05, 1] }} transition={{ duration: 1.8, repeat: Infinity, ease: 'easeInOut' }} className="mt-1.5 inline-flex items-center gap-1 rounded-full bg-ink-900 px-4 py-2 text-xs font-bold text-white shadow-lg hover:bg-brand-600">
                    <Sparkles className="h-3.5 w-3.5" /> Style with {first}
                  </motion.button>
                ) : (
                  <button onClick={() => selectPersona(n.s.slug)} className="mt-1.5 inline-flex items-center gap-1 rounded-full border border-brand-200 bg-white px-3.5 py-1.5 text-xs font-semibold text-brand-600 hover:border-brand-500">
                    Style with {first}
                  </button>
                )}
              </div>
            </div>
          )
        })}

        {/* Final reward: the Ultimate Stylist palace */}
        <div className="absolute z-20 flex flex-col items-center" style={{ left: finalPos.x, top: finalPos.y, transform: 'translate(-50%, -50%)' }}>
          <motion.div animate={allUnlocked ? { y: [0, -8, 0], scale: [1, 1.04, 1] } : {}} transition={{ duration: 3.5, repeat: Infinity, ease: 'easeInOut' }} className={cn('relative flex h-32 w-32 items-center justify-center rounded-3xl border-4 text-6xl shadow-xl', allUnlocked ? 'border-amber-300' : 'border-line')} style={{ background: allUnlocked ? 'linear-gradient(135deg,#FFE7A3,#FFC46B)' : 'linear-gradient(135deg,#ece6ee,#ddd3dc)', boxShadow: allUnlocked ? '0 0 48px -6px rgba(255,196,107,0.85)' : undefined }}>
            <span className={cn(!allUnlocked && 'opacity-40 grayscale')}>🏰</span>
            {!allUnlocked && (
              <span className="absolute -bottom-2 -right-2 flex h-9 w-9 items-center justify-center rounded-full border-2 border-white bg-ink-400 text-white">
                <Lock className="h-4 w-4" />
              </span>
            )}
          </motion.div>
          <p className="mt-3 flex items-center gap-1 font-display text-base font-black tracking-wide text-ink-900">
            <Gift className="h-4 w-4 text-brand-500" /> Ultimate Stylist
          </p>
          <p className={cn('text-[11px] font-semibold', allUnlocked ? 'text-amber-600' : 'text-ink-400')}>
            {allUnlocked ? 'Unlocked — the crown is yours 👑' : 'Unlock every stylist to reveal'}
          </p>
        </div>

        {/* Golden halo on freshly-unlocked nodes */}
        <AnimatePresence>
          {celebrating.map((slug) => {
            const n = nodes.find((x) => x.s.slug === slug)
            if (!n) return null
            return (
              <motion.div key={`halo-${slug}`} className="pointer-events-none absolute z-10 rounded-full" style={{ left: n.x, top: n.y, transform: 'translate(-50%,-50%)', width: R * 4, height: R * 4, background: 'radial-gradient(closest-side, rgba(255,196,107,0.55), transparent)' }} initial={{ scale: 0.3, opacity: 0.9 }} animate={{ scale: 1.4, opacity: 0 }} exit={{ opacity: 0 }} transition={{ duration: 1.4, ease: 'easeOut' }} />
            )
          })}
        </AnimatePresence>
      </div>
    </div>
  )
}
