import { useState, type FormEvent, type KeyboardEvent } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { ArrowUp } from 'lucide-react'
import { LuxIcon } from '@/components/starcloset/LuxuryIcons'

const MAX_CHARS = 280

interface PromptInputProps {
  value: string
  onChange: (value: string) => void
  onSubmit: (value: string) => void
  disabled?: boolean
}

// PromptInput — the stylist composer. A single elegant glass field: champagne
// hairline at rest, a soft pink glow when focused, and a spring-loaded gradient
// send button. No accessory buttons — just the conversation.
export function PromptInput({ value, onChange, onSubmit, disabled }: PromptInputProps) {
  const [focused, setFocused] = useState(false)
  const canSend = !!value.trim() && !disabled

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!canSend) return
    onSubmit(value)
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (canSend) onSubmit(value)
    }
  }

  return (
    <motion.form
      onSubmit={handleSubmit}
      className="rounded-2xl border bg-white/80 p-2.5 backdrop-blur-xl"
      animate={{
        borderColor: focused ? 'rgba(255,63,108,0.45)' : 'rgba(231,206,143,0.55)',
        boxShadow: focused
          ? '0 14px 40px -14px rgba(255,63,108,0.35), 0 0 0 4px rgba(255,63,108,0.07), inset 0 0 0 1px rgba(255,255,255,0.6)'
          : '0 10px 30px -14px rgba(17,24,39,0.22), inset 0 0 0 1px rgba(255,255,255,0.6)',
      }}
      transition={{ duration: 0.35, ease: 'easeOut' }}
    >
      <div className="flex items-start gap-2 px-1.5 pt-1.5">
        {/* Atelier mark — a quiet gold sparkle where the clutter used to be */}
        <motion.span
          className="mt-1 grid h-6 w-6 shrink-0 place-items-center rounded-full"
          style={{ background: 'linear-gradient(135deg,#FBF3DD,#EBD9A6)', boxShadow: 'inset 0 0 0 1px rgba(201,162,75,0.5)' }}
          animate={focused ? { scale: [1, 1.12, 1] } : {}}
          transition={{ duration: 1.6, repeat: focused ? Infinity : 0, ease: 'easeInOut' }}
        >
          <LuxIcon name="sparkle" className="h-3.5 w-3.5 text-[#A9822F]" strokeWidth={1.6} />
        </motion.span>

        <textarea
          value={value}
          onChange={(e) => onChange(e.target.value.slice(0, MAX_CHARS))}
          onKeyDown={handleKeyDown}
          onFocus={() => setFocused(true)}
          onBlur={() => setFocused(false)}
          placeholder="Describe the look you have in mind…"
          rows={1}
          disabled={disabled}
          className="max-h-28 w-full resize-none bg-transparent py-1 text-sm text-ink-900 placeholder:text-ink-300 focus:outline-none"
        />
      </div>

      <div className="flex items-center justify-between pl-2.5 pr-1 pt-1.5">
        <div className="flex items-center gap-2">
          <span className="text-[9px] font-semibold uppercase tracking-[0.2em] text-ink-300">Enter to send</span>
          <AnimatePresence>
            {value.length > MAX_CHARS - 60 && (
              <motion.span
                initial={{ opacity: 0, y: 4 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: 4 }}
                className="text-[10px] font-medium tabular-nums"
                style={{ color: value.length >= MAX_CHARS ? '#eb2f5c' : '#A9822F' }}
              >
                {value.length}/{MAX_CHARS}
              </motion.span>
            )}
          </AnimatePresence>
        </div>

        <motion.button
          type="submit"
          disabled={!canSend}
          aria-label="Send message"
          whileHover={canSend ? { scale: 1.07, y: -1 } : {}}
          whileTap={canSend ? { scale: 0.93 } : {}}
          animate={{ opacity: canSend ? 1 : 0.35 }}
          transition={{ type: 'spring', stiffness: 320, damping: 20 }}
          className="relative flex h-10 w-10 items-center justify-center overflow-hidden rounded-full text-white disabled:pointer-events-none"
          style={{ background: 'linear-gradient(135deg,#FF3F6C,#E7B84B)', boxShadow: canSend ? '0 10px 24px -8px rgba(255,63,108,0.65), inset 0 0 0 1px rgba(255,255,255,0.25)' : 'none' }}
        >
          {canSend && (
            <motion.span
              className="pointer-events-none absolute inset-y-0 -left-1/2 w-1/3 skew-x-[-20deg] bg-white/35"
              animate={{ x: ['0%', '480%'] }}
              transition={{ duration: 2.4, repeat: Infinity, repeatDelay: 1.2, ease: 'easeInOut' }}
            />
          )}
          <ArrowUp className="h-4.5 w-4.5" />
        </motion.button>
      </div>
    </motion.form>
  )
}
