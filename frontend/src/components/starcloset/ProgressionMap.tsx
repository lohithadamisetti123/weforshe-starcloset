import { useEffect, useMemo, useRef, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Lock, Check, Sparkles, Crown, ArrowLeftRight, Star, Gift } from 'lucide-react'
import { useStarCloset } from '@/context/StarClosetContext'
import { Confetti } from '@/components/starcloset/Confetti'
import { cn } from '@/lib/utils'
import type { StarStylist } from '@/types/starcloset'

// ProgressionMap — a full-screen fashion adventure that snakes LEFT↔RIGHT across
// the page (Candy Crush / Mario World Map style) over the department's own
// environment art. Backgrounds are the existing per-department images; the
// ribbon, nodes, milestones, rewards and decorations are all rendered
// dynamically on top. Purely presentational over the existing unlock data — no
// logic, API, persona or engine change.

const BG: Record<string, string> = {
  Women: '/images/backgrounds/women.png',
  Men: '/images/backgrounds/men.png',
  Kids: '/images/backgrounds/kids.png',
}

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
  { icon: '🛍', label: 'Shopping Master', rarity: 'silver', tip: '3-day styling streak' },
  { icon: '⭐', label: 'Style Points +50', rarity: 'bronze', tip: '+50 style points earned' },
  { icon: '💖', label: 'Fashion XP', rarity: 'bronze', tip: 'Fashion XP boost' },
  { icon: '🏆', label: 'Collection Master', rarity: 'diamond', tip: 'Completed a full collection' },
  { icon: '👗', label: 'Outfit Complete', rarity: 'silver', tip: 'A full outfit built' },
  { icon: '🔥', label: 'Trending Look', rarity: 'gold', tip: 'This look is trending now' },
]
const DECOR = ['👜', '👠', '💄', '🧴', '🌸', '🦋', '💎', '👑', '⭐', '✨', '📖', '🎁', '🪞', '🌺']
const MILESTONES = [
  { icon: '👑', label: 'Fashion Queen' },
  { icon: '💎', label: 'Luxury Icon' },
  { icon: '👗', label: 'Wedding Expert' },
  { icon: '🛍', label: 'Shopping Master' },
  { icon: '✨', label: 'Style Legend' },
]

const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v))

function initials(name: string) {
  const p = name.trim().split(/\s+/)
  return ((p[0]?.[0] ?? '') + (p.length > 1 ? p[p.length - 1][0] : p[0]?.[1] ?? '')).toUpperCase()
}

// Smooth link between two waypoints — horizontal along a row, an S-turn between
// rows.
function curve(a: { x: number; y: number }, b: { x: number; y: number }) {
  if (Math.abs(a.y - b.y) < 4) {
    const mx = (a.x + b.x) / 2
    return `C ${mx} ${a.y}, ${mx} ${b.y}, ${b.x} ${b.y}`
  }
  const my = (a.y + b.y) / 2
  return `C ${a.x} ${my}, ${b.x} ${my}, ${b.x} ${b.y}`
}

function NodePortrait({ stylist, locked }: { stylist: StarStylist; locked: boolean }) {
  const [imgOk, setImgOk] = useState(true)
  return (
    <div className="relative h-full w-full overflow-hidden rounded-full" style={{ backgroundImage: `linear-gradient(135deg, ${stylist.accentFrom}, ${stylist.accentTo})` }}>
      {imgOk && stylist.portraitUrl ? (
        <img src={stylist.portraitUrl} alt={stylist.displayName} loading="lazy" decoding="async" onError={() => setImgOk(false)} className="h-full w-full object-cover" style={locked ? { filter: 'grayscale(1) blur(2px) brightness(0.92)' } : undefined} />
      ) : (
        <div className="flex h-full w-full items-center justify-center font-display text-2xl font-bold text-white">{initials(stylist.displayName)}</div>
      )}
      {locked && <div className="absolute inset-0 bg-white/25 backdrop-blur-[2px]" />}
    </div>
  )
}

function RewardBadge({ reward, open, onToggle }: { reward: (typeof REWARDS)[number]; open: boolean; onToggle: () => void }) {
  const r = RARITY[reward.rarity]
  const diamond = reward.rarity === 'diamond'
  return (
    <motion.button onClick={(e) => { e.stopPropagation(); onToggle() }} className="pointer-events-auto relative" animate={{ y: [0, -9, 0], rotate: [-1.5, 1.5, -1.5] }} transition={{ duration: 6 + (reward.label.length % 4), repeat: Infinity, ease: 'easeInOut' }} whileHover={{ scale: 1.12 }}>
      <div className="relative flex items-center gap-1.5 overflow-hidden rounded-full border-2 bg-white/85 px-3 py-1.5 shadow-lg backdrop-blur-md" style={{ borderColor: r.border, boxShadow: `0 6px 20px -6px ${r.glow}, 0 0 0 3px rgba(255,255,255,0.5)` }}>
        <span className="text-base leading-none">{reward.icon}</span>
        <span className="whitespace-nowrap font-display text-[11px] font-bold text-ink-800">{reward.label}</span>
        <span className="text-[9px]">{r.chip}</span>
        <span className="pointer-events-none absolute inset-y-0 -left-1/3 w-1/3 skew-x-[-20deg] bg-white/60" style={{ animation: 'sc-shimmer 3.4s ease-in-out infinite' }} />
        {diamond && <span className="pointer-events-none absolute inset-0 rounded-full" style={{ boxShadow: `inset 0 0 12px ${r.glow}`, animation: 'sc-pulse 2.2s ease-in-out infinite' }} />}
      </div>
      <AnimatePresence>
        {open && (
          <motion.div initial={{ opacity: 0, y: 6, scale: 0.9 }} animate={{ opacity: 1, y: 0, scale: 1 }} exit={{ opacity: 0, y: 6, scale: 0.9 }} className="absolute left-1/2 top-full z-40 mt-2 w-max max-w-[180px] -translate-x-1/2 rounded-xl border border-white/70 bg-ink-900/90 px-3 py-1.5 text-center text-[10px] font-semibold text-white shadow-xl backdrop-blur">
            {reward.tip}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.button>
  )
}

export function ProgressionMap() {
  const { currentDept, unlockScore, isUnlocked, itemsToUnlock, selectedSlug, selectPersona, clearDepartment, nextLocked } = useStarCloset()

  const [confetti, setConfetti] = useState(false)
  const [celebrating, setCelebrating] = useState<string[]>([])
  const [openBadge, setOpenBadge] = useState<number | null>(null)
  const [vw, setVw] = useState(() => (typeof window !== 'undefined' ? window.innerWidth : 1024))
  const nodeRefs = useRef<Record<string, HTMLDivElement | null>>({})

  useEffect(() => {
    const on = () => setVw(window.innerWidth)
    window.addEventListener('resize', on)
    return () => window.removeEventListener('resize', on)
  }, [])

  const stylists = useMemo(() => currentDept?.stylists ?? [], [currentDept])
  const mapW = clamp(vw - 16, 340, 1360)
  const isMobile = mapW < 560

  // ── Zig-zag geometry (boustrophedon rows) ────────────────────────────────────
  const geo = useMemo(() => {
    const R = isMobile ? 48 : 62
    const marginX = R + 30
    const perRow = clamp(Math.floor((mapW - 2 * marginX) / 300), 2, 4)
    const columns =
      perRow === 1
        ? [mapW / 2]
        : Array.from({ length: perRow }, (_, c) => marginX + (c * (mapW - 2 * marginX)) / (perRow - 1))
    const rowGap = R * 2 + 150
    const TOP = 200 // ≥120px of air below the pinned progress card
    const N = stylists.length
    const rows = Math.max(1, Math.ceil(N / perRow))

    const nodes = stylists.map((s, i) => {
      const row = Math.floor(i / perRow)
      const pos = i - row * perRow
      const colIndex = row % 2 === 0 ? pos : perRow - 1 - pos
      return { s, i, x: columns[clamp(colIndex, 0, perRow - 1)], y: TOP + row * rowGap }
    })
    const last = nodes[N - 1] ?? { x: mapW / 2, y: TOP }
    const finalPos = { x: last.x, y: TOP + rows * rowGap }
    const totalH = finalPos.y + 200

    const pts = [...nodes.map((n) => ({ x: n.x, y: n.y })), finalPos]
    let fullPath = `M ${pts[0].x} ${pts[0].y}`
    for (let i = 1; i < pts.length; i++) fullPath += ' ' + curve(pts[i - 1], pts[i])

    // Free decoration slots — a grid minus anything near a node.
    const slots: { x: number; y: number }[] = []
    for (let gy = TOP - rowGap * 0.3; gy <= finalPos.y + rowGap * 0.2; gy += rowGap / 2) {
      for (const cxv of columns) slots.push({ x: cxv, y: gy })
      slots.push({ x: marginX * 0.55, y: gy })
      slots.push({ x: mapW - marginX * 0.55, y: gy })
    }
    const occupied = [...nodes, finalPos]
    const free = slots.filter(
      (s) => s.x > 40 && s.x < mapW - 40 && occupied.every((o) => Math.hypot(o.x - s.x, o.y - s.y) > R + 66),
    )

    return { R, marginX, perRow, columns, rowGap, TOP, nodes, finalPos, totalH, fullPath, free }
  }, [stylists, mapW, isMobile])

  const { R, nodes, finalPos, totalH, fullPath, free } = geo

  const unlockedList = stylists.filter((s) => isUnlocked(s))
  const allUnlocked = unlockedList.length === stylists.length && stylists.length > 0
  const featured =
    (selectedSlug && stylists.find((s) => s.slug === selectedSlug && isUnlocked(s))) ||
    unlockedList[unlockedList.length - 1] ||
    stylists[0]

  // Scatter badges / decorations / milestones across free slots.
  const scatter = useMemo(() => {
    const badgeN = isMobile ? 6 : Math.min(REWARDS.length, free.length)
    const badges = REWARDS.slice(0, badgeN).map((reward, i) => ({ reward, ...free[(i * 2) % Math.max(1, free.length)] }))
    const used = new Set(badges.map((b) => `${b.x},${b.y}`))
    const rest = free.filter((s) => !used.has(`${s.x},${s.y}`))
    const decor = rest.slice(0, isMobile ? 6 : 12).map((s, i) => ({ emoji: DECOR[i % DECOR.length], ...s }))
    // milestones on a couple of the widest-apart free slots
    const msSlots = rest.filter((_, i) => i % 5 === 3).slice(0, 2)
    const milestones = msSlots.map((s, i) => ({ ...MILESTONES[i % MILESTONES.length], ...s }))
    return { badges, decor, milestones }
  }, [free, isMobile])

  const total = stylists.length
  const unlockedCount = unlockedList.length
  const maxThreshold = Math.max(1, ...stylists.map((s) => s.unlockAtItems))
  const pct = allUnlocked ? 100 : Math.min(100, Math.round((unlockScore / maxThreshold) * 100))

  // Camera: centre the current stylist.
  useEffect(() => {
    const el = featured ? nodeRefs.current[featured.slug] : null
    if (el) {
      const t = setTimeout(() => el.scrollIntoView({ behavior: 'smooth', block: 'center' }), 280)
      return () => clearTimeout(t)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [featured?.slug, currentDept?.key])

  // Celebrate stylists unlocked since last visit.
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
      setTimeout(() => nodeRefs.current[fresh[fresh.length - 1]]?.scrollIntoView({ behavior: 'smooth', block: 'center' }), 340)
      const t = setTimeout(() => { setConfetti(false); setCelebrating([]) }, 3200)
      sessionStorage.setItem(key, JSON.stringify(now))
      return () => clearTimeout(t)
    }
    sessionStorage.setItem(key, JSON.stringify(now))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [unlockScore, currentDept?.key])

  if (!currentDept) return null
  const bg = BG[currentDept.key] ?? BG.Women

  return (
    <div className="relative min-h-[calc(100vh-5rem)]" onClick={() => setOpenBadge(null)}>
      <style>{`
        @keyframes sc-twinkle {0%,100%{opacity:.2;transform:scale(.7)}50%{opacity:.95;transform:scale(1.2)}}
        @keyframes sc-sheen {0%{background-position:-180% 0}100%{background-position:180% 0}}
        @keyframes sc-flow {to{stroke-dashoffset:-320}}
        @keyframes sc-shimmer {0%{transform:translateX(-30%) skewX(-20deg)}60%,100%{transform:translateX(560%) skewX(-20deg)}}
        @keyframes sc-pulse {0%,100%{opacity:.4}50%{opacity:1}}
      `}</style>

      {confetti && <Confetti count={70} />}

      {/* ── Fixed department environment (stays put while the map scrolls) ──── */}
      <div aria-hidden className="pointer-events-none sticky top-0 z-0 -mb-[100vh] h-screen w-full overflow-hidden">
        <img src={bg} alt="" className="h-full w-full object-cover object-center" />
        {/* subtle readability wash */}
        <div className="absolute inset-0 bg-gradient-to-b from-black/25 via-black/10 to-black/35" />
        <div className="absolute inset-0 bg-white/10" />
        {/* ambient sparkles over the scene */}
        {Array.from({ length: 16 }).map((_, i) => (
          <span key={i} className="absolute rounded-full bg-white" style={{ left: `${(i * 61) % 100}%`, top: `${(i * 37) % 100}%`, width: 3 + (i % 3), height: 3 + (i % 3), boxShadow: '0 0 8px rgba(255,255,255,0.9)', animation: `sc-twinkle ${2.6 + (i % 5) * 0.7}s ease-in-out ${(i % 6) * 0.3}s infinite` }} />
        ))}
      </div>

      {/* ── Everything renders on top of the environment ───────────────────── */}
      <div className="relative z-10">
        {/* Progress card, pinned below the navbar */}
        <div className="sticky top-[84px] z-30 px-4">
          <div className="mx-auto max-w-md rounded-3xl border border-white/60 bg-white/80 p-4 shadow-[0_16px_50px_-16px_rgba(0,0,0,0.4)] backdrop-blur-2xl">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2.5">
                <span className="flex h-9 w-9 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-500 to-rose-400 text-lg shadow">{currentDept.emoji}</span>
                <div>
                  <p className="font-display text-sm font-bold text-ink-900">{currentDept.label}'s Fashion Journey</p>
                  <div className="mt-1 flex items-center gap-2">
                    <span className="rounded-full bg-gradient-to-r from-brand-500 to-rose-400 px-2 py-0.5 text-[10px] font-black uppercase tracking-wide text-white">Level {unlockedCount}</span>
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

        {/* The zig-zag map */}
        <div className="relative mx-auto mt-6" style={{ width: mapW, height: totalH }}>
          {/* Luxury ribbon */}
          <svg className="absolute inset-0" width={mapW} height={totalH} viewBox={`0 0 ${mapW} ${totalH}`} fill="none">
            <defs>
              <linearGradient id="sc-ribbon" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stopColor="#FF9EC5" />
                <stop offset="50%" stopColor="#FF3F6C" />
                <stop offset="100%" stopColor="#FFC46B" />
              </linearGradient>
              <filter id="sc-glow" x="-20%" y="-20%" width="140%" height="140%">
                <feGaussianBlur stdDeviation="5" result="b" />
                <feMerge><feMergeNode in="b" /><feMergeNode in="SourceGraphic" /></feMerge>
              </filter>
            </defs>
            <path d={fullPath} stroke="#ffffff" strokeOpacity={0.5} strokeWidth={22} strokeLinecap="round" strokeLinejoin="round" />
            {nodes.map((n, i) => {
              const to = i < nodes.length - 1 ? nodes[i + 1] : finalPos
              const reached = i < nodes.length - 1 ? isUnlocked(nodes[i + 1].s) : allUnlocked
              const d = `M ${n.x} ${n.y} ${curve({ x: n.x, y: n.y }, { x: to.x, y: to.y })}`
              return reached ? (
                <path key={i} d={d} stroke="url(#sc-ribbon)" strokeWidth={12} strokeLinecap="round" strokeLinejoin="round" filter="url(#sc-glow)" />
              ) : (
                <path key={i} d={d} stroke="#ffffff" strokeOpacity={0.55} strokeWidth={8} strokeLinecap="round" strokeDasharray="2 14" />
              )
            })}
            <path d={fullPath} stroke="#ffffff" strokeWidth={5} strokeLinecap="round" opacity={0.9} filter="url(#sc-glow)" strokeDasharray="22 300" style={{ animation: 'sc-flow 3.2s linear infinite' }} />
          </svg>

          {/* Floating decorations */}
          {scatter.decor.map((d, i) => (
            <motion.div key={`dec-${i}`} className="absolute z-10 text-2xl drop-shadow-lg" style={{ left: d.x, top: d.y, transform: 'translate(-50%,-50%)' }} animate={{ y: [0, -12, 0], rotate: [0, 8, -8, 0] }} transition={{ duration: 4 + (i % 4), repeat: Infinity, ease: 'easeInOut' }}>
              {d.emoji}
            </motion.div>
          ))}

          {/* Milestone islands */}
          {scatter.milestones.map((m, i) => (
            <motion.div key={`ms-${i}`} className="absolute z-10" style={{ left: m.x, top: m.y, transform: 'translate(-50%,-50%)' }} animate={{ y: [0, -7, 0] }} transition={{ duration: 5.5, repeat: Infinity, ease: 'easeInOut' }}>
              <div className="flex flex-col items-center rounded-2xl border border-amber-300 bg-amber-50/90 px-3 py-2 shadow-lg backdrop-blur">
                <span className="text-2xl drop-shadow">🏝️</span>
                <span className="mt-0.5 flex items-center gap-1 text-[10px] font-black text-amber-700">
                  <span>{m.icon}</span> {m.label}
                </span>
              </div>
            </motion.div>
          ))}

          {/* Reward badges */}
          {scatter.badges.map((b, i) => (
            <div key={`badge-${i}`} className="absolute z-20" style={{ left: b.x, top: b.y, transform: 'translate(-50%,-50%)' }}>
              <RewardBadge reward={b.reward} open={openBadge === i} onToggle={() => setOpenBadge((v) => (v === i ? null : i))} />
            </div>
          ))}

          {/* Stylist nodes */}
          {nodes.map((n) => {
            const locked = !isUnlocked(n.s)
            const current = !locked && featured && n.s.slug === featured.slug
            const completed = !locked && !current
            const isFresh = celebrating.includes(n.s.slug)
            const first = n.s.displayName.split(' ')[0]

            return (
              <div key={n.s.slug} ref={(el) => { nodeRefs.current[n.s.slug] = el }} className="absolute z-20 flex flex-col items-center" style={{ left: n.x, top: n.y, transform: 'translate(-50%, -50%)' }}>
                <div className="relative flex items-center justify-center" style={{ width: R * 2, height: R * 2 }}>
                  {current && (
                    <div className="pointer-events-none absolute -z-10" style={{ left: '50%', top: '50%', width: R * 3.2, height: R * 3.2, transform: 'translate(-50%,-50%)' }}>
                      <motion.div className="h-full w-full rounded-full" style={{ background: `radial-gradient(closest-side, ${n.s.accentFrom}66, transparent)` }} animate={{ scale: [1, 1.14, 1], opacity: [0.75, 1, 0.75] }} transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }} />
                    </div>
                  )}
                  {current && Array.from({ length: 3 }).map((_, k) => (
                    <motion.div key={k} className="pointer-events-none absolute left-1/2 top-1/2 h-0 w-0" initial={{ rotate: k * 120 }} animate={{ rotate: k * 120 + 360 }} transition={{ duration: 8, repeat: Infinity, ease: 'linear' }}>
                      <span className="absolute text-amber-300" style={{ transform: `translate(-50%,-50%) translateY(-${R + 22}px)` }}><Sparkles className="h-4 w-4 drop-shadow" /></span>
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
                    style={{ width: R * 2, height: R * 2, padding: 5, background: locked ? 'linear-gradient(135deg,#e9e2ea,#d8cdd6)' : `linear-gradient(135deg, ${n.s.accentFrom}, ${n.s.accentTo})`, boxShadow: locked ? '0 10px 24px -8px rgba(0,0,0,0.5)' : `0 0 0 6px #ffffff, 0 18px 44px -10px ${n.s.accentTo}` }}
                  >
                    <div className="relative h-full w-full rounded-full ring-2 ring-white/70">
                      <NodePortrait stylist={n.s} locked={locked} />
                      <span className={cn('absolute -bottom-1 -right-1 flex h-9 w-9 items-center justify-center rounded-full border-[3px] border-white text-white shadow', current ? 'bg-brand-500' : locked ? 'bg-ink-400' : 'bg-emerald-500')}>
                        {current ? <Star className="h-4 w-4 fill-white" /> : locked ? <Lock className="h-4 w-4" /> : (
                          <motion.span initial={{ scale: 0 }} animate={{ scale: 1 }} transition={{ type: 'spring', stiffness: 300, damping: 12 }}><Check className="h-4 w-4" /></motion.span>
                        )}
                      </span>
                      {completed && <span className="absolute -top-4 left-1/2 -translate-x-1/2 text-xl drop-shadow">👑</span>}
                      {!locked && <span className="pointer-events-none absolute -right-2 -top-2 text-amber-300 opacity-0 transition-opacity duration-200 group-hover:opacity-100"><Sparkles className="h-4 w-4" /></span>}
                    </div>
                  </motion.button>
                </div>

                {/* Readable label + CTA over any background */}
                <div className="mt-3 w-40 rounded-2xl border border-white/50 bg-white/75 px-2 py-1.5 text-center shadow-md backdrop-blur">
                  <p className={cn('font-display text-sm font-bold', locked ? 'text-ink-400' : 'text-ink-900')}>{n.s.displayName}</p>
                  {locked ? (
                    <p className="mt-0.5 text-[11px] font-semibold text-ink-400">{itemsToUnlock(n.s) === 1 ? 'Almost there — 1 outfit' : `${itemsToUnlock(n.s)} outfits remaining`}</p>
                  ) : current ? (
                    <motion.button onClick={() => selectPersona(n.s.slug)} animate={{ scale: [1, 1.05, 1] }} transition={{ duration: 1.8, repeat: Infinity, ease: 'easeInOut' }} className="mt-1 inline-flex items-center gap-1 rounded-full bg-ink-900 px-3.5 py-1.5 text-xs font-bold text-white shadow-lg hover:bg-brand-600">
                      <Sparkles className="h-3.5 w-3.5" /> Style with {first}
                    </motion.button>
                  ) : (
                    <button onClick={() => selectPersona(n.s.slug)} className="mt-1 inline-flex items-center gap-1 rounded-full border border-brand-200 bg-white px-3 py-1 text-xs font-semibold text-brand-600 hover:border-brand-500">
                      Style with {first}
                    </button>
                  )}
                </div>
              </div>
            )
          })}

          {/* Final reward: the Ultimate Stylist palace */}
          <div className="absolute z-20 flex flex-col items-center" style={{ left: finalPos.x, top: finalPos.y, transform: 'translate(-50%, -50%)' }}>
            <motion.div animate={allUnlocked ? { y: [0, -8, 0], scale: [1, 1.04, 1] } : {}} transition={{ duration: 3.5, repeat: Infinity, ease: 'easeInOut' }} className={cn('relative flex h-32 w-32 items-center justify-center rounded-3xl border-4 text-6xl shadow-xl', allUnlocked ? 'border-amber-300' : 'border-white/60')} style={{ background: allUnlocked ? 'linear-gradient(135deg,#FFE7A3,#FFC46B)' : 'linear-gradient(135deg,#ece6ee,#ddd3dc)', boxShadow: allUnlocked ? '0 0 48px -6px rgba(255,196,107,0.85)' : undefined }}>
              <span className={cn(!allUnlocked && 'opacity-40 grayscale')}>🏰</span>
              {!allUnlocked && <span className="absolute -bottom-2 -right-2 flex h-9 w-9 items-center justify-center rounded-full border-2 border-white bg-ink-400 text-white"><Lock className="h-4 w-4" /></span>}
            </motion.div>
            <div className="mt-3 rounded-2xl border border-white/50 bg-white/80 px-3 py-1.5 text-center shadow-md backdrop-blur">
              <p className="flex items-center justify-center gap-1 font-display text-base font-black tracking-wide text-ink-900"><Gift className="h-4 w-4 text-brand-500" /> Ultimate Stylist</p>
              <p className={cn('text-[11px] font-semibold', allUnlocked ? 'text-amber-600' : 'text-ink-500')}>{allUnlocked ? 'Unlocked — the crown is yours 👑' : 'Unlock every stylist to reveal'}</p>
            </div>
          </div>

          {/* Golden halo on freshly-unlocked nodes */}
          <AnimatePresence>
            {celebrating.map((slug) => {
              const n = nodes.find((x) => x.s.slug === slug)
              if (!n) return null
              return <motion.div key={`halo-${slug}`} className="pointer-events-none absolute z-10 rounded-full" style={{ left: n.x, top: n.y, transform: 'translate(-50%,-50%)', width: R * 4, height: R * 4, background: 'radial-gradient(closest-side, rgba(255,196,107,0.55), transparent)' }} initial={{ scale: 0.3, opacity: 0.9 }} animate={{ scale: 1.4, opacity: 0 }} exit={{ opacity: 0 }} transition={{ duration: 1.4, ease: 'easeOut' }} />
            })}
          </AnimatePresence>
        </div>
      </div>
    </div>
  )
}
