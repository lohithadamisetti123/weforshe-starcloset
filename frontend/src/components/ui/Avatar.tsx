import { cn } from '@/lib/utils'
import { Sparkles, User } from 'lucide-react'

interface AvatarProps {
  variant: 'ai' | 'user'
  className?: string
}

export function Avatar({ variant, className }: AvatarProps) {
  if (variant === 'ai') {
    return (
      <div
        className={cn(
          'flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-brand-500 to-ink-900 text-white shadow-md',
          className
        )}
      >
        <Sparkles className="h-4 w-4" />
      </div>
    )
  }
  return (
    <div
      className={cn(
        'flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-ink-900 text-white',
        className
      )}
    >
      <User className="h-4 w-4" />
    </div>
  )
}
