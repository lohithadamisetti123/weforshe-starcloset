import { useEffect } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { CheckCircle2, X } from 'lucide-react'

interface ToastProps {
  message: string | null
  onDismiss: () => void
  variant?: 'success' | 'error'
}

export function Toast({ message, onDismiss, variant = 'success' }: ToastProps) {
  useEffect(() => {
    if (!message) return
    const t = setTimeout(onDismiss, 3000)
    return () => clearTimeout(t)
  }, [message, onDismiss])

  return (
    <AnimatePresence>
      {message && (
        <motion.div
          initial={{ opacity: 0, y: 60, scale: 0.96 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: 40, scale: 0.96 }}
          transition={{ type: 'spring', damping: 24 }}
          className={`fixed bottom-6 left-1/2 z-[200] flex -translate-x-1/2 items-center gap-2.5 rounded-full px-5 py-3 text-sm font-semibold shadow-xl ${
            variant === 'success'
              ? 'bg-ink-900 text-white'
              : 'bg-red-600 text-white'
          }`}
        >
          <CheckCircle2 className="h-4 w-4 shrink-0 text-brand-400" />
          {message}
          <button onClick={onDismiss} className="ml-1">
            <X className="h-3.5 w-3.5 opacity-60 hover:opacity-100" />
          </button>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
