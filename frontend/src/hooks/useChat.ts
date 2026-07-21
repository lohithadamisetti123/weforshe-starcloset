import { useCallback, useState } from 'react'
import { useShop } from '@/context/AppContext'
import { apiFetch } from '@/lib/api'
import type {
  AIRecommendationResponse,
  CollectionResponse,
  CollectionsResponse,
  ChatMessage,
} from '@/types'

function makeId() {
  return Math.random().toString(36).slice(2, 10)
}

export const departmentPrompts: Record<string, string[]> = {
  Women: [
    'Wedding collection under ₹5000',
    'Office wear for summer',
    'Party collection under ₹3000',
    'Casual college outfits',
    'Vacation collection',
    'Navratri festive look',
    'Date night outfit',
  ],
  Men: [
    'Formal office collection',
    'Casual weekend collection',
    'Gym wear',
    'Travel outfits',
    'Wedding sherwani',
    'Smart casual collection',
  ],
  Footwear: [
    'Formal shoes under ₹3000',
    'Casual sneakers',
    'Wedding footwear',
    'Running shoes',
    'Block heels for party',
  ],
  Accessories: [
    'Office accessories under ₹2000',
    'Wedding jewellery set',
    'Casual watches',
    'Party handbag',
    'Festive earrings',
  ],
  Beauty: [
    'Everyday skincare routine',
    'Festive makeup products',
    'Office-friendly makeup',
    'Wedding day beauty',
  ],
}

// One request returns the full set of themed, mutually-distinct collections.
// The backend builds them together so no product repeats across them — a
// guarantee the previous three-parallel-call approach (with query suffixes)
// silently discarded, because each stateless call rebuilt and then threw away
// the cross-collection de-duplication.
//
// excludeIds lets "More Collections" ask for sets drawn from products not
// already shown, so a re-fetch returns genuinely new collections rather than
// the same deterministic result.
async function fetchCollections(
  department: string,
  prompt: string,
  excludeIds: string[] = [],
): Promise<CollectionResponse[]> {
  const res = await apiFetch<CollectionsResponse>('/assistant/collections', {
    method: 'POST',
    body: JSON.stringify({
      department,
      prompt,
      ...(excludeIds.length > 0 ? { excludeProductIds: excludeIds } : {}),
    }),
  })
  return res.collections ?? []
}

export function useChat() {
  const { addChatHistory } = useShop()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [isThinking, setIsThinking] = useState(false)
  const [department, setDepartment] = useState<string | null>(null)

  const selectDepartment = useCallback((dept: string) => {
    setDepartment(dept)
    setMessages([])
  }, [])

  const sendMessage = useCallback(
    async (content: string) => {
      const trimmed = content.trim()
      if (!trimmed || !department) return

      const userMessage: ChatMessage = {
        id: makeId(),
        role: 'user',
        content: trimmed,
        timestamp: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, userMessage])
      setIsThinking(true)

      try {
        const collections = await fetchCollections(department, trimmed)

        if (collections.length > 0) {
          addChatHistory(department, trimmed, collections)
        }

        const assistantMessage: ChatMessage = {
          id: makeId(),
          role: 'assistant',
          content: `${collections.length} collections for "${trimmed}"`,
          timestamp: new Date().toISOString(),
          collections,
        }
        setMessages((prev) => [...prev, assistantMessage])
      } catch (err) {
        const errorMessage: ChatMessage = {
          id: makeId(),
          role: 'assistant',
          content: err instanceof Error ? err.message : 'Something went wrong. Please try again.',
          timestamp: new Date().toISOString(),
        }
        setMessages((prev) => [...prev, errorMessage])
      } finally {
        setIsThinking(false)
      }
    },
    [department, addChatHistory],
  )

  const loadMoreCollections = useCallback(
    async (messageId: string, prompt: string) => {
      if (!department) return
      setIsThinking(true)

      // Signature of a collection = its sorted product-id set. Used to drop any
      // batch that duplicates something already on screen, so "More" can never
      // append visible repeats even if the catalogue is nearly exhausted.
      const sig = (c: CollectionResponse) =>
        c.items
          .map((it) => it.product.id)
          .filter(Boolean)
          .sort()
          .join('|')

      try {
        const current = messages.find((m) => m.id === messageId)?.collections ?? []
        const shownIds = Array.from(
          new Set(current.flatMap((c) => c.items.map((it) => it.product.id).filter(Boolean))),
        )
        const shownSigs = new Set(current.map(sig))

        const more = await fetchCollections(department, prompt, shownIds)
        const fresh = more.filter((c) => c.items.length > 0 && !shownSigs.has(sig(c)))

        if (fresh.length > 0) {
          setMessages((prev) =>
            prev.map((m) =>
              m.id === messageId
                ? { ...m, collections: [...(m.collections ?? []), ...fresh] }
                : m,
            ),
          )
        } else {
          // Nothing new to show — the catalogue for this search is exhausted.
          // Say so explicitly instead of letting the button appear dead.
          setMessages((prev) => [
            ...prev,
            {
              id: makeId(),
              role: 'assistant',
              content: `That's every distinct collection I can build for "${prompt}" right now. Try a broader search, or adjust the budget.`,
              timestamp: new Date().toISOString(),
            },
          ])
        }
      } catch (err) {
        // Surface the failure instead of swallowing it, so a backend error is
        // visible rather than a silently unresponsive button.
        setMessages((prev) => [
          ...prev,
          {
            id: makeId(),
            role: 'assistant',
            content:
              err instanceof Error
                ? `Couldn't load more collections: ${err.message}`
                : "Couldn't load more collections. Please try again.",
            timestamp: new Date().toISOString(),
          },
        ])
      } finally {
        setIsThinking(false)
      }
    },
    [department, messages],
  )

  const loadBrowseAll = useCallback(
    async (messageId: string, prompt: string) => {
      if (!department) return
      setIsThinking(true)
      try {
        const response = await apiFetch<AIRecommendationResponse>('/assistant/chat', {
          method: 'POST',
          body: JSON.stringify({ prompt: `${department}: ${prompt}` }),
        })
        setMessages((prev) =>
          prev.map((m) => (m.id === messageId ? { ...m, response } : m)),
        )
      } finally {
        setIsThinking(false)
      }
    },
    [department],
  )

  const restoreMessages = useCallback((msgs: ChatMessage[]) => {
    setMessages(msgs)
    setIsThinking(false)
  }, [])

  const clearChat = useCallback(() => {
    setMessages([])
    setIsThinking(false)
    setDepartment(null)
  }, [])

  const resetMessages = useCallback(() => {
    setMessages([])
    setIsThinking(false)
  }, [])

  return {
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
  }
}
