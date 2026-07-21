import { useState, type FormEvent, type KeyboardEvent } from 'react'
import { motion } from 'framer-motion'
import { ArrowUp, Mic, Paperclip } from 'lucide-react'
import { cn } from '@/lib/utils'

const MAX_CHARS = 280

interface PromptInputProps {
  value: string
  onChange: (value: string) => void
  onSubmit: (value: string) => void
  disabled?: boolean
}

export function PromptInput({ value, onChange, onSubmit, disabled }: PromptInputProps) {
  const [recording, setRecording] = useState(false)

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!value.trim() || disabled) return
    onSubmit(value)
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (value.trim() && !disabled) onSubmit(value)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="glass rounded-xl border border-line p-2.5 shadow-[0_8px_28px_-12px_rgba(17,24,39,0.18)]">
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value.slice(0, MAX_CHARS))}
        onKeyDown={handleKeyDown}
        placeholder="Describe the outfit you're looking for..."
        rows={1}
        disabled={disabled}
        className="max-h-28 w-full resize-none bg-transparent px-2.5 py-2 text-sm text-ink-900 placeholder:text-ink-300 focus:outline-none"
      />
      <div className="flex items-center justify-between px-1 pt-1">
        <div className="flex items-center gap-1">
          <button
            type="button"
            aria-label="Attach file"
            className="flex h-9 w-9 items-center justify-center rounded-full text-ink-500 transition-colors hover:bg-ink-900/5 hover:text-ink-900"
          >
            <Paperclip className="h-4 w-4" />
          </button>
          <button
            type="button"
            aria-label="Voice input"
            aria-pressed={recording}
            onClick={() => setRecording((v) => !v)}
            className={cn(
              'relative flex h-9 w-9 items-center justify-center rounded-full transition-colors',
              recording ? 'bg-brand-50 text-brand-600' : 'text-ink-500 hover:bg-ink-900/5 hover:text-ink-900'
            )}
          >
            {recording && (
              <motion.span
                className="absolute inset-0 rounded-full bg-brand-500/20"
                animate={{ scale: [1, 1.6], opacity: [0.6, 0] }}
                transition={{ duration: 1.2, repeat: Infinity }}
              />
            )}
            <Mic className="h-4 w-4" />
          </button>
          <span className="ml-1 text-[11px] text-ink-300">
            {value.length}/{MAX_CHARS}
          </span>
        </div>

        <button
          type="submit"
          disabled={!value.trim() || disabled}
          aria-label="Send message"
          className="flex h-10 w-10 items-center justify-center rounded-full bg-brand-500 text-white shadow-[0_8px_20px_-6px_rgba(255,63,108,0.6)] transition-all hover:bg-brand-600 disabled:pointer-events-none disabled:opacity-40"
        >
          <ArrowUp className="h-4.5 w-4.5" />
        </button>
      </div>
    </form>
  )
}
