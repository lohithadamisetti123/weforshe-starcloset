import { useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { Clock, Heart, Trash2, X } from 'lucide-react'
import type { ChatHistoryEntry, CollectionResponse } from '@/types'
import { useShop } from '@/context/AppContext'

function groupByDate(entries: ChatHistoryEntry[]) {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const yesterday = new Date(today)
  yesterday.setDate(today.getDate() - 1)

  const groups: Record<string, ChatHistoryEntry[]> = {}
  for (const e of entries) {
    const d = new Date(e.timestamp)
    d.setHours(0, 0, 0, 0)
    let label: string
    if (d.getTime() === today.getTime()) label = 'Today'
    else if (d.getTime() === yesterday.getTime()) label = 'Yesterday'
    else label = 'Earlier'
    if (!groups[label]) groups[label] = []
    groups[label].push(e)
  }
  return groups
}

interface HistorySidebarProps {
  open: boolean
  onClose: () => void
  onReopen: (entry: ChatHistoryEntry) => void
}

export function HistorySidebar({ open, onClose, onReopen }: HistorySidebarProps) {
  const { chatHistory, removeChatHistory, clearChatHistory } = useShop()
  const groups = groupByDate(chatHistory)
  const ORDER = ['Today', 'Yesterday', 'Earlier']

  return (
    <AnimatePresence>
      {open && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 z-40 bg-ink-900/30 backdrop-blur-sm"
          />

          {/* Drawer */}
          <motion.aside
            initial={{ x: '-100%' }}
            animate={{ x: 0 }}
            exit={{ x: '-100%' }}
            transition={{ type: 'spring', damping: 28, stiffness: 280 }}
            className="fixed left-0 top-0 z-50 flex h-full w-[320px] flex-col bg-white shadow-2xl"
          >
            {/* Header */}
            <div className="flex items-center justify-between border-b border-line px-5 py-4">
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4 text-brand-500" />
                <h2 className="font-display text-sm font-bold text-ink-900">Chat History</h2>
              </div>
              <div className="flex items-center gap-2">
                {chatHistory.length > 0 && (
                  <button
                    onClick={clearChatHistory}
                    className="text-[11px] text-ink-400 hover:text-red-500"
                  >
                    Clear all
                  </button>
                )}
                <button onClick={onClose} className="rounded-full p-1 hover:bg-surface">
                  <X className="h-4 w-4 text-ink-500" />
                </button>
              </div>
            </div>

            {/* Entries */}
            <div className="no-scrollbar flex-1 overflow-y-auto py-3">
              {chatHistory.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-20 text-center">
                  <Clock className="h-10 w-10 text-ink-200" />
                  <p className="mt-3 text-sm font-medium text-ink-400">No conversations yet</p>
                  <p className="mt-1 text-xs text-ink-300">Ask the AI for a collection to get started</p>
                </div>
              ) : (
                ORDER.filter((g) => groups[g]?.length).map((group) => (
                  <div key={group}>
                    <p className="px-5 py-2 text-[10px] font-bold uppercase tracking-widest text-ink-400">
                      {group}
                    </p>
                    {groups[group].map((entry) => (
                      <HistoryItem
                        key={entry.id}
                        entry={entry}
                        onReopen={() => { onReopen(entry); onClose() }}
                        onDelete={() => removeChatHistory(entry.id)}
                      />
                    ))}
                  </div>
                ))
              )}
            </div>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  )
}

function HistoryItem({
  entry,
  onReopen,
  onDelete,
}: {
  entry: ChatHistoryEntry
  onReopen: () => void
  onDelete: () => void
}) {
  const time = new Date(entry.timestamp).toLocaleTimeString('en-IN', {
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <div className="group flex items-start gap-3 px-5 py-3 hover:bg-surface">
      <button onClick={onReopen} className="flex min-w-0 flex-1 items-start gap-3 text-left">
        <Heart className="mt-0.5 h-3.5 w-3.5 shrink-0 text-brand-400" />
        <div className="min-w-0">
          <p className="truncate text-xs font-semibold text-ink-900">
            {entry.collections[0]?.title ?? entry.prompt}
          </p>
          <p className="truncate text-[10px] text-ink-400">
            {entry.department} · {time}
          </p>
        </div>
      </button>
      <button
        onClick={(e) => { e.stopPropagation(); onDelete() }}
        className="mt-0.5 shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
      >
        <Trash2 className="h-3 w-3 text-ink-300 hover:text-red-500" />
      </button>
    </div>
  )
}
