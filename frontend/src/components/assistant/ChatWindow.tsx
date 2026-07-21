import { useEffect, useRef, useState, useCallback } from 'react'
import { motion } from 'framer-motion'
import { ArrowLeft, RotateCcw, Sparkles } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useChat, departmentPrompts } from '@/hooks/useChat'
import { apiFetch } from '@/lib/api'
import { ChatBubble } from '@/components/assistant/ChatBubble'
import { AILoader } from '@/components/assistant/AILoader'
import { DepartmentSelector } from '@/components/assistant/DepartmentSelector'
import { PromptChip } from '@/components/assistant/PromptChip'
import { PromptInput } from '@/components/assistant/PromptInput'
import type { DepartmentOption, ChatHistoryEntry, ChatMessage, CollectionResponse } from '@/types'

// Optimistic fallback shown while the /assistant/departments query loads,
// matching the backend's actual list so the screen is never blank.
const FALLBACK_DEPARTMENTS: DepartmentOption[] = [
  { name: 'Men', available: true },
  { name: 'Women', available: true },
  { name: 'Kids - Girls', available: false },
  { name: 'Kids - Boys', available: false },
  { name: 'Beauty', available: false },
  { name: 'Footwear', available: true },
  { name: 'Accessories', available: true },
]

interface ChatWindowProps {
  reopenEntry?: ChatHistoryEntry | null
  onReopenHandled?: () => void
}

export function ChatWindow({ reopenEntry, onReopenHandled }: ChatWindowProps = {}) {
  const {
    messages,
    isThinking,
    sendMessage,
    loadMoreCollections,
    loadBrowseAll,
    clearChat,
    resetMessages,
    restoreMessages,
    department,
    selectDepartment,
  } = useChat()
  const [draft, setDraft] = useState('')
  const scrollRef = useRef<HTMLDivElement>(null)

  const { data: departments } = useQuery<DepartmentOption[]>({
    queryKey: ['departments'],
    queryFn: () => apiFetch<DepartmentOption[]>('/assistant/departments'),
    staleTime: 60 * 60 * 1000,
  })

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' })
  }, [messages, isThinking])

  // Restore a past conversation when the user clicks a history item
  useEffect(() => {
    if (!reopenEntry) return
    selectDepartment(reopenEntry.department)
    // Convert stored collections back into a chat message pair
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
    // We inject directly via the hook's setter by re-sending; simpler is to
    // just set messages directly. useChat exposes a restoreMessages helper for this.
    restoreMessages([userMsg, assistantMsg])
    onReopenHandled?.()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [reopenEntry])

  const handleSend = (value: string) => {
    sendMessage(value)
    setDraft('')
  }

  const prompts = department ? (departmentPrompts[department] ?? []) : []

  return (
    <div className="mx-auto flex h-[calc(100vh-6rem)] max-w-4xl flex-col px-5 lg:px-8">
      {/* ── Header ─────────────────────────────────────────────────────── */}
      <div className="flex items-center justify-between border-b border-line py-4">
        <div className="flex items-center gap-2.5">
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-brand-500 to-ink-900">
            <Sparkles className="h-4 w-4 text-white" />
          </span>
          <div>
            <p className="font-display text-sm font-bold text-ink-900">Bharat AI Stylist</p>
            {department ? (
              <p className="text-[11px] font-medium text-brand-500">● {department}</p>
            ) : (
              <p className="text-[11px] text-emerald-600">● Online</p>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2">
          {/* "Change" goes back to step 1 (department picker) */}
          {department && (
            <button
              onClick={clearChat}
              className="flex items-center gap-1.5 rounded-full border border-line px-3 py-1.5 text-xs font-medium text-ink-500 transition-colors hover:border-ink-900 hover:text-ink-900"
            >
              <ArrowLeft className="h-3 w-3" />
              Change
            </button>
          )}
          {/* "New Chat" keeps the department, clears messages only */}
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

      {/* ── Scroll area ─────────────────────────────────────────────────── */}
      <div ref={scrollRef} className="no-scrollbar flex-1 space-y-6 overflow-y-auto py-6">
        {/* Step 1 — no department selected */}
        {!department && (
          <DepartmentSelector
            departments={departments ?? FALLBACK_DEPARTMENTS}
            onSelect={selectDepartment}
          />
        )}

        {/* Step 2 — department selected, no messages yet */}
        {department && messages.length === 0 && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex h-full flex-col items-center justify-center py-16 text-center"
          >
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-brand-50 to-brand-100">
              <Sparkles className="h-7 w-7 text-brand-500" />
            </div>
            <h2 className="mt-5 font-display text-xl font-bold text-ink-900">
              What are you looking for today?
            </h2>
            <p className="mt-2 max-w-sm text-sm text-ink-500">
              Tell me the occasion, your budget, or the vibe — I'll build a complete{' '}
              <strong>{department}</strong> collection for you.
            </p>
          </motion.div>
        )}

        {/* Chat messages */}
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

      {/* ── Input (only visible once a department is chosen) ──────────── */}
      {department && (
        <div className="pb-6">
          {/* Context-aware prompt chips — only shown before first message */}
          {messages.length === 0 && prompts.length > 0 && (
            <div className="no-scrollbar mb-3 flex gap-2 overflow-x-auto pb-1">
              {prompts.map((prompt) => (
                <PromptChip key={prompt} label={prompt} onClick={() => setDraft(prompt)} />
              ))}
            </div>
          )}
          <PromptInput
            value={draft}
            onChange={setDraft}
            onSubmit={handleSend}
            disabled={isThinking}
          />
        </div>
      )}
    </div>
  )
}
