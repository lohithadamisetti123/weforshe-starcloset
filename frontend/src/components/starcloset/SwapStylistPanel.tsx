import { motion, AnimatePresence } from 'framer-motion'
import { X } from 'lucide-react'
import { PersonaPicker } from '@/components/starcloset/PersonaPicker'
import { usePersona } from '@/context/PersonaContext'

// SwapStylistPanel is the slide-over the Stylist Studio uses to change stylists
// mid-session. Choosing an unlocked stylist swaps the active persona and closes
// the panel.
export function SwapStylistPanel({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { unlockScore } = usePersona()

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          className="fixed inset-0 z-[60] flex justify-end"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
        >
          <div className="absolute inset-0 bg-ink-900/40 backdrop-blur-sm" onClick={onClose} />
          <motion.aside
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', stiffness: 320, damping: 34 }}
            className="relative flex h-full w-full max-w-2xl flex-col overflow-y-auto bg-surface p-6 shadow-2xl"
          >
            <div className="flex items-center justify-between">
              <div>
                <h2 className="font-display text-xl font-bold text-ink-900">Swap your stylist</h2>
                <p className="text-sm text-ink-500">{unlockScore} items in your cart · unlock more as you shop</p>
              </div>
              <button
                onClick={onClose}
                aria-label="Close"
                className="flex h-9 w-9 items-center justify-center rounded-full border border-line text-ink-500 hover:border-ink-900 hover:text-ink-900"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
            <div className="mt-6">
              <PersonaPicker onSelected={onClose} />
            </div>
          </motion.aside>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
