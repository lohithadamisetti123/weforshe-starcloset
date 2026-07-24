import { useState } from 'react'
import { cn } from '@/lib/utils'
import type { StylistPersona } from '@/types/persona'

// StylistAvatar renders a persona's portrait over its signature accent gradient,
// falling back to the stylist's initials when the image is missing or fails to
// load — so a persona with no real photo (every seed persona) still looks
// intentional rather than broken.
interface StylistAvatarProps {
  persona: Pick<StylistPersona, 'displayName' | 'portraitUrl' | 'accentFrom' | 'accentTo'>
  size?: 'sm' | 'md' | 'lg' | 'xl'
  locked?: boolean
  className?: string
}

const SIZES: Record<NonNullable<StylistAvatarProps['size']>, string> = {
  sm: 'h-9 w-9 text-xs',
  md: 'h-12 w-12 text-sm',
  lg: 'h-16 w-16 text-lg',
  xl: 'h-24 w-24 text-2xl',
}

function initials(name: string): string {
  const parts = name.trim().split(/\s+/)
  if (parts.length === 0) return '?'
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}

export function StylistAvatar({ persona, size = 'md', locked = false, className }: StylistAvatarProps) {
  const [imgOk, setImgOk] = useState(true)
  const showImg = imgOk && !!persona.portraitUrl

  return (
    <span
      className={cn(
        'relative flex shrink-0 items-center justify-center overflow-hidden rounded-full font-display font-bold text-white shadow-sm',
        SIZES[size],
        locked && 'grayscale',
        className,
      )}
      style={{
        backgroundImage: `linear-gradient(135deg, ${persona.accentFrom || '#FF3F6C'}, ${persona.accentTo || '#1F1C18'})`,
      }}
    >
      {showImg ? (
        <img
          src={persona.portraitUrl}
          alt={persona.displayName}
          loading="lazy"
          decoding="async"
          className="h-full w-full object-cover"
          onError={() => setImgOk(false)}
        />
      ) : (
        <span className="select-none">{initials(persona.displayName)}</span>
      )}
    </span>
  )
}
