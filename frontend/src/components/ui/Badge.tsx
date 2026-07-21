import type { HTMLAttributes } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const badgeVariants = cva('inline-flex items-center gap-1.5 rounded-full font-medium', {
  variants: {
    variant: {
      brand: 'bg-brand-50 text-brand-600 border border-brand-100',
      dark: 'bg-ink-900 text-white',
      outline: 'border border-ink-900/15 text-ink-700',
      success: 'bg-emerald-50 text-emerald-600 border border-emerald-100',
    },
    size: {
      sm: 'px-2.5 py-1 text-[11px]',
      md: 'px-3 py-1.5 text-xs',
    },
  },
  defaultVariants: { variant: 'brand', size: 'sm' },
})

interface BadgeProps extends HTMLAttributes<HTMLSpanElement>, VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, size, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant, size }), className)} {...props} />
}
