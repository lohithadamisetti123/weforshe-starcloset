import { type ButtonHTMLAttributes, forwardRef } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-full font-semibold transition-all duration-200 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50 active:scale-[0.97]',
  {
    variants: {
      variant: {
        primary:
          'bg-brand-500 text-white shadow-[0_8px_24px_-8px_rgba(255,63,108,0.6)] hover:bg-brand-600 hover:shadow-[0_12px_28px_-8px_rgba(255,63,108,0.7)]',
        secondary: 'bg-ink-900 text-white hover:bg-ink-700',
        outline: 'border border-ink-900/15 text-ink-900 bg-white hover:border-brand-500 hover:text-brand-600',
        ghost: 'text-ink-700 hover:bg-ink-900/5',
        glass: 'glass border border-white/40 text-ink-900 hover:bg-white/90',
      },
      size: {
        sm: 'h-9 px-4 text-sm',
        md: 'h-11 px-6 text-sm',
        lg: 'h-13 px-8 text-base',
        icon: 'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  }
)

export interface ButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return <button ref={ref} className={cn(buttonVariants({ variant, size }), className)} {...props} />
  }
)
Button.displayName = 'Button'
