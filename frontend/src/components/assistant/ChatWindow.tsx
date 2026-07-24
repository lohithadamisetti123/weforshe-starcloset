import { useEffect, useRef, useState } from 'react'
import { motion } from 'framer-motion'
import { RotateCcw, ArrowLeftRight, Users } from 'lucide-react'
import { useChat } from '@/hooks/useChat'
import { useStarCloset } from '@/context/StarClosetContext'
import { ChatBubble } from '@/components/assistant/ChatBubble'
import { AILoader } from '@/components/assistant/AILoader'
import { PromptChip } from '@/components/assistant/PromptChip'
import { PromptInput } from '@/components/assistant/PromptInput'
import { StylistAvatar } from '@/components/starcloset/StylistAvatar'
import { DepartmentPicker } from '@/components/starcloset/DepartmentPicker'
import { ProgressionMap } from '@/components/starcloset/ProgressionMap'
import { Confetti } from '@/components/starcloset/Confetti'
import { Toast } from '@/components/ui/Toast'
import type { ChatHistoryEntry, ChatMessage } from '@/types'

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
    resetMessages,
    restoreMessages,
    department,
    selectDepartment,
  } = useChat()
  const {
    selectedDepartment,
    currentDept,
    selectedPersona,
    collectionsDepartment,
    clearDepartment,
    clearPersona,
    unlockScore,
    nextLocked,
    itemsToUnlock,
    recordCartAdd,
  } = useStarCloset()

  const [draft, setDraft] = useState('')
  const [confetti, setConfetti] = useState(false)
  const [toast, setToast] = useState<string | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Keep the session's engine department in step with the chosen StarCloset
  // department + stylist. Switching department (or, within a department,
  // switching to a new stylist) starts the chat from a clean slate so each
  // stylist gives fresh looks.
  const journeyKey = `${selectedDepartment}|${selectedPersona?.slug}|${collectionsDepartment}`
  const prevKey = useRef<string | null>(null)
  useEffect(() => {
    if (!selectedPersona || !collectionsDepartment) return
    if (prevKey.current === journeyKey) return
    prevKey.current = journeyKey
    if (department !== collectionsDepartment) {
      selectDepartment(collectionsDepartment) // also clears the transcript
    } else {
      resetMessages()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [journeyKey])

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

  // A cart-add advances the current department's journey; a threshold crossing
  // triggers the celebration.
  const handleCartAdd = async () => {
    const unlocked = await recordCartAdd()
    if (unlocked.length > 0) {
      setConfetti(true)
      setToast(`🎉 ${unlocked.map((s) => s.displayName).join(' & ')} unlocked!`)
      setTimeout(() => setConfetti(false), 2800)
    }
  }

  // ── Step 1 — no department chosen ──────────────────────────────────────────
  if (!selectedDepartment) {
    return <DepartmentPicker />
  }

  // ── Step 2 — department chosen, no stylist yet: the progression map ────────
  if (!selectedPersona) {
    return <ProgressionMap />
  }

  // ── Step 3 — chat with the chosen stylist ──────────────────────────────────
  const prompts = selectedPersona.samplePrompts ?? []

  return (
    <div className="mx-auto flex h-[calc(100vh-6rem)] max-w-4xl flex-col px-5 lg:px-8">
      {confetti && <Confetti />}

      {/* Header: department + active stylist */}
      <div className="flex items-center justify-between border-b border-line py-4">
        <div className="flex items-center gap-3">
          <StylistAvatar persona={selectedPersona} size="md" />
          <div>
            <p className="font-display text-sm font-bold text-ink-900">{selectedPersona.displayName}</p>
            <p className="text-[11px] font-medium text-brand-500">
              {currentDept?.emoji} {currentDept?.label} · {unlockScore} item{unlockScore === 1 ? '' : 's'}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={clearPersona}
            className="flex items-center gap-1.5 rounded-full border border-line px-3 py-1.5 text-xs font-medium text-ink-500 transition-colors hover:border-ink-900 hover:text-ink-900"
          >
            <Users className="h-3 w-3" /> Stylists
          </button>
          <button
            onClick={clearDepartment}
            className="flex items-center gap-1.5 rounded-full border border-line px-3 py-1.5 text-xs font-medium text-ink-500 transition-colors hover:border-ink-900 hover:text-ink-900"
          >
            <ArrowLeftRight className="h-3 w-3" /> Department
          </button>
          {messages.length > 0 && (
            <button
              onClick={resetMessages}
              className="flex items-center gap-1.5 rounded-full border border-line px-3 py-1.5 text-xs font-medium text-ink-500 transition-colors hover:border-ink-900 hover:text-ink-900"
            >
              <RotateCcw className="h-3 w-3" /> New Chat
            </button>
          )}
        </div>
      </div>

      {/* Scroll area */}
      <div ref={scrollRef} className="no-scrollbar flex-1 space-y-6 overflow-y-auto py-6">
        {messages.length === 0 && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex h-full flex-col items-center justify-center py-12 text-center"
          >
            <StylistAvatar persona={selectedPersona} size="xl" />
            <h2 className="mt-5 font-display text-xl font-bold text-ink-900">{selectedPersona.displayName}</h2>
            <p className="mt-3 max-w-md text-sm leading-relaxed text-ink-600">{selectedPersona.greeting}</p>
          </motion.div>
        )}

        {messages.map((message) => (
          <ChatBubble
            key={message.id}
            message={message}
            isThinking={isThinking}
            onLoadMore={loadMoreCollections}
            onBrowseAll={loadBrowseAll}
            onCartAdd={handleCartAdd}
          />
        ))}

        {isThinking && <AILoader />}
      </div>

      {/* Input + unlock hint */}
      <div className="space-y-3 pb-6">
        {nextLocked && (
          <div className="flex items-center gap-2 rounded-2xl border border-line bg-white px-4 py-2.5 text-xs">
            <StylistAvatar persona={nextLocked} size="sm" locked />
            <span className="font-semibold text-ink-700">
              Add {itemsToUnlock(nextLocked)} more item{itemsToUnlock(nextLocked) === 1 ? '' : 's'} to unlock{' '}
              <span className="text-brand-600">{nextLocked.displayName}</span> in {currentDept?.label}
            </span>
          </div>
        )}

        {messages.length === 0 && prompts.length > 0 && (
          <div className="no-scrollbar flex gap-2 overflow-x-auto pb-1">
            {prompts.map((prompt) => (
              <PromptChip key={prompt} label={prompt} onClick={() => setDraft(prompt)} />
            ))}
          </div>
        )}
        <PromptInput value={draft} onChange={setDraft} onSubmit={handleSend} disabled={isThinking} />
      </div>

      <Toast message={toast} onDismiss={() => setToast(null)} />
    </div>
  )
}
