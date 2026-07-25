import type { CSSProperties } from 'react'

// LuxuryIcons — a cohesive set of minimalist, editorial line icons for the
// StarCloset journey. They replace every emoji in the experience with premium
// vector illustrations drawn on a single 24×24 grid, hairline strokes, round
// caps/joins, and `currentColor` so a parent can tint them champagne-gold,
// platinum, ink or crystal. Fashion-first, Dior/Chanel-minimal — never playful.

export type LuxIconName =
  | 'handbag'
  | 'perfume'
  | 'heel'
  | 'diamond'
  | 'crystal'
  | 'ribbon'
  | 'medal'
  | 'star'
  | 'pin'
  | 'boutique'
  | 'giftbox'
  | 'hanger'
  | 'sparkle'
  | 'lantern'
  | 'vase'
  | 'token'
  | 'seal'
  | 'lock'
  | 'arch'
  | 'runway'
  | 'gallery'
  | 'laurel'
  | 'bridge'
  | 'staircase'
  | 'fountain'
  | 'platform'
  | 'bench'

// Each entry is the icon's inner geometry; the wrapper supplies the <svg> frame,
// stroke colour (currentColor) and width so the whole set stays uniform.
const PATHS: Record<LuxIconName, React.ReactNode> = {
  // Structured top-handle handbag
  handbag: (
    <>
      <path d="M5.6 8.5h12.8l-1 10.2a1.2 1.2 0 0 1-1.2 1.1H7.8a1.2 1.2 0 0 1-1.2-1.1L5.6 8.5Z" />
      <path d="M8.7 8.5V7a3.3 3.3 0 0 1 6.6 0v1.5" />
      <path d="M12 12.4v2.1" />
    </>
  ),
  // Faceted perfume flacon
  perfume: (
    <>
      <rect x="9.9" y="3.6" width="4.2" height="3" rx="0.8" />
      <path d="M10.4 6.6v2.1M13.6 6.6v2.1" />
      <path d="M8.6 10.2c0-.9.6-1.5 1.6-1.5h3.6c1 0 1.6.6 1.6 1.5v8.2a1.6 1.6 0 0 1-1.6 1.6h-3.6a1.6 1.6 0 0 1-1.6-1.6v-8.2Z" />
      <path d="M11 13.2h2" />
    </>
  ),
  // Elegant stiletto pump
  heel: (
    <>
      <path d="M4.5 7.4v5.9c0 1.2.9 2.1 2.1 2.1h12.9c-3.1-.2-6.1-1.5-8.4-3.7-1.9-1.8-4.1-3.4-6.6-4.3Z" />
      <path d="M16.6 15.4 18 19" />
      <path d="M4.5 15.4h4.2" />
    </>
  ),
  // Brilliant-cut diamond
  diamond: (
    <>
      <path d="M12 20.5 4.5 9.2 8 4.2h8l3.5 5-7.5 11.3Z" />
      <path d="M4.5 9.2h15" />
      <path d="M8 4.2 12 9.2l4-5M9.4 9.2 12 20.5l2.6-11.3" />
    </>
  ),
  // Tall faceted crystal
  crystal: (
    <>
      <path d="M12 3.2 7.5 8 12 20.8 16.5 8 12 3.2Z" />
      <path d="M7.5 8h9M12 3.2v17.6" />
    </>
  ),
  // Award ribbon / rosette
  ribbon: (
    <>
      <circle cx="12" cy="8.4" r="5" />
      <path d="M9.2 12.7 7.4 20.4l4.6-2.8 4.6 2.8-1.8-7.7" />
      <path d="M12 6.1l.9 1.8 2 .3-1.45 1.4.35 2L12 11.9l-1.8.7.35-2L9.1 8.2l2-.3.9-1.8Z" />
    </>
  ),
  // Medal on twin ribbons
  medal: (
    <>
      <path d="M9.3 9.2 7.2 3.4M14.7 9.2l2.1-5.8" />
      <circle cx="12" cy="14.4" r="5" />
      <path d="M12 11.7l.8 1.6 1.8.26-1.3 1.27.31 1.79L12 16.75l-1.6.84.3-1.79-1.3-1.27 1.8-.26.8-1.6Z" />
    </>
  ),
  // Slim editorial star
  star: (
    <path d="M12 3.4l2.55 5.17 5.7.83-4.12 4.02.97 5.68L12 16.42l-5.07 2.66.97-5.68L3.75 9.4l5.7-.83L12 3.4Z" />
  ),
  // Brooch / fashion pin
  pin: (
    <>
      <circle cx="12" cy="8.6" r="4.9" />
      <circle cx="12" cy="8.6" r="1.7" />
      <path d="M12 13.5v6.8" />
    </>
  ),
  // Boutique storefront with scalloped awning
  boutique: (
    <>
      <path d="M4.4 9.2 5.6 5.8h12.8l1.2 3.4" />
      <path d="M4.4 9.2c1.3 1.5 2.6 1.5 3.9 0 1.3 1.5 2.6 1.5 3.9 0 1.3 1.5 2.6 1.5 3.9 0 1.3 1.5 2.6 1.5 3.5.4" />
      <path d="M5.3 9.8v9.6a.8.8 0 0 0 .8.8h11.8a.8.8 0 0 0 .8-.8V9.8" />
      <path d="M10 20.2v-4.4a2 2 0 0 1 4 0v4.4" />
    </>
  ),
  // Ribboned designer gift box
  giftbox: (
    <>
      <rect x="4.6" y="9.4" width="14.8" height="10.4" rx="1" />
      <path d="M4.6 13.2h14.8M12 9.4v10.4" />
      <path d="M12 9.4C10.3 9.4 8.3 8.6 8.3 7c0-1.1 1-1.6 1.9-1.1C11.3 6.6 12 8 12 9.4Zm0 0c1.7 0 3.7-.8 3.7-2.4 0-1.1-1-1.6-1.9-1.1C12.7 6.6 12 8 12 9.4Z" />
    </>
  ),
  // Wardrobe hanger
  hanger: (
    <>
      <path d="M12 6.6a1.9 1.9 0 1 1 1.35 1.82c-.5.17-.85.63-.85 1.16v.6" />
      <path d="M12 10.2 4 16.1c-.7.5-.34 1.6.5 1.6h15c.84 0 1.2-1.1.5-1.6L12 10.2Z" />
    </>
  ),
  // Four-point sparkle
  sparkle: (
    <path d="M12 3.2c.55 4.6 1.65 5.7 6.2 6.25-4.55.55-5.65 1.65-6.2 6.25-.55-4.6-1.65-5.7-6.2-6.25C10.35 8.9 11.45 7.8 12 3.2Z" />
  ),
  // Boutique lantern
  lantern: (
    <>
      <path d="M9.8 3.6h4.4M12 3.6v2.2" />
      <rect x="7.9" y="5.8" width="8.2" height="12" rx="3.2" />
      <path d="M8 17.6h8v1.1a2.1 2.1 0 0 1-2.1 2.1h-3.8A2.1 2.1 0 0 1 8 18.7v-1.1Z" />
      <path d="M10.1 6v11.4M13.9 6v11.4" />
    </>
  ),
  // Sculptural vase with a stem
  vase: (
    <>
      <path d="M9 9.4c-1.3 2.6-1.3 6 0 8.4a3.4 3.4 0 0 0 6 0c1.3-2.4 1.3-5.8 0-8.4Z" />
      <path d="M9 9.4h6" />
      <path d="M12 9.4V5.4c0-1.1.9-2 2.4-2.2" />
    </>
  ),
  // Faceted premium token / coin
  token: (
    <>
      <circle cx="12" cy="12" r="8" />
      <path d="M12 6.6 15 9.4 13.9 13 10.1 13 9 9.4 12 6.6Z" />
      <path d="M10.1 13 12 17.4 13.9 13" />
    </>
  ),
  // Certified rosette seal (verification badge)
  seal: (
    <>
      <path d="M12 2.8l1.9 1.35 2.32-.28.72 2.22 2.1 1.02-.5 2.29L20 11.9l-1.46 1.83.5 2.29-2.1 1.02-.72 2.22-2.32-.28L12 20.4l-1.9-1.35-2.32.28-.72-2.22-2.1-1.02.5-2.29L4 11.9l1.46-1.83-.5-2.29 2.1-1.02.72-2.22 2.32.28L12 2.8Z" />
      <path d="M9 12l2.1 2.1L15.2 10" />
    </>
  ),
  // Padlock
  lock: (
    <>
      <rect x="5.2" y="10.4" width="13.6" height="9.4" rx="2.2" />
      <path d="M8.3 10.4V7.6a3.7 3.7 0 0 1 7.4 0v2.8" />
      <path d="M12 14v2.2" />
    </>
  ),
  // Grand archway (milestone landmark)
  arch: (
    <>
      <path d="M4.4 20.4V9.2a7.6 7.6 0 0 1 15.2 0v11.2" />
      <path d="M3.4 20.4h17.2" />
      <path d="M8.2 20.4v-7a3.8 3.8 0 0 1 7.6 0v7" />
    </>
  ),
  // Runway strip (milestone)
  runway: (
    <>
      <path d="M9.4 3.8h5.2l2.6 16.4H6.8L9.4 3.8Z" />
      <path d="M12 4v16" />
      <path d="M8.6 9.6h6.8M7.9 14.6h8.2" />
    </>
  ),
  // Framed gallery portrait (milestone)
  gallery: (
    <>
      <rect x="4.2" y="4.2" width="15.6" height="15.6" rx="1.2" />
      <circle cx="9.2" cy="9.6" r="1.7" />
      <path d="M4.4 16.6 9 12l3.2 3.2L15.6 11l4.2 4.6" />
    </>
  ),
  // Arched luxury bridge (environmental landmark)
  bridge: (
    <>
      <path d="M3 13.4h18" />
      <path d="M5.4 13.4c0 3.2 2.6 5.2 6.6 5.2s6.6-2 6.6-5.2" />
      <path d="M4.6 10.2h14.8" />
      <path d="M6.4 10.2v3.2M12 10.2v3.2M17.6 10.2v3.2" />
    </>
  ),
  // Golden staircase (environmental landmark)
  staircase: (
    <>
      <path d="M3.6 19.2h4.2v-3.4H12v-3.4h4.2V9h4.2" />
      <path d="M20.4 9V7.8" />
      <circle cx="20.4" cy="6.6" r="1.1" />
    </>
  ),
  // Tiered luxury fountain (environmental landmark)
  fountain: (
    <>
      <path d="M12 4.2v2.2" />
      <path d="M8.6 6.4h6.8c-.4 2-1.6 3-3.4 3s-3-1-3.4-3Z" />
      <path d="M6.2 11h11.6c-.5 2.6-2.4 4-5.8 4s-5.3-1.4-5.8-4Z" />
      <path d="M12 9.4V11M12 15v2.2" />
      <path d="M4.4 17.2h15.2" />
      <path d="M6 17.2c0 1.6 2.4 2.6 6 2.6s6-1 6-2.6" />
    </>
  ),
  // Marble platform podium (environmental landmark)
  platform: (
    <>
      <ellipse cx="12" cy="14.6" rx="8.4" ry="3.1" />
      <path d="M3.6 14.6v2.6c0 1.8 3.8 3.2 8.4 3.2s8.4-1.4 8.4-3.2v-2.6" />
      <path d="M12 4.6l.95 1.95 2.15.31-1.55 1.52.36 2.14L12 9.51l-1.91 1.01.36-2.14L8.9 6.86l2.15-.31L12 4.6Z" />
    </>
  ),
  // Boulevard bench (environmental landmark)
  bench: (
    <>
      <path d="M3.6 10.2h16.8" />
      <path d="M4.4 10.2V6.8M12 10.2V7.4M19.6 10.2V6.8" />
      <path d="M3 13.8h18" />
      <path d="M5 13.8v5M19 13.8v5" />
      <path d="M3.6 10.2v3.6M20.4 10.2v3.6" />
    </>
  ),
  // Laurel crown (ultimate crest)
  laurel: (
    <>
      <path d="M12 21c-4.4-1-7.4-4.2-7.4-9.4 3.2 0 5.3 2.1 5.3 5.3" />
      <path d="M12 21c4.4-1 7.4-4.2 7.4-9.4-3.2 0-5.3 2.1-5.3 5.3" />
      <path d="M12 21v-9.6" />
      <path d="M12 3.6l1.2 2.5 2.7.4-2 1.9.5 2.7L12 9.7 9.6 11l.5-2.7-2-1.9 2.7-.4L12 3.6Z" />
    </>
  ),
}

interface LuxIconProps {
  name: LuxIconName
  className?: string
  style?: CSSProperties
  strokeWidth?: number
  title?: string
}

// LuxIcon renders one icon. `currentColor` drives the stroke, so tint with a
// text-colour class or an inline `color`.
export function LuxIcon({ name, className, style, strokeWidth = 1.5, title }: LuxIconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={className}
      style={style}
      fill="none"
      stroke="currentColor"
      strokeWidth={strokeWidth}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden={title ? undefined : true}
      role={title ? 'img' : undefined}
    >
      {title ? <title>{title}</title> : null}
      {PATHS[name]}
    </svg>
  )
}
