import { useMemo } from 'react'
import { motion, useReducedMotion } from 'framer-motion'

// ChatAmbience — a quiet, luxurious backdrop for the stylist conversation:
// translucent blush-and-champagne bubbles drifting slowly upward with a gentle
// sway, over two large soft colour blooms. Purely decorative: pointer-
// transparent, GPU-only transforms, honours reduced motion, and sits behind
// the chat via the parent's isolated stacking context.
export function ChatAmbience() {
  const reduce = useReducedMotion() ?? false

  const bubbles = useMemo(
    () =>
      Array.from({ length: 14 }, (_, i) => {
        const gold = i % 3 === 0
        return {
          id: i,
          left: 4 + Math.random() * 92, // %
          size: 26 + Math.random() * 88,
          rise: 14 + Math.random() * 12, // s — slow, unhurried
          delay: Math.random() * -26, // negative → mid-flight on mount
          sway: 14 + Math.random() * 26,
          opacity: 0.38 + Math.random() * 0.3,
          gold,
        }
      }),
    [],
  )

  return (
    <div aria-hidden className="pointer-events-none absolute inset-0 -z-10 overflow-hidden">
      {/* soft colour blooms */}
      <div className="absolute -left-24 top-1/4 h-96 w-96 rounded-full opacity-50" style={{ background: 'radial-gradient(closest-side, rgba(255,63,108,0.18), transparent)' }} />
      <div className="absolute -right-28 bottom-1/4 h-[26rem] w-[26rem] rounded-full opacity-60" style={{ background: 'radial-gradient(closest-side, rgba(201,162,75,0.18), transparent)' }} />

      {/* floating bubbles */}
      {bubbles.map((b) =>
        reduce ? (
          <span
            key={b.id}
            className="absolute rounded-full"
            style={{
              left: `${b.left}%`,
              top: `${15 + ((b.id * 13) % 70)}%`,
              width: b.size,
              height: b.size,
              opacity: b.opacity * 0.7,
              background: b.gold
                ? 'radial-gradient(circle at 32% 30%, rgba(255,255,255,0.95), rgba(231,206,143,0.7) 55%, rgba(201,162,75,0.35))'
                : 'radial-gradient(circle at 32% 30%, rgba(255,255,255,0.95), rgba(255,158,197,0.65) 55%, rgba(255,63,108,0.3))',
              boxShadow: 'inset 0 0 0 1.5px rgba(255,255,255,0.85), 0 6px 22px -6px rgba(255,63,108,0.25)',
            }}
          />
        ) : (
          <motion.span
            key={b.id}
            className="absolute rounded-full"
            style={{
              left: `${b.left}%`,
              top: '104%',
              width: b.size,
              height: b.size,
              background: b.gold
                ? 'radial-gradient(circle at 32% 30%, rgba(255,255,255,0.95), rgba(231,206,143,0.7) 55%, rgba(201,162,75,0.35))'
                : 'radial-gradient(circle at 32% 30%, rgba(255,255,255,0.95), rgba(255,158,197,0.65) 55%, rgba(255,63,108,0.3))',
              boxShadow: 'inset 0 0 0 1.5px rgba(255,255,255,0.85), 0 6px 22px -6px rgba(255,63,108,0.25)',
            }}
            animate={{
              y: ['0vh', '-118vh'],
              x: [0, b.sway, -b.sway * 0.6, 0],
              opacity: [0, b.opacity, b.opacity, 0],
              scale: [0.85, 1, 1.06],
            }}
            transition={{
              duration: b.rise,
              delay: b.delay,
              repeat: Infinity,
              ease: 'linear',
              x: { duration: b.rise / 2, repeat: Infinity, repeatType: 'mirror', ease: 'easeInOut' },
              opacity: { duration: b.rise, delay: b.delay, repeat: Infinity, times: [0, 0.12, 0.85, 1], ease: 'linear' },
            }}
          />
        ),
      )}
    </div>
  )
}
