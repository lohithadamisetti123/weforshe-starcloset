import { useEffect, useMemo, useRef, useState } from 'react'
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion'
import { useNavigate } from 'react-router-dom'
import { ArrowLeftRight } from 'lucide-react'
import { useStarCloset } from '@/context/StarClosetContext'
import { Confetti } from '@/components/starcloset/Confetti'
import { LuxIcon, type LuxIconName } from '@/components/starcloset/LuxuryIcons'
import { cn } from '@/lib/utils'
import type { StarStylist } from '@/types/starcloset'

// ProgressionMap — the StarCloset Fashion District. The stylist journey is no
// longer a game map: it is a vertical luxury boulevard where every stylist owns
// a designer boutique. Locked boutiques stand dark behind frosted glass; a
// fresh unlock plays a lights-on sequence (street lamps → facade colour → doors
// → window glow → shimmer) before the portrait appears; entering a boutique
// zooms through the doors into the stylist chat. Between boutiques, named
// plazas (Entrance Plaza, Designer Walk, Luxury Square…) give each block its
// own identity. Everything renders over the existing department environment
// art and derives purely from the existing unlock data — no logic, API,
// persona, chat or engine change. Vector art only, no emojis.

const BG: Record<string, string> = {
  Women: '/images/backgrounds/women.png',
  Men: '/images/backgrounds/men.png',
  Kids: '/images/backgrounds/kids.png',
}

const CHAMPAGNE = 'linear-gradient(135deg,#F4E4B8 0%,#C9A24B 55%,#A9822F 100%)'
const AURA = 'conic-gradient(from 0deg,#FF3F6C,#E7CE8F,#FF9EC5,#C9A24B,#FF3F6C)'

// Every celebrity's boutique is their Studio: Alia Studio, Sonam Studio…
const boutiqueName = (s: StarStylist) => `${s.displayName.split(' ')[0]} Studio`
// Named blocks between boutiques — each section of the boulevard has its own identity.
const PLAZAS: { name: string; icon: LuxIconName }[] = [
  { name: 'Entrance Plaza', icon: 'arch' },
  { name: 'Designer Walk', icon: 'boutique' },
  { name: 'Luxury Square', icon: 'fountain' },
  { name: 'Fashion Avenue', icon: 'lantern' },
  { name: 'Luxury Garden', icon: 'vase' },
]
// Street furniture cycled along the boulevard edges.
const FURNITURE: LuxIconName[] = ['lantern', 'bench', 'vase', 'fountain', 'platform', 'lantern', 'vase', 'bench']

function hash(s: string) {
  let h = 2166136261
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i)
    h = Math.imul(h, 16777619)
  }
  return h >>> 0
}

function initials(name: string) {
  const p = name.trim().split(/\s+/)
  return ((p[0]?.[0] ?? '') + (p.length > 1 ? p[p.length - 1][0] : p[0]?.[1] ?? '')).toUpperCase()
}

// Soft scroll-reveal — the camera easing each block into view.
function Reveal({ children, reduce, delay = 0 }: { children: React.ReactNode; reduce: boolean; delay?: number }) {
  if (reduce) return <>{children}</>
  return (
    <motion.div
      initial={{ opacity: 0, y: 34, scale: 0.94 }}
      whileInView={{ opacity: 1, y: 0, scale: 1 }}
      viewport={{ once: true, margin: '-10% 0px -10% 0px' }}
      transition={{ type: 'spring', stiffness: 110, damping: 20, delay }}
    >
      {children}
    </motion.div>
  )
}

// A dress-form mannequin silhouette for the display windows.
function Mannequin({ tint, lit }: { tint: string; lit: boolean }) {
  return (
    <svg viewBox="0 0 40 92" className="h-full w-auto" fill="none">
      <circle cx="20" cy="9" r="5.2" stroke={lit ? tint : 'rgba(255,255,255,0.45)'} strokeWidth="1.6" />
      <path d="M20 14.5v4" stroke={lit ? tint : 'rgba(255,255,255,0.45)'} strokeWidth="1.6" strokeLinecap="round" />
      <path
        d="M11.5 24c.4-3 3.6-5.5 8.5-5.5s8.1 2.5 8.5 5.5l1.4 12c.3 2.8-1 5.4-3.4 7.2L23 47v14h-6V47l-3.5-3.8c-2.4-1.8-3.7-4.4-3.4-7.2l1.4-12Z"
        fill={lit ? `${tint}2e` : 'rgba(255,255,255,0.07)'}
        stroke={lit ? tint : 'rgba(255,255,255,0.45)'}
        strokeWidth="1.6"
        strokeLinejoin="round"
      />
      <path d="M20 61v18" stroke={lit ? tint : 'rgba(255,255,255,0.45)'} strokeWidth="1.6" strokeLinecap="round" />
      <path d="M11 82c2.2-2 15.8-2 18 0" stroke={lit ? tint : 'rgba(255,255,255,0.45)'} strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  )
}

// One glass display window: spotlight, contents, reflections; dark when locked.
function DisplayWindow({
  arch, lit, locked, tint, children, reduce,
}: {
  arch: boolean; lit: boolean; locked: boolean; tint: string; children: React.ReactNode; reduce: boolean
}) {
  return (
    <div
      className={cn('relative flex-1 overflow-hidden border', arch ? 'rounded-t-[46%] rounded-b-lg' : 'rounded-lg')}
      style={{
        minHeight: 118,
        borderColor: lit ? 'rgba(231,206,143,0.65)' : 'rgba(255,255,255,0.28)',
        background: lit
          ? 'linear-gradient(180deg, rgba(255,236,200,0.28) 0%, rgba(255,222,168,0.16) 45%, rgba(20,16,28,0.35) 100%)'
          : 'linear-gradient(180deg, rgba(30,28,40,0.72), rgba(14,12,22,0.85))',
        boxShadow: lit ? 'inset 0 0 26px rgba(255,214,150,0.28), inset 0 0 0 1px rgba(255,255,255,0.12)' : 'inset 0 0 0 1px rgba(255,255,255,0.05)',
      }}
    >
      {/* interior spotlight cone */}
      <motion.div
        className="pointer-events-none absolute inset-x-0 top-0 h-full"
        style={{ background: 'radial-gradient(120% 90% at 50% -8%, rgba(255,230,180,0.55), transparent 62%)' }}
        animate={{ opacity: lit ? 1 : 0 }}
        transition={{ duration: 1.1, ease: 'easeOut' }}
      />
      {/* contents */}
      <div className="absolute inset-x-0 bottom-1.5 top-4 flex items-end justify-center pb-1" style={{ opacity: lit ? 1 : 0.55 }}>
        {children}
      </div>
      {/* moving glass reflection */}
      {lit && !reduce && (
        <motion.div
          className="pointer-events-none absolute inset-y-0 w-1/2 skew-x-[-18deg]"
          style={{ background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.16), transparent)' }}
          animate={{ x: ['-130%', '330%'] }}
          transition={{ duration: 6.5, repeat: Infinity, repeatDelay: 2.4, ease: 'easeInOut' }}
        />
      )}
      {/* frosted glass on locked windows */}
      {locked && <div className="absolute inset-0 backdrop-blur-[3px]" style={{ background: 'rgba(255,255,255,0.06)' }} />}
      <span className="sr-only">{tint}</span>
    </div>
  )
}

// ── The boutique storefront ──────────────────────────────────────────────────
function Boutique({
  s, index, locked, current, fresh, itemsAway, onEnter, reduce, nodeRef,
}: {
  s: StarStylist
  index: number
  locked: boolean
  current: boolean
  fresh: boolean
  itemsAway: number
  onEnter: () => void
  reduce: boolean
  nodeRef: (el: HTMLDivElement | null) => void
}) {
  const [imgOk, setImgOk] = useState(true)
  const h = hash(s.slug)
  const scallop = h % 2 === 0
  const arch = (h >> 1) % 2 === 0
  const house = boutiqueName(s)
  const lit = !locked
  // A fresh unlock staggers the lights-on sequence; otherwise everything is settled.
  const t = (base: number) => (fresh ? base : 0)

  return (
    <div ref={nodeRef} className="relative mx-auto w-full max-w-[560px]">
      {/* spotlight + breathing glow on the current boutique */}
      {current && (
        <>
          <div className="pointer-events-none absolute -top-16 left-1/2 h-40 w-64 -translate-x-1/2" style={{ background: 'radial-gradient(60% 100% at 50% 0%, rgba(255,241,214,0.34), transparent 75%)', animation: reduce ? undefined : 'sc-breathe-o 3.6s ease-in-out infinite' }} />
          {!reduce && Array.from({ length: 4 }).map((_, i) => (
            <motion.span key={i} className="pointer-events-none absolute rounded-full" style={{ left: `${18 + i * 21}%`, bottom: '30%', width: 4, height: 4, background: i % 2 ? '#FFC7DD' : '#FFE9B8', boxShadow: '0 0 8px rgba(255,220,190,0.9)' }} animate={{ y: [0, -90], opacity: [0, 0.9, 0] }} transition={{ duration: 3.4 + i * 0.7, delay: i * 0.8, repeat: Infinity, ease: 'easeOut' }} />
          ))}
        </>
      )}

      <motion.div
        className="relative overflow-hidden rounded-[26px] border backdrop-blur-xl"
        initial={false}
        animate={{
          filter: locked ? 'grayscale(0.9) brightness(0.82)' : 'grayscale(0) brightness(1)',
          borderColor: current ? 'rgba(231,206,143,0.9)' : lit ? 'rgba(255,255,255,0.45)' : 'rgba(255,255,255,0.22)',
        }}
        transition={{ duration: fresh ? 1.3 : 0.5, delay: t(0.3), ease: 'easeOut' }}
        style={{
          background: lit
            ? `linear-gradient(180deg, rgba(252,248,242,0.9) 0%, rgba(244,236,228,0.86) 58%, rgba(30,24,38,0.55) 100%)`
            : 'linear-gradient(180deg, rgba(66,62,76,0.55), rgba(24,22,34,0.72))',
          boxShadow: current
            ? '0 34px 80px -30px rgba(255,63,108,0.45), 0 18px 50px -22px rgba(11,13,20,0.7), inset 0 0 0 1px rgba(255,255,255,0.25)'
            : '0 24px 60px -28px rgba(11,13,20,0.75), inset 0 0 0 1px rgba(255,255,255,0.14)',
        }}
        whileHover={!locked ? { y: -6 } : {}}
      >
        {/* ── Signage ── */}
        <div className="relative flex flex-col items-center px-6 pt-5">
          {current && !reduce && (
            <motion.span className="pointer-events-none absolute inset-x-10 top-0 h-px" style={{ background: AURA, backgroundSize: '300% 100%' }} animate={{ backgroundPositionX: ['0%', '300%'] }} transition={{ duration: 6, repeat: Infinity, ease: 'linear' }} />
          )}
          <p className="text-[9px] font-semibold uppercase tracking-[0.3em]" style={{ color: lit ? '#A9822F' : 'rgba(255,255,255,0.55)' }}>Maison {String(index + 1).padStart(2, '0')}</p>
          <p className={cn('mt-0.5 font-display text-[21px] font-semibold tracking-tight', lit ? 'text-ink-900' : 'text-white/80')}>{house}</p>
          <p className={cn('text-[10.5px] font-medium', lit ? 'text-ink-500' : 'text-white/45')}>by {s.displayName}</p>
        </div>

        {/* ── Awning ── */}
        <div className="mt-3 flex justify-center gap-0 px-8" aria-hidden>
          {Array.from({ length: 7 }).map((_, i) => (
            <span
              key={i}
              className={cn('h-3.5 flex-1', scallop ? 'rounded-b-full' : 'rounded-b-sm')}
              style={{ background: i % 2 === 0 ? (lit ? `linear-gradient(180deg, ${s.accentFrom}, ${s.accentTo})` : 'rgba(120,118,132,0.7)') : lit ? 'rgba(252,248,242,0.95)' : 'rgba(70,68,82,0.7)', boxShadow: 'inset 0 -2px 3px rgba(0,0,0,0.18)' }}
            />
          ))}
        </div>

        {/* ── Facade: windows + door ── */}
        <div className="mt-2 flex items-stretch gap-2.5 px-5 pb-2" style={{ minHeight: 208 }}>
          <DisplayWindow arch={arch} lit={lit} locked={locked} tint={s.accentFrom} reduce={reduce}>
            <div className="h-[86%]"><Mannequin tint={s.accentFrom} lit={lit} /></div>
          </DisplayWindow>

          {/* Entrance */}
          <div className="relative w-[44%] shrink-0 overflow-hidden rounded-t-[40%] rounded-b-lg border sm:w-[38%]" style={{ borderColor: lit ? 'rgba(231,206,143,0.7)' : 'rgba(255,255,255,0.28)', background: lit ? 'linear-gradient(180deg, rgba(255,240,210,0.4), rgba(28,22,36,0.5))' : 'linear-gradient(180deg, rgba(34,30,46,0.85), rgba(16,14,24,0.92))', perspective: 300 }}>
            {/* warm interior + portrait behind the doors */}
            <motion.div className="absolute inset-0" style={{ background: 'radial-gradient(80% 70% at 50% 30%, rgba(255,214,150,0.5), transparent 75%)' }} animate={{ opacity: lit ? 1 : 0 }} transition={{ duration: 1, delay: t(0.9) }} />
            <motion.div className="absolute inset-x-0 top-3 flex justify-center" initial={false} animate={{ opacity: lit ? 1 : 0, scale: lit ? 1 : 0.8 }} transition={{ type: 'spring', stiffness: 200, damping: 18, delay: t(1.0) }}>
              <span className="relative block h-32 w-32 overflow-hidden rounded-full sm:h-36 sm:w-36" style={{ boxShadow: `0 0 0 3.5px rgba(255,255,255,0.95), 0 0 32px ${s.accentTo}88`, backgroundImage: `linear-gradient(135deg, ${s.accentFrom}, ${s.accentTo})` }}>
                {imgOk && s.portraitUrl ? (
                  <img src={s.portraitUrl} alt={s.displayName} loading="lazy" decoding="async" onError={() => setImgOk(false)} className="h-full w-full object-cover" />
                ) : (
                  <span className="flex h-full w-full items-center justify-center font-display text-2xl font-bold text-white">{initials(s.displayName)}</span>
                )}
              </span>
            </motion.div>
            {/* double doors — swing open when unlocked */}
            <motion.span className="absolute bottom-0 left-0 top-0 w-1/2 border-r" style={{ transformOrigin: 'left center', background: 'linear-gradient(90deg, rgba(60,54,72,0.94), rgba(84,76,96,0.9))', borderColor: 'rgba(231,206,143,0.4)' }} initial={false} animate={{ rotateY: lit ? -68 : 0 }} transition={{ duration: 1.1, delay: t(0.8), ease: [0.32, 0.72, 0.3, 1] }}>
              <span className="absolute right-1 top-1/2 h-3 w-1 rounded-full" style={{ background: 'rgba(231,206,143,0.85)' }} />
            </motion.span>
            <motion.span className="absolute bottom-0 right-0 top-0 w-1/2 border-l" style={{ transformOrigin: 'right center', background: 'linear-gradient(270deg, rgba(60,54,72,0.94), rgba(84,76,96,0.9))', borderColor: 'rgba(231,206,143,0.4)' }} initial={false} animate={{ rotateY: lit ? 68 : 0 }} transition={{ duration: 1.1, delay: t(0.8), ease: [0.32, 0.72, 0.3, 1] }}>
              <span className="absolute left-1 top-1/2 h-3 w-1 rounded-full" style={{ background: 'rgba(231,206,143,0.85)' }} />
            </motion.span>
            {/* lock plate on closed doors */}
            {locked && (
              <span className="absolute left-1/2 top-1/2 grid h-9 w-9 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full border border-white/40 bg-ink-950/70 text-white backdrop-blur">
                <LuxIcon name="lock" className="h-4 w-4" strokeWidth={1.6} />
              </span>
            )}
          </div>

          <DisplayWindow arch={arch} lit={lit} locked={locked} tint={s.accentTo} reduce={reduce}>
            <div className="flex h-[62%] items-end gap-2 pb-1 text-current" style={{ color: lit ? s.accentTo : 'rgba(255,255,255,0.45)' }}>
              <LuxIcon name="handbag" className="h-9 w-9" strokeWidth={1.3} />
              <LuxIcon name="perfume" className="h-7 w-7 opacity-80" strokeWidth={1.3} />
            </div>
          </DisplayWindow>
        </div>

        {/* fresh-unlock shimmer across the whole glass facade */}
        {fresh && !reduce && (
          <motion.div className="pointer-events-none absolute inset-y-0 w-1/3 skew-x-[-18deg]" style={{ background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.4), transparent)' }} initial={{ x: '-150%' }} animate={{ x: '420%' }} transition={{ duration: 1.2, delay: 1.4, ease: 'easeInOut' }} />
        )}

        {/* ── Marble step + action ── */}
        <div className="relative flex flex-col items-center gap-1.5 px-6 pb-5 pt-1">
          <span aria-hidden className="h-1.5 w-2/3 rounded-full" style={{ background: lit ? 'linear-gradient(90deg, transparent, rgba(231,206,143,0.75), transparent)' : 'linear-gradient(90deg, transparent, rgba(255,255,255,0.25), transparent)' }} />
          {locked ? (
            <p className="mt-1 text-[11px] font-medium text-white/70">{itemsAway === 1 ? 'One look away from opening' : `${itemsAway} looks away from opening`}</p>
          ) : (
            <motion.button
              onClick={onEnter}
              initial={fresh ? { opacity: 0, y: 8 } : false}
              animate={{ opacity: 1, y: 0, ...(current && !reduce ? { scale: [1, 1.04, 1] } : {}) }}
              transition={{ opacity: { duration: 0.5, delay: t(1.8) }, y: { duration: 0.5, delay: t(1.8) }, scale: { duration: 2.2, repeat: Infinity, ease: 'easeInOut' } }}
              whileHover={{ scale: 1.06, y: -2 }}
              whileTap={{ scale: 0.95 }}
              className={cn('mt-1 inline-flex items-center gap-1.5 rounded-full px-5 py-2 text-[12px] font-semibold tracking-tight', current ? 'text-ink-900' : 'border border-ink-900/20 bg-white/85 text-ink-900 hover:border-ink-900')}
              style={current ? { background: CHAMPAGNE, boxShadow: '0 12px 30px -10px rgba(201,162,75,0.75), inset 0 0 0 1px rgba(255,255,255,0.5)' } : undefined}
            >
              <LuxIcon name="sparkle" className="h-3.5 w-3.5" strokeWidth={1.5} /> Enter
            </motion.button>
          )}
        </div>
      </motion.div>
    </div>
  )
}

// A named plaza block between boutiques.
function Plaza({ plaza, lampsLit, index, reduce }: { plaza: (typeof PLAZAS)[number]; lampsLit: boolean; index: number; reduce: boolean }) {
  const left = index % 2 === 0
  return (
    <div className="relative mx-auto flex w-full max-w-[560px] flex-col items-center py-2">
      {/* street furniture on alternating sides */}
      <div className={cn('pointer-events-none absolute top-1/2 hidden -translate-y-1/2 sm:block', left ? '-left-24' : '-right-24')} style={{ opacity: 0.5 }}>
        <LuxIcon name={FURNITURE[index % FURNITURE.length]} className="h-16 w-16 text-white" strokeWidth={1.05} style={{ filter: 'drop-shadow(0 10px 18px rgba(11,13,20,0.5))' }} />
      </div>
      <div className={cn('pointer-events-none absolute top-1/2 hidden -translate-y-1/2 sm:block', left ? '-right-20' : '-left-20')} style={{ opacity: 0.42 }}>
        <div className="relative">
          <LuxIcon name="lantern" className="h-12 w-12 text-white" strokeWidth={1.05} />
          <motion.span className="absolute left-1/2 top-[42%] h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full" style={{ background: 'radial-gradient(circle, rgba(255,226,166,0.95), transparent 70%)' }} animate={{ opacity: lampsLit ? (reduce ? 0.9 : [0.55, 1, 0.55]) : 0.08 }} transition={{ duration: 3.2, repeat: lampsLit && !reduce ? Infinity : 0, ease: 'easeInOut' }} />
        </div>
      </div>

      <div className="flex items-center gap-3">
        <span className="h-px w-10" style={{ background: 'linear-gradient(90deg,transparent,rgba(231,206,143,0.85))' }} />
        <span className="flex items-center gap-2 rounded-full border border-white/45 bg-white/70 px-3.5 py-1.5 backdrop-blur-xl" style={{ boxShadow: '0 10px 26px -14px rgba(11,13,20,0.6)' }}>
          <LuxIcon name={plaza.icon} className="h-3.5 w-3.5 text-[#A9822F]" strokeWidth={1.5} />
          <span className="text-[10px] font-bold uppercase tracking-[0.26em] text-ink-700">{plaza.name}</span>
        </span>
        <span className="h-px w-10" style={{ background: 'linear-gradient(90deg,rgba(231,206,143,0.85),transparent)' }} />
      </div>
    </div>
  )
}

// ── The Grand Fashion Palace — the district's iconic final destination ───────
// Not a stylist: the symbolic headquarters of StarCloset, waiting in soft
// darkness behind golden chains until every boutique on the boulevard is open.
function GrandPalace({
  open, fresh, deptLabel, onEnter, reduce, nodeRef,
}: {
  open: boolean; fresh: boolean; deptLabel: string; onEnter: () => void; reduce: boolean; nodeRef: (el: HTMLDivElement | null) => void
}) {
  const t = (base: number) => (fresh ? base : 0)
  const lit = open

  return (
    <div ref={nodeRef} className="relative mx-auto w-full max-w-[620px]">
      {/* curiosity glow behind the building while locked; triumphant bloom when open */}
      <div className="pointer-events-none absolute -inset-8 -z-10" style={{ background: lit ? 'radial-gradient(60% 55% at 50% 45%, rgba(244,214,140,0.4), transparent 75%)' : 'radial-gradient(55% 50% at 50% 45%, rgba(201,162,75,0.22), transparent 72%)', animation: reduce ? undefined : 'sc-breathe-o 4.4s ease-in-out infinite' }} />

      <motion.button
        onClick={() => open && onEnter()}
        disabled={!open}
        whileHover={open ? { y: -6 } : {}}
        whileTap={open ? { scale: 0.985 } : {}}
        className={cn('relative block w-full overflow-hidden rounded-[30px] border text-left backdrop-blur-xl', open ? 'cursor-pointer border-transparent' : 'cursor-not-allowed border-white/25')}
        initial={false}
        animate={{ filter: lit ? 'grayscale(0) brightness(1)' : 'grayscale(0.55) brightness(0.85)' }}
        transition={{ duration: fresh ? 1.4 : 0.5, delay: t(0.5), ease: 'easeOut' }}
        style={{ boxShadow: lit ? '0 40px 90px -30px rgba(201,162,75,0.6), inset 0 0 0 1px rgba(255,255,255,0.35)' : '0 26px 60px -28px rgba(11,13,20,0.8), inset 0 0 0 1px rgba(255,255,255,0.1)' }}
      >
        {/* ── Sunset sky + glass dome ── */}
        <div className="relative h-40 overflow-hidden" style={{ background: lit ? 'linear-gradient(180deg, #F6D8AE 0%, #EFB98C 55%, #E29D7E 100%)' : 'linear-gradient(180deg, #2E2940 0%, #241F35 60%, #1B1729 100%)' }}>
          {/* birds crossing the sky during the celebration */}
          {fresh && !reduce && Array.from({ length: 3 }).map((_, i) => (
            <motion.svg key={i} viewBox="0 0 20 8" className="absolute h-2.5 w-6" style={{ top: `${18 + i * 16}%` }} initial={{ x: '-3rem', opacity: 0 }} animate={{ x: '42rem', opacity: [0, 1, 1, 0], y: [0, -6, 2, -4] }} transition={{ duration: 5.2 + i, delay: 2.4 + i * 0.7, ease: 'easeInOut' }} fill="none" stroke={lit ? '#7A5540' : '#C9C4D8'} strokeWidth="1.4" strokeLinecap="round">
              <path d="M2 6 Q6 1.5 10 6 M10 6 Q14 1.5 18 6" />
            </motion.svg>
          ))}
          {/* glass dome */}
          <div className="absolute bottom-0 left-1/2 h-[74px] w-[150px] -translate-x-1/2 overflow-hidden rounded-t-full border border-b-0" style={{ borderColor: lit ? 'rgba(201,162,75,0.85)' : 'rgba(255,255,255,0.3)', background: lit ? 'linear-gradient(180deg, rgba(255,244,214,0.6), rgba(255,222,160,0.28))' : 'linear-gradient(180deg, rgba(255,255,255,0.1), rgba(255,255,255,0.03))' }}>
            {[18, 42, 66, 90, 114, 132].map((x) => (
              <span key={x} className="absolute bottom-0 top-0 w-px" style={{ left: x, background: lit ? 'rgba(201,162,75,0.55)' : 'rgba(255,255,255,0.2)' }} />
            ))}
            <span className="absolute inset-x-0 bottom-0 h-px" style={{ background: lit ? 'rgba(201,162,75,0.8)' : 'rgba(255,255,255,0.25)' }} />
          </div>
          <span className="absolute bottom-[72px] left-1/2 h-4 w-1 -translate-x-1/2 rounded-full" style={{ background: lit ? '#C9A24B' : 'rgba(255,255,255,0.35)' }} />
        </div>

        {/* ── Facade: banners · colonnade · golden gates ── */}
        <div className="relative px-7 pb-1 pt-4" style={{ background: lit ? 'linear-gradient(180deg, rgba(252,247,238,0.94), rgba(243,232,214,0.9))' : 'linear-gradient(180deg, rgba(56,52,70,0.72), rgba(30,27,42,0.85))' }}>
          <p className="mb-1 text-center text-[9px] font-semibold uppercase tracking-[0.34em]" style={{ color: lit ? '#A9822F' : 'rgba(255,255,255,0.5)' }}>StarCloset Headquarters</p>
          <p className={cn('text-center font-display text-2xl font-semibold tracking-tight', lit ? 'text-ink-900' : 'text-white/85')}>Grand Fashion Palace</p>

          <div className="mt-4 flex items-end justify-center gap-3">
            {/* left banner + columns */}
            <span aria-hidden className="hidden h-24 w-3.5 rounded-b-md sm:block" style={{ background: lit ? CHAMPAGNE : 'rgba(120,116,134,0.6)' }} />
            {[0, 1].map((c) => (
              <span key={`cl-${c}`} aria-hidden className="relative h-28 w-4 rounded-t-sm" style={{ background: lit ? 'linear-gradient(90deg,#EFE6D4,#FBF6EA,#E4D6BC)' : 'linear-gradient(90deg,#4A4560,#5C5674,#413C55)' }}>
                <span className="absolute -top-1 left-1/2 h-1.5 w-6 -translate-x-1/2 rounded-sm" style={{ background: lit ? '#D9C591' : '#5C5674' }} />
              </span>
            ))}

            {/* golden gates */}
            <div className="relative h-36 w-32 shrink-0 overflow-hidden rounded-t-[52%]" style={{ perspective: 340, background: lit ? 'radial-gradient(85% 75% at 50% 25%, rgba(255,224,158,0.9), rgba(60,44,28,0.55))' : 'radial-gradient(85% 75% at 50% 25%, rgba(64,58,82,0.9), rgba(18,16,28,0.95))', boxShadow: lit ? 'inset 0 0 34px rgba(255,206,120,0.65)' : 'inset 0 0 0 1px rgba(255,255,255,0.08)' }}>
            {/* warm light spilling out */}
              <motion.span className="absolute inset-0" style={{ background: 'radial-gradient(70% 60% at 50% 45%, rgba(255,232,178,0.85), transparent 78%)' }} initial={false} animate={{ opacity: lit ? 1 : 0 }} transition={{ duration: 1.2, delay: t(1.6) }} />
              {/* gate panels */}
              <motion.span className="absolute bottom-0 left-0 top-0 w-1/2" style={{ transformOrigin: 'left center', background: 'linear-gradient(90deg,#8A6A34,#C9A24B 60%,#A9822F)', boxShadow: 'inset 0 0 0 1px rgba(255,240,200,0.4)' }} initial={false} animate={{ rotateY: lit ? -74 : 0 }} transition={{ duration: 1.6, delay: t(1.2), ease: [0.3, 0.7, 0.25, 1] }}>
                {[8, 18, 28].map((y) => <span key={y} className="absolute left-1/4 h-px w-1/2" style={{ top: `${y * 2.5}%`, background: 'rgba(255,240,200,0.5)' }} />)}
              </motion.span>
              <motion.span className="absolute bottom-0 right-0 top-0 w-1/2" style={{ transformOrigin: 'right center', background: 'linear-gradient(270deg,#8A6A34,#C9A24B 60%,#A9822F)', boxShadow: 'inset 0 0 0 1px rgba(255,240,200,0.4)' }} initial={false} animate={{ rotateY: lit ? 74 : 0 }} transition={{ duration: 1.6, delay: t(1.2), ease: [0.3, 0.7, 0.25, 1] }}>
                {[8, 18, 28].map((y) => <span key={y} className="absolute left-1/4 h-px w-1/2" style={{ top: `${y * 2.5}%`, background: 'rgba(255,240,200,0.5)' }} />)}
              </motion.span>
              {/* golden chains + lock while closed */}
              {!lit && (
                <>
                  <span className="absolute left-1/2 top-1/2 h-1 w-[130%] -translate-x-1/2 -translate-y-1/2 rotate-[16deg] rounded-full" style={{ background: 'repeating-linear-gradient(90deg, rgba(201,162,75,0.9) 0 6px, rgba(140,110,52,0.9) 6px 12px)' }} />
                  <span className="absolute left-1/2 top-1/2 h-1 w-[130%] -translate-x-1/2 -translate-y-1/2 rotate-[-16deg] rounded-full" style={{ background: 'repeating-linear-gradient(90deg, rgba(201,162,75,0.9) 0 6px, rgba(140,110,52,0.9) 6px 12px)' }} />
                  <span className="absolute left-1/2 top-1/2 grid h-10 w-10 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full border border-[#E7CE8F]/70 bg-ink-950/80 text-[#E7CE8F] backdrop-blur">
                    <LuxIcon name="lock" className="h-4.5 w-4.5" strokeWidth={1.6} />
                  </span>
                </>
              )}
              {/* celebration shimmer across the gates */}
              {fresh && !reduce && (
                <motion.span className="pointer-events-none absolute inset-y-0 w-1/3 skew-x-[-18deg]" style={{ background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.5), transparent)' }} initial={{ x: '-140%' }} animate={{ x: '400%' }} transition={{ duration: 1.1, delay: 2.6, ease: 'easeInOut' }} />
              )}
            </div>

            {[0, 1].map((c) => (
              <span key={`cr-${c}`} aria-hidden className="relative h-28 w-4 rounded-t-sm" style={{ background: lit ? 'linear-gradient(90deg,#EFE6D4,#FBF6EA,#E4D6BC)' : 'linear-gradient(90deg,#4A4560,#5C5674,#413C55)' }}>
                <span className="absolute -top-1 left-1/2 h-1.5 w-6 -translate-x-1/2 rounded-sm" style={{ background: lit ? '#D9C591' : '#5C5674' }} />
              </span>
            ))}
            <span aria-hidden className="hidden h-24 w-3.5 rounded-b-md sm:block" style={{ background: lit ? CHAMPAGNE : 'rgba(120,116,134,0.6)' }} />
          </div>

          {/* forecourt lamps — light one by one during the celebration */}
          <div className="mt-3 flex items-center justify-center gap-8">
            {Array.from({ length: 4 }).map((_, i) => (
              <span key={i} className="relative" style={{ color: lit ? '#A9822F' : 'rgba(255,255,255,0.4)' }}>
                <LuxIcon name="lantern" className="h-6 w-6" strokeWidth={1.3} />
                <motion.span className="absolute left-1/2 top-[40%] h-2.5 w-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full" style={{ background: 'radial-gradient(circle, rgba(255,226,166,1), transparent 72%)' }} initial={false} animate={{ opacity: lit ? 1 : 0.06 }} transition={{ duration: 0.5, delay: t(0.4 + i * 0.35) }} />
              </span>
            ))}
          </div>
        </div>

        {/* ── Marble staircase ── */}
        <div aria-hidden>
          {[0.62, 0.8, 1].map((w, i) => (
            <div key={i} className="mx-auto h-2.5" style={{ width: `${w * 100}%`, background: lit ? 'linear-gradient(180deg,#F7F0E2,#E2D5BC)' : 'linear-gradient(180deg,#4E4964,#37334C)', borderRadius: 3, boxShadow: 'inset 0 -1.5px 2px rgba(0,0,0,0.18)' }} />
          ))}
        </div>

        {/* ── Forecourt: gardens · fountain · reflective marble ── */}
        <div className="relative flex items-end justify-center gap-9 px-6 pb-5 pt-3" style={{ background: lit ? 'linear-gradient(180deg, rgba(247,240,226,0.92), rgba(214,200,176,0.85))' : 'linear-gradient(180deg, rgba(42,38,56,0.85), rgba(22,20,34,0.92))' }}>
          <span style={{ color: lit ? '#B98A5E' : 'rgba(255,255,255,0.35)' }}><LuxIcon name="vase" className="h-9 w-9" strokeWidth={1.25} /></span>
          <span className="relative" style={{ color: lit ? '#A9822F' : 'rgba(255,255,255,0.4)' }}>
            <LuxIcon name="fountain" className="h-12 w-12" strokeWidth={1.25} />
            {/* fountain water — flows once the palace is open */}
            {lit && !reduce && Array.from({ length: 3 }).map((_, i) => (
              <motion.span key={i} className="absolute left-1/2 top-1 h-1.5 w-1.5 -translate-x-1/2 rounded-full" style={{ background: 'rgba(190,225,255,0.9)', boxShadow: '0 0 6px rgba(190,225,255,0.8)' }} animate={{ y: [0, -10, 14], x: [(i - 1) * 2, (i - 1) * 7, (i - 1) * 10], opacity: [0, 1, 0] }} transition={{ duration: 1.6, delay: t(2.8) + i * 0.45, repeat: Infinity, ease: 'easeOut' }} />
            ))}
          </span>
          <span style={{ color: lit ? '#B98A5E' : 'rgba(255,255,255,0.35)' }}><LuxIcon name="vase" className="h-9 w-9" strokeWidth={1.25} /></span>
          {/* reflective marble sheen */}
          <span aria-hidden className="pointer-events-none absolute inset-x-10 bottom-1 h-6 rounded-[50%]" style={{ background: lit ? 'radial-gradient(50% 100% at 50% 0%, rgba(255,240,200,0.4), transparent 80%)' : 'radial-gradient(50% 100% at 50% 0%, rgba(255,255,255,0.08), transparent 80%)' }} />
        </div>

        {/* frosted veil while locked */}
        {!lit && <span className="pointer-events-none absolute inset-0 backdrop-blur-[2px]" style={{ background: 'rgba(20,18,30,0.12)' }} />}

        {/* status line */}
        <div className="absolute inset-x-0 bottom-0 flex justify-center pb-1.5">
          <span className={cn('rounded-full px-3 py-0.5 text-[9.5px] font-semibold uppercase tracking-[0.22em] backdrop-blur', lit ? 'bg-white/70 text-[#A9822F]' : 'bg-ink-950/55 text-white/65')}>
            {lit ? 'Tap to enter' : `Open all boutiques of the ${deptLabel} district`}
          </span>
        </div>
      </motion.button>
    </div>
  )
}

// ── Completion experience shown after entering the palace ────────────────────
function CompletionScreen({
  deptLabel, total, others, onExplore, onHome, onClose, reduce,
}: {
  deptLabel: string
  total: number
  others: { key: string; label: string }[]
  onExplore: (key: string) => void
  onHome: () => void
  onClose: () => void
  reduce: boolean
}) {
  const particles = useMemo(() => Array.from({ length: 14 }, (_, i) => ({ id: i, left: 4 + Math.random() * 92, size: 5 + Math.random() * 9, dur: 7 + Math.random() * 8, delay: Math.random() * -12 })), [])
  return (
    <motion.div className="fixed inset-0 z-[90] overflow-y-auto" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} transition={{ duration: 0.4 }} style={{ background: 'radial-gradient(120% 110% at 50% 0%, #FDF8EE 0%, #F6E8C8 55%, #EBD3A4 100%)' }}>
      <Confetti count={80} />
      {/* rising gold particles */}
      {!reduce && particles.map((p) => (
        <motion.span key={p.id} className="pointer-events-none fixed rounded-full" style={{ left: `${p.left}%`, bottom: -20, width: p.size, height: p.size, background: 'radial-gradient(circle at 35% 30%, #FFF3D6, #C9A24B)', boxShadow: '0 0 10px rgba(201,162,75,0.6)' }} animate={{ y: ['0vh', '-112vh'], opacity: [0, 0.85, 0] }} transition={{ duration: p.dur, delay: p.delay, repeat: Infinity, ease: 'linear' }} />
      ))}

      <div className="relative z-10 mx-auto flex min-h-full w-full max-w-md flex-col items-center justify-center px-5 py-14">
        <motion.div className="w-full overflow-hidden rounded-[28px] border border-white/70 bg-white/85 px-7 py-8 text-center backdrop-blur-2xl" initial={{ opacity: 0, y: 30, scale: 0.94 }} animate={{ opacity: 1, y: 0, scale: 1 }} transition={{ type: 'spring', stiffness: 120, damping: 18, delay: 0.15 }} style={{ boxShadow: '0 40px 90px -30px rgba(169,130,47,0.55), inset 0 0 0 1px rgba(255,255,255,0.6)' }}>
          {/* animated trophy */}
          <motion.span className="relative mx-auto grid h-20 w-20 place-items-center rounded-full" style={{ background: CHAMPAGNE, boxShadow: '0 18px 44px -14px rgba(169,130,47,0.8), inset 0 0 0 2px rgba(255,255,255,0.55)' }} initial={{ scale: 0, rotate: -14 }} animate={{ scale: 1, rotate: 0 }} transition={{ type: 'spring', stiffness: 200, damping: 13, delay: 0.35 }}>
            <motion.span animate={reduce ? {} : { rotate: [-3, 3, -3] }} transition={{ duration: 4.2, repeat: Infinity, ease: 'easeInOut' }}>
              <LuxIcon name="laurel" className="h-11 w-11 text-white" strokeWidth={1.3} />
            </motion.span>
            {!reduce && <motion.span className="absolute inset-0 rounded-full" style={{ boxShadow: '0 0 0 3px rgba(231,206,143,0.6)' }} animate={{ scale: [1, 1.35], opacity: [0.7, 0] }} transition={{ duration: 2, repeat: Infinity, ease: 'easeOut' }} />}
          </motion.span>

          <motion.p className="mt-4 font-display text-[26px] font-semibold tracking-tight text-ink-900" initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.5 }}>Congratulations</motion.p>
          <motion.p className="mt-1 text-[13px] font-medium text-ink-700" initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.62 }}>
            You have completed the {deptLabel} Fashion District.
          </motion.p>

          <motion.div className="mt-4 flex flex-wrap items-center justify-center gap-2" initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.75 }}>
            <span className="flex items-center gap-1.5 rounded-full border border-[#C9A24B]/45 bg-[#FBF3DD] px-3 py-1 text-[10px] font-bold uppercase tracking-[0.14em] text-[#A9822F]">
              <LuxIcon name="seal" className="h-3.5 w-3.5" strokeWidth={1.6} /> All Stylists Unlocked
            </span>
            <span className="flex items-center gap-1.5 rounded-full px-3 py-1 text-[10px] font-bold uppercase tracking-[0.14em] text-ink-900" style={{ background: CHAMPAGNE }}>
              <LuxIcon name="medal" className="h-3.5 w-3.5" strokeWidth={1.6} /> District Master
            </span>
          </motion.div>

          <motion.div className="mt-4 flex items-center justify-center gap-2" initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.85 }}>
            <span className="text-[11px] font-semibold text-ink-500">{total} / {total} Stylists</span>
            <span className="flex items-center gap-1">
              {Array.from({ length: total }).map((_, i) => (
                <motion.span key={i} initial={{ scale: 0 }} animate={{ scale: 1 }} transition={{ type: 'spring', stiffness: 260, damping: 14, delay: 0.9 + i * 0.09 }}>
                  <LuxIcon name="boutique" className="h-3.5 w-3.5 text-[#C9A24B]" strokeWidth={1.5} />
                </motion.span>
              ))}
            </span>
          </motion.div>

          {/* next actions */}
          <motion.div className="mt-6 flex flex-col gap-2.5" initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 1.05 }}>
            {others.map((d) => (
              <button key={d.key} onClick={() => onExplore(d.key)} className="group flex items-center justify-between rounded-2xl border border-[#C9A24B]/35 bg-white/80 px-4 py-3 text-left transition-all hover:-translate-y-0.5 hover:border-[#C9A24B]" style={{ boxShadow: '0 10px 26px -16px rgba(169,130,47,0.5)' }}>
                <span className="flex items-center gap-3">
                  <span className="grid h-9 w-9 place-items-center rounded-full font-display text-sm font-semibold text-ink-900" style={{ background: 'linear-gradient(135deg,#FBF3DD,#EBD9A6)', boxShadow: 'inset 0 0 0 1px rgba(201,162,75,0.5)' }}>{d.label[0]}</span>
                  <span>
                    <span className="block text-[13px] font-semibold tracking-tight text-ink-900">Explore {d.label}’s Fashion District</span>
                    <span className="block text-[10.5px] font-medium text-ink-500">A new boulevard of boutiques awaits</span>
                  </span>
                </span>
                <span className="text-[#A9822F] transition-transform group-hover:translate-x-1"><LuxIcon name="sparkle" className="h-4 w-4" strokeWidth={1.5} /></span>
              </button>
            ))}
            <button onClick={onHome} className="rounded-2xl border border-ink-900/15 bg-white/70 px-4 py-2.5 text-[12.5px] font-semibold tracking-tight text-ink-700 transition-colors hover:border-ink-900 hover:text-ink-900">
              Return to Home
            </button>
            <button onClick={onClose} className="mt-0.5 text-[10.5px] font-semibold uppercase tracking-[0.18em] text-ink-400 transition-colors hover:text-ink-700">
              Stay on the boulevard
            </button>
          </motion.div>
        </motion.div>
      </div>
    </motion.div>
  )
}

export function ProgressionMap() {
  const { currentDept, departments, unlockScore, isUnlocked, itemsToUnlock, selectedSlug, selectPersona, selectDepartment, clearDepartment, nextLocked } = useStarCloset()
  const navigate = useNavigate()

  const reduce = useReducedMotion() ?? false
  const [confetti, setConfetti] = useState(false)
  const [celebrating, setCelebrating] = useState<string[]>([])
  const [entering, setEntering] = useState<StarStylist | null>(null)
  const [palaceZoom, setPalaceZoom] = useState(false)
  const [showCompletion, setShowCompletion] = useState(false)
  const nodeRefs = useRef<Record<string, HTMLDivElement | null>>({})
  const palaceRef = useRef<HTMLDivElement | null>(null)

  const stylists = useMemo(() => currentDept?.stylists ?? [], [currentDept])

  const unlockedList = stylists.filter((s) => isUnlocked(s))
  const allUnlocked = unlockedList.length === stylists.length && stylists.length > 0
  const featured =
    (selectedSlug && stylists.find((s) => s.slug === selectedSlug && isUnlocked(s))) ||
    unlockedList[unlockedList.length - 1] ||
    stylists[0]

  const total = stylists.length
  const unlockedCount = unlockedList.length
  const maxThreshold = Math.max(1, ...stylists.map((s) => s.unlockAtItems))
  const pct = allUnlocked ? 100 : Math.min(100, Math.round((unlockScore / maxThreshold) * 100))

  // ── Camera: focus the active boutique on open / department change ────────────
  useEffect(() => {
    const el = featured ? nodeRefs.current[featured.slug] : null
    if (el) {
      const t = setTimeout(() => el.scrollIntoView({ behavior: 'smooth', block: 'center' }), 320)
      return () => clearTimeout(t)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [featured?.slug, currentDept?.key])

  // ── Camera pan + lights-on celebration on a fresh unlock ─────────────────────
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
      setTimeout(() => nodeRefs.current[fresh[fresh.length - 1]]?.scrollIntoView({ behavior: 'smooth', block: 'center' }), 360)
      // Full completion: after the boutique moment, pan the camera down the
      // boulevard to the Grand Fashion Palace for its lights-on cinematic.
      const completedAll = now.length === stylists.length && stylists.length > 0
      const panT = completedAll ? setTimeout(() => palaceRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' }), 1500) : null
      const t = setTimeout(() => { setConfetti(false); setCelebrating([]) }, completedAll ? 6500 : 4200)
      sessionStorage.setItem(key, JSON.stringify(now))
      return () => { clearTimeout(t); if (panT) clearTimeout(panT) }
    }
    sessionStorage.setItem(key, JSON.stringify(now))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [unlockScore, currentDept?.key])

  // ── Entering a boutique: zoom through the doors, then open the chat ──────────
  const enterBoutique = (s: StarStylist) => {
    if (entering) return
    setEntering(s)
    setTimeout(() => selectPersona(s.slug), reduce ? 80 : 760)
  }

  if (!currentDept) return null
  const bg = BG[currentDept.key] ?? BG.Women
  const monogram = (currentDept.label?.[0] ?? 'S').toUpperCase()

  return (
    <div className="relative min-h-[calc(100vh-5rem)]">
      <style>{`
        @keyframes sc-mote {0%,100%{opacity:.15;transform:translateY(0) scale(.8)}50%{opacity:.65;transform:translateY(-10px) scale(1.1)}}
        @keyframes sc-sheen {0%{background-position:-180% 0}100%{background-position:180% 0}}
        @keyframes sc-breathe-o {0%,100%{opacity:.5}50%{opacity:1}}
        @keyframes sc-sun {0%,100%{opacity:.24}50%{opacity:.44}}
      `}</style>

      {confetti && <Confetti count={64} />}

      {/* ── Environment: department art + editorial scrims + animated sunlight ── */}
      <div aria-hidden className="pointer-events-none sticky top-0 z-0 -mb-[100vh] h-screen w-full overflow-hidden">
        <img src={bg} alt="" className="h-full w-full object-cover object-center" />
        <div className="absolute inset-0 bg-gradient-to-b from-ink-950/55 via-ink-950/32 to-ink-950/62" />
        <div className="absolute inset-0 bg-[radial-gradient(120%_80%_at_50%_-10%,rgba(255,63,108,0.13),transparent_58%)]" />
        <div className="absolute inset-0 bg-[radial-gradient(120%_90%_at_50%_120%,rgba(11,13,20,0.6),transparent_55%)]" />
        {/* slow sunlight beam sweeping the district */}
        <div className="absolute -left-1/4 -top-1/4 h-[150%] w-1/2 rotate-[24deg]" style={{ background: 'linear-gradient(90deg, transparent, rgba(255,238,204,0.16), transparent)', animation: reduce ? undefined : 'sc-sun 9s ease-in-out infinite' }} />
        {Array.from({ length: 12 }).map((_, i) => (
          <span key={i} className="absolute rounded-full" style={{ left: `${(i * 61) % 100}%`, top: `${(i * 37) % 100}%`, width: 2 + (i % 3), height: 2 + (i % 3), background: 'rgba(255,236,214,0.9)', boxShadow: '0 0 8px rgba(255,214,180,0.8)', animation: reduce ? undefined : `sc-mote ${5 + (i % 5) * 0.8}s ease-in-out ${(i % 6) * 0.4}s infinite` }} />
        ))}
      </div>

      <div className="relative z-10">
        {/* ── District card — glass, pinned below the navbar ─────────────────── */}
        <div className="sticky top-[84px] z-40 px-4">
          <div className="mx-auto max-w-md overflow-hidden rounded-[22px] border border-white/60 bg-white/75 p-4 backdrop-blur-2xl" style={{ boxShadow: '0 20px 60px -20px rgba(11,13,20,0.6), inset 0 0 0 1px rgba(255,255,255,0.5)' }}>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className="relative grid h-11 w-11 place-items-center rounded-full font-display text-lg font-semibold text-ink-900" style={{ background: 'linear-gradient(135deg,#FBF3DD,#EBD9A6)', boxShadow: 'inset 0 0 0 1px rgba(201,162,75,0.55), 0 6px 16px -8px rgba(201,162,75,0.6)' }}>
                  {monogram}
                  <span className="pointer-events-none absolute inset-0 rounded-full" style={{ boxShadow: 'inset 0 0 0 3px rgba(255,255,255,0.6)' }} />
                </span>
                <div>
                  <p className="text-[9px] font-semibold uppercase tracking-[0.22em] text-ink-500">StarCloset · Fashion District</p>
                  <p className="font-display text-[15px] font-semibold tracking-tight text-ink-900">{currentDept.label}’s Boulevard</p>
                </div>
              </div>
              <button onClick={clearDepartment} className="flex items-center gap-1.5 rounded-full border border-line bg-white/80 px-3 py-1.5 text-xs font-semibold text-ink-700 backdrop-blur transition-colors hover:border-ink-900 hover:text-ink-900">
                <ArrowLeftRight className="h-3 w-3" /> Switch
              </button>
            </div>

            <div className="mt-3 flex items-center justify-between">
              <span className="rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.14em] text-ink-900" style={{ background: 'linear-gradient(135deg,#F4E4B8,#D9BE7A)', boxShadow: 'inset 0 0 0 1px rgba(201,162,75,0.5)' }}>
                {unlockedCount} of {total} boutiques open
              </span>
              <span className="flex items-center gap-1">
                {Array.from({ length: total }).map((_, i) => (
                  <LuxIcon key={i} name="boutique" className="h-3.5 w-3.5" strokeWidth={1.5} style={{ color: i < unlockedCount ? '#C9A24B' : '#d1d5db' }} />
                ))}
              </span>
            </div>

            <div className="mt-3 h-2 w-full overflow-hidden rounded-full bg-ink-900/8" style={{ boxShadow: 'inset 0 0 0 1px rgba(17,24,39,0.06)' }}>
              <motion.div className="h-full rounded-full" style={{ background: 'linear-gradient(90deg,#FF9EC5,#FF3F6C 45%,#E7B84B)', backgroundSize: '200% 100%', animation: reduce ? undefined : 'sc-sheen 2.6s linear infinite' }} initial={{ width: 0 }} animate={{ width: `${pct}%` }} transition={{ type: 'spring', stiffness: 90, damping: 20 }} />
            </div>
            <div className="mt-2 flex items-center justify-between text-[11px]">
              <span className="font-semibold tracking-tight text-brand-600">{pct}% · XP</span>
              {nextLocked ? (
                <span className="text-ink-500">Next opening — <span className="font-semibold text-ink-900">{nextLocked.displayName.split(' ')[0]}’s boutique</span> · {itemsToUnlock(nextLocked)} away</span>
              ) : (
                <span className="font-semibold" style={{ color: '#A9822F' }}>Every boutique is open</span>
              )}
            </div>
          </div>
        </div>

        {/* ── The boulevard ──────────────────────────────────────────────────── */}
        <div className="relative mx-auto mt-10 flex w-full max-w-3xl flex-col items-stretch gap-9 px-4 pb-40">
          {/* marble walkway down the centre */}
          <div aria-hidden className="pointer-events-none absolute inset-y-0 left-1/2 -z-10 w-14 -translate-x-1/2 sm:w-20" style={{ background: 'linear-gradient(180deg, rgba(255,255,255,0.14), rgba(255,255,255,0.07))', boxShadow: 'inset 0 0 0 1px rgba(255,255,255,0.16), 0 0 40px rgba(255,255,255,0.06)', borderRadius: 999, maskImage: 'linear-gradient(180deg, transparent, black 4%, black 96%, transparent)', WebkitMaskImage: 'linear-gradient(180deg, transparent, black 4%, black 96%, transparent)' }}>
            {Array.from({ length: 26 }).map((_, i) => (
              <span key={i} className="absolute left-1/2 h-px w-7 -translate-x-1/2 sm:w-10" style={{ top: `${(i + 0.5) * 3.8}%`, background: 'rgba(255,255,255,0.2)' }} />
            ))}
          </div>

          {stylists.map((s, i) => {
            const locked = !isUnlocked(s)
            const current = !locked && featured && s.slug === featured.slug
            const fresh = celebrating.includes(s.slug)
            const plaza = PLAZAS[Math.min(i, PLAZAS.length - 1)]
            // lamps in a plaza glow once the boutique after it is open (or opening)
            const lampsLit = !locked || fresh
            return (
              <div key={s.slug} className="flex flex-col gap-9">
                <Reveal reduce={reduce}>
                  <Plaza plaza={plaza} lampsLit={lampsLit} index={i} reduce={reduce} />
                </Reveal>
                <Reveal reduce={reduce} delay={0.06}>
                  <Boutique
                    s={s}
                    index={i}
                    locked={locked}
                    current={!!current}
                    fresh={fresh}
                    itemsAway={itemsToUnlock(s)}
                    onEnter={() => enterBoutique(s)}
                    reduce={reduce}
                    nodeRef={(el) => { nodeRefs.current[s.slug] = el }}
                  />
                </Reveal>
              </div>
            )
          })}

          {/* ── The Grand Fashion Palace — the district's final destination ── */}
          <Reveal reduce={reduce}>
            <div className="flex flex-col items-center gap-7">
              {/* grand divider */}
              <div className="flex w-full max-w-[560px] items-center gap-3 px-2" aria-hidden>
                <span className="h-px flex-1" style={{ background: 'linear-gradient(90deg,transparent,rgba(231,206,143,0.9))' }} />
                <span className="grid h-8 w-8 place-items-center rounded-full border border-[#E7CE8F]/60 bg-white/70 backdrop-blur">
                  <LuxIcon name="laurel" className="h-4 w-4 text-[#A9822F]" strokeWidth={1.5} />
                </span>
                <span className="h-px flex-1" style={{ background: 'linear-gradient(90deg,rgba(231,206,143,0.9),transparent)' }} />
              </div>
              <GrandPalace
                open={allUnlocked}
                fresh={allUnlocked && celebrating.length > 0}
                deptLabel={currentDept.label}
                onEnter={() => {
                  if (palaceZoom || showCompletion) return
                  setPalaceZoom(true)
                  setTimeout(() => { setShowCompletion(true); setPalaceZoom(false) }, reduce ? 80 : 820)
                }}
                reduce={reduce}
                nodeRef={(el) => { palaceRef.current = el }}
              />
            </div>
          </Reveal>
        </div>
      </div>

      {/* ── Entering transition: zoom through the doors into the chat ─────────── */}
      <AnimatePresence>
        {entering && (
          <motion.div
            className="fixed inset-0 z-[80] flex items-center justify-center"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.3 }}
            style={{ background: 'radial-gradient(120% 120% at 50% 50%, rgba(252,248,242,0.97), rgba(244,228,184,0.97))' }}
          >
            <motion.div
              className="flex flex-col items-center"
              initial={{ scale: 0.7, opacity: 0 }}
              animate={{ scale: reduce ? 1 : [0.7, 1.06, 1], opacity: 1 }}
              transition={{ duration: 0.6, ease: 'easeOut' }}
            >
              <span className="grid h-16 w-16 place-items-center rounded-full" style={{ background: CHAMPAGNE, boxShadow: '0 18px 44px -14px rgba(201,162,75,0.8)' }}>
                <LuxIcon name="boutique" className="h-8 w-8 text-white" strokeWidth={1.3} />
              </span>
              <p className="mt-3 font-display text-xl font-semibold tracking-tight text-ink-900">
                {boutiqueName(entering)}
              </p>
              <p className="text-[11px] font-medium uppercase tracking-[0.24em] text-ink-500">Welcome in</p>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* ── Palace entrance: cinematic zoom through the golden gates ──────────── */}
      <AnimatePresence>
        {palaceZoom && (
          <motion.div
            className="fixed inset-0 z-[85] flex items-center justify-center"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.35 }}
            style={{ background: 'radial-gradient(120% 120% at 50% 45%, rgba(255,244,214,0.98), rgba(235,211,164,0.98))' }}
          >
            <motion.div className="flex flex-col items-center" initial={{ scale: 0.65, opacity: 0 }} animate={{ scale: reduce ? 1 : [0.65, 1.08, 1], opacity: 1 }} transition={{ duration: 0.65, ease: 'easeOut' }}>
              <span className="grid h-18 w-18 place-items-center rounded-full p-4" style={{ background: CHAMPAGNE, boxShadow: '0 22px 50px -16px rgba(169,130,47,0.85)' }}>
                <LuxIcon name="laurel" className="h-9 w-9 text-white" strokeWidth={1.3} />
              </span>
              <p className="mt-3 font-display text-xl font-semibold tracking-tight text-ink-900">Grand Fashion Palace</p>
              <p className="text-[11px] font-medium uppercase tracking-[0.24em] text-ink-500">The doors open for you</p>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* ── Completion experience ─────────────────────────────────────────────── */}
      <AnimatePresence>
        {showCompletion && currentDept && (
          <CompletionScreen
            deptLabel={currentDept.label}
            total={total}
            others={departments.filter((d) => d.key !== currentDept.key).map((d) => ({ key: d.key, label: d.label }))}
            onExplore={(key) => { setShowCompletion(false); selectDepartment(key); window.scrollTo({ top: 0, behavior: 'auto' }) }}
            onHome={() => { setShowCompletion(false); navigate('/') }}
            onClose={() => setShowCompletion(false)}
            reduce={reduce}
          />
        )}
      </AnimatePresence>
    </div>
  )
}
