import { useMemo } from 'react'
import { motion } from 'framer-motion'

// Confetti is a one-shot celebration burst — no external library. It renders a
// fixed set of coloured particles that fan out and fall once, then the parent
// unmounts it (typically on a timer). Purely decorative and pointer-transparent
// so it never blocks the UI underneath.
// Pink-gold celebration palette — Myntra pink with champagne and rose golds.
const COLORS = ['#FF3F6C', '#FF9EC5', '#E7B84B', '#C9A24B', '#F4E4B8', '#FFC7DD', '#FFF3D6']

export function Confetti({ count = 44 }: { count?: number }) {
  const pieces = useMemo(
    () =>
      Array.from({ length: count }, (_, i) => ({
        id: i,
        left: Math.random() * 100,
        color: COLORS[i % COLORS.length],
        delay: Math.random() * 0.2,
        drift: (Math.random() - 0.5) * 160,
        rotate: Math.random() * 720 - 360,
        size: 6 + Math.random() * 8,
        duration: 1.6 + Math.random() * 1.2,
      })),
    [count],
  )

  return (
    <div className="pointer-events-none fixed inset-0 z-[100] overflow-hidden">
      {pieces.map((p) => (
        <motion.span
          key={p.id}
          initial={{ opacity: 1, y: -20, x: 0, rotate: 0 }}
          animate={{ opacity: [1, 1, 0], y: '100vh', x: p.drift, rotate: p.rotate }}
          transition={{ duration: p.duration, delay: p.delay, ease: 'easeOut' }}
          style={{
            position: 'absolute',
            top: '18%',
            left: `${p.left}%`,
            width: p.size,
            height: p.size * 0.6,
            borderRadius: 2,
            backgroundColor: p.color,
          }}
        />
      ))}
    </div>
  )
}
