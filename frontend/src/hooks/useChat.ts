import { useCallback, useState } from 'react'
import { useShop } from '@/context/AppContext'
import { useSession } from '@/context/SessionContext'
import { useStarCloset } from '@/context/StarClosetContext'
import { capExclusions } from '@/context/sessionState'
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
    'Wedding collection under ₹10000',
    'Party collection under ₹3000',
    'Office wear under ₹5000',
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
  'Kids - Girls' : [
    'Birthday party dress',
    'School uniform collection',
    'Casual summer outfits',
    'Festive lehenga set',
    'Dance recital costume',
  ],
  'Kids - Boys' : [
    'School uniform collection',
    'Casual summer outfits',
    'Sportswear for kids',
    'Birthday party outfit',
    'Festive kurta set',
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
// excludeIds carries products the shopper has already been shown. It serves two
// jobs: "More Collections" asks for sets drawn from products not yet on screen,
// and — since the session now accumulates ids across the whole conversation —
// turn N no longer re-recommends what turn N-1 already offered.
async function fetchCollections(
  department: string,
  prompt: string,
  personaSlug: string | null,
  excludeIds: string[] = [],
): Promise<{ collections: CollectionResponse[]; stylistReply: string }> {
  const res = await apiFetch<CollectionsResponse>('/assistant/collections', {
    method: 'POST',
    body: JSON.stringify({
      department,
      prompt,
      ...(personaSlug ? { personaSlug } : {}),
      ...(excludeIds.length > 0 ? { excludeProductIds: excludeIds } : {}),
    }),
  })
  return { collections: res.collections ?? [], stylistReply: res.stylistReply ?? '' }
}

export function useChat() {
  const { addChatHistory } = useShop()
  const { selectedSlug } = useStarCloset()
  const {
    department,
    selectDepartment: selectSessionDepartment,
    shownProductIds,
    recordTurn,
    noteShown,
    resetConversation,
    clearSession,
  } = useSession()

  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [isThinking, setIsThinking] = useState(false)

  // Department now lives in the session, not in this hook. Switching department
  // still clears the on-screen transcript, but the department itself survives a
  // refresh or a trip to /cart and back.
  const selectDepartment = useCallback(
    (dept: string) => {
      selectSessionDepartment(dept)
      setMessages([])
    },
    [selectSessionDepartment],
  )

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
        // Everything already shown in this department is excluded, so the
        // assistant behaves like someone who remembers what they just suggested.
        const { collections, stylistReply } = await fetchCollections(
          department,
          trimmed,
          selectedSlug,
          capExclusions(shownProductIds),
        )

        if (collections.length > 0) {
          addChatHistory(department, trimmed, collections)
        }
        // Recorded even when empty, so an exhausted search still advances the
        // conversation record rather than silently vanishing.
        recordTurn(trimmed, collections)

        const assistantMessage: ChatMessage = {
          id: makeId(),
          role: 'assistant',
          content:
            collections.length > 0
              ? `${collections.length} collections for "${trimmed}"`
              : `I couldn't build anything new for "${trimmed}" that you haven't already seen. Try a different occasion, or widen the budget.`,
          timestamp: new Date().toISOString(),
          collections,
          stylistReply,
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
    [department, addChatHistory, shownProductIds, recordTurn, selectedSlug],
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
        const visibleIds = current.flatMap((c) => c.items.map((it) => it.product.id))
        const shownSigs = new Set(current.map(sig))

        // Ids visible in this message come first so they always survive the
        // cap — they are the repeats the shopper would notice immediately.
        const { collections: more } = await fetchCollections(
          department,
          prompt,
          selectedSlug,
          capExclusions(visibleIds, shownProductIds),
        )
        const fresh = more.filter((c) => c.items.length > 0 && !shownSigs.has(sig(c)))

        if (fresh.length > 0) {
          noteShown(fresh.flatMap((c) => c.items.map((it) => it.product.id)))
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
    [department, messages, shownProductIds, noteShown, selectedSlug],
  )

  const loadBrowseAll = useCallback(
    async (messageId: string, prompt: string) => {
      if (!department) return
      setIsThinking(true)
      try {
        const response = await apiFetch<AIRecommendationResponse>('/assistant/chat', {
          method: 'POST',
          body: JSON.stringify({
            prompt: `${department}: ${prompt}`,
            ...(selectedSlug ? { personaSlug: selectedSlug } : {}),
          }),
        })
        noteShown((response.products ?? []).map((p) => p.id))
        setMessages((prev) =>
          prev.map((m) => (m.id === messageId ? { ...m, response } : m)),
        )
      } catch (err) {
        // Previously this path had no catch at all: a failed request left the
        // thinking indicator to be cleared by `finally` with no explanation on
        // screen, so the button simply looked broken.
        setMessages((prev) => [
          ...prev,
          {
            id: makeId(),
            role: 'assistant',
            content:
              err instanceof Error
                ? `Couldn't load the full list: ${err.message}`
                : "Couldn't load the full list. Please try again.",
            timestamp: new Date().toISOString(),
          },
        ])
      } finally {
        setIsThinking(false)
      }
    },
    [department, noteShown, selectedSlug],
  )

  const restoreMessages = useCallback(
    (msgs: ChatMessage[]) => {
      setMessages(msgs)
      setIsThinking(false)
      // Reopening a past conversation puts its products back on screen, so they
      // count as shown — otherwise the next turn could re-recommend exactly what
      // the shopper is looking at.
      noteShown(
        msgs.flatMap((m) =>
          (m.collections ?? []).flatMap((c) => c.items.map((it) => it.product.id)),
        ),
      )
    },
    [noteShown],
  )

  // "Change" — back to the department picker. Drops the whole session so the
  // next department starts genuinely fresh.
  const clearChat = useCallback(() => {
    setMessages([])
    setIsThinking(false)
    clearSession()
  }, [clearSession])

  // "New Chat" — same department, clean slate. Clears the conversation record
  // and the shown-product memory for this department, so the shopper gets the
  // best matches again rather than page two of what they already rejected.
  const resetMessages = useCallback(() => {
    setMessages([])
    setIsThinking(false)
    resetConversation()
  }, [resetConversation])

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
