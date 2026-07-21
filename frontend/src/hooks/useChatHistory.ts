import { useCallback, useEffect, useState } from 'react'
import type { ChatHistoryEntry, CollectionResponse } from '@/types'

const KEY = 'myntra-chat-history'
const MAX = 50 // keep the last 50 conversations

function load(): ChatHistoryEntry[] {
  try { return JSON.parse(localStorage.getItem(KEY) ?? '[]') }
  catch { return [] }
}

function persist(entries: ChatHistoryEntry[]) {
  try { localStorage.setItem(KEY, JSON.stringify(entries)) }
  catch {}
}

export function useChatHistory() {
  const [entries, setEntries] = useState<ChatHistoryEntry[]>(load)

  useEffect(() => { persist(entries) }, [entries])

  // Called from useChat every time an assistant reply with collections lands
  const addEntry = useCallback((
    department: string,
    prompt: string,
    collections: CollectionResponse[],
  ) => {
    const entry: ChatHistoryEntry = {
      id: crypto.randomUUID(),
      conversationId: crypto.randomUUID(),
      timestamp: new Date().toISOString(),
      department,
      prompt,
      collections,
    }
    setEntries((prev) => [entry, ...prev].slice(0, MAX))
    return entry
  }, [])

  const remove = useCallback((id: string) => {
    setEntries((prev) => prev.filter((e) => e.id !== id))
  }, [])

  const clear = useCallback(() => setEntries([]), [])

  return { entries, addEntry, remove, clear }
}
