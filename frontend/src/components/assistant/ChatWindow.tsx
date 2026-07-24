import { useEffect, useRef, useState } from 'react'
import { motion } from 'framer-motion'
import { RotateCcw, Repeat } from 'lucide-react'
import { useChat } from '@/hooks/useChat'
import { usePersona } from '@/context/PersonaContext'
import { ChatBubble } from '@/components/assistant/ChatBubble'
import { AILoader } from '@/components/assistant/AILoader'
import { PromptChip } from '@/components/assistant/PromptChip'
import { PromptInput } from '@/components/assistant/PromptInput'
import { StylistAvatar } from '@/components/starcloset/StylistAvatar'
import { SwapStylistPanel } from '@/components/starcloset/SwapStylistPanel'
import { UnlockProgressBanner } from '@/components/starcloset/UnlockProgressBanner'
import { PersonaPicker } from '@/components/starcloset/PersonaPicker'
import type { ChatHistoryEntry, ChatMessage } from '@/types'

interface ChatWindowProps {
  reopenEntry?: ChatHistoryEntry | null
  onReopenHandled?: () => void
}

// genderToDept maps a persona's default gender onto the guided flow's department
// so choosing a stylist scopes the studio without a separate department step.
function genderToDept(gender: string): string {
  switch (gender) {
    case 'Men':
      return 'Men'
    case 'Girls':
      return 'Kids - Girls'
    case 'Boys':
      return 'Kids - Boys'
    default:
      return 'Women'
  }
}

export function ChatWindow({ reopenEntry, onReopenHandled }: ChatWindowProps = {}) {
  const {
    messages,
    isThinking,
    sendMessage,
    loadMoreCollections,
    loadBrowseAll,
    resetMessages,
    restoreMessages,
    department,
    selectDepartment,
  } = useChat()
  const { selectedPersona } = usePersona()
  const [draft, setDraft] = useState('')
  const [swapOpen, setSwapOpen] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)

  // The chosen stylist scopes the department. Keyed on the mapped department so
  // it re-runs only when the stylist (or their gender) actually changes.
  const mappedDept = selectedPersona ? genderToDept(selectedPersona.defaultGender) : null
  useEffect(() => {
    if (mappedDept && mappedDept !== department) {
      selectDepartment(mappedDept)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mappedDept])

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' })
  }, [messages, isThinking])

  // Restore a past conversation when the user clicks a history item.
  useEffect(() => {
    if (!reopenEntry) return
    selectDepartment(reopenEntry.department)
    const userMsg: ChatMessage = {
      id: Math.random().toString(36).slice(2),
      role: 'user',
      content: reopenEntry.prompt,
      timestamp: reopenEntry.timestamp,
    }
    const assistantMsg: ChatMessage = {
      id: Math.random().toString(36).slice(2),
      role: 'assistant',
      content: `${reopenEntry.collections.length} collections for "${reopenEntry.prompt}"`,
      timestamp: reopenEntry.timestamp,
      collections: reopenEntry.collections,
    }
    restoreMessages([userMsg, assistantMsg])
    onReopenHandled?.()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [reopenEntry])

  const handleSend = (value: string) => {
    sendMessage(value)
    setDraft('')
  }

  // ── No stylist chosen yet — the studio opens on the roster. ────────────────
  if (!selectedPersona) {
    return (
      <div className="mx-auto max-w-4xl px-5 py-10 lg:px-8">
        <div className="text-center">
          <h1 className="font-display text-2xl font-bold text-ink-900">Choose your stylist</h1>
          <p className="mt-2 text-sm text-ink-500">
            Pick a celebrity stylist to start. Unlock more as you add pieces to your cart.
          </p>
        </div>
        <div className="mt-8">
          <PersonaPicker />
        </div>
      </div>
    )
  }

  const prompts = selectedPersona.samplePrompts ?? []

  return (
    <div className="mx-auto flex h-[calc(100vh-6rem)] max-w-4xl flex-col px-5 lg:px-8">
      {/* ── Header: active stylist ─────────────────────────────────────────── */}
      <div className="flex items-center justify-between border-b border-line py-4">
        <div className="flex items-center gap-3">
          <StylistAvatar persona={selectedPersona} size="md" />
          <div>
            <p className="font-display text-sm font-bold text-ink-900">{selectedPersona.displayName}</p>
            <p className="text-[11px] font-medium text-brand-500">● Styling {department}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setSwapOpen(true)}
            className="flex items-center gap-1.5 rounded-full border border-line px-3 py-1.5 text-xs font-medium text-ink-500 transition-colors hover:border-ink-900 hover:text-ink-900"
          >
            <Repeat className="h-3 w-3" />
            Swap stylist
          </button>
          {messages.length > 0 && (
            <button
              onClick={resetMessages}
              className="flex items-center gap-1.5 rounded-full border border-line px-3 py-1.5 text-xs font-medium text-ink-500 transition-colors hover:border-ink-900 hover:text-ink-900"
            >
              <RotateCcw className="h-3 w-3" />
              New Chat
            </button>
          )}
        </div>
      </div>

      {/* ── Scroll area ────────────────────────────────────────────────────── */}
      <div ref={scrollRef} className="no-scrollbar flex-1 space-y-6 overflow-y-auto py-6">
        {messages.length === 0 && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex h-full flex-col items-center justify-center py-12 text-center"
          >
            <StylistAvatar persona={selectedPersona} size="xl" />
            <h2 className="mt-5 font-display text-xl font-bold text-ink-900">
              {selectedPersona.displayName}
            </h2>
            <p className="mt-3 max-w-md text-sm leading-relaxed text-ink-600">
              {selectedPersona.greeting}
            </p>
          </motion.div>
        )}

        {messages.map((message) => (
          <ChatBubble
            key={message.id}
            message={message}
            isThinking={isThinking}
            onLoadMore={loadMoreCollections}
            onBrowseAll={loadBrowseAll}
          />
        ))}

        {isThinking && <AILoader />}
      </div>

      {/* ── Input + unlock loop ────────────────────────────────────────────── */}
      <div className="space-y-3 pb-6">
        <UnlockProgressBanner />

        {messages.length === 0 && prompts.length > 0 && (
          <div className="no-scrollbar flex gap-2 overflow-x-auto pb-1">
            {prompts.map((prompt) => (
              <PromptChip key={prompt} label={prompt} onClick={() => setDraft(prompt)} />
            ))}
          </div>
        )}
        <PromptInput value={draft} onChange={setDraft} onSubmit={handleSend} disabled={isThinking} />
      </div>

      <SwapStylistPanel open={swapOpen} onClose={() => setSwapOpen(false)} />
    </div>
  )
}
