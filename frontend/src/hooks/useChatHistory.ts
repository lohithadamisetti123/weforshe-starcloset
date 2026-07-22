import { useCallback, useEffect, useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import type { ChatHistoryEntry, CollectionResponse } from '@/types'

const MAX = 50 // keep the last 50 conversations

// Chat history stays on the client (it is a convenience log of prompts, not
// account data the server needs), but the storage key is namespaced per user id
// so two accounts sharing a browser never read each other's conversations. A
// signed-out visitor gets no key at all.
function keyFor(userId: string | null): string | null {
  return userId ? `myntra-chat-history:${userId}` : null
}

function load(key: string | null): ChatHistoryEntry[] {
  if (!key) return []
  try { return JSON.parse(localStorage.getItem(key) ?? '[]') }
  catch { return [] }
}

function persist(key: string | null, entries: ChatHistoryEntry[]) {
  if (!key) return
  try { localStorage.setItem(key, JSON.stringify(entries)) }
  catch {}
}

export function useChatHistory() {
  const { user } = useAuth()
  const userId = user?.id ?? null
  const key = keyFor(userId)

  const [entries, setEntries] = useState<ChatHistoryEntry[]>(() => load(key))

  // Re-read (or empty) whenever the signed-in identity changes, so switching
  // accounts swaps the history instead of carrying it across.
  useEffect(() => {
    setEntries(load(keyFor(userId)))
  }, [userId])

  useEffect(() => { persist(key, entries) }, [key, entries])

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
