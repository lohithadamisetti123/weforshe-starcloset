import { useState, useCallback } from 'react'
import { Clock } from 'lucide-react'
import { ChatWindow } from '@/components/assistant/ChatWindow'
import { HistorySidebar } from '@/components/assistant/HistorySidebar'
import { OnboardingFlow } from '@/components/onboarding/OnboardingFlow'
import { useProfile } from '@/hooks/useProfile'
import { useShop } from '@/context/AppContext'
import type { ChatHistoryEntry } from '@/types'

export function AssistantPage() {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const { chatHistory } = useShop()
  const [reopenEntry, setReopenEntry] = useState<ChatHistoryEntry | null>(null)

  // Onboarding gate: shown once, before the stylist, for a shopper who has not
  // completed it. `dismissed` lets them skip for this session without blocking
  // shopping. While the status is still loading (onboarded === null) we render
  // the stylist rather than a spinner, so a returning shopper sees no flash.
  const { onboarded, saving, save } = useProfile()
  const [dismissed, setDismissed] = useState(false)
  const showOnboarding = onboarded === false && !dismissed

  if (showOnboarding) {
    return (
      <div className="min-h-[calc(100vh-4rem)] bg-surface">
        <OnboardingFlow
          mode="onboarding"
          saving={saving}
          onSave={save}
          onDone={() => setDismissed(true)}
          onSkip={() => setDismissed(true)}
        />
      </div>
    )
  }

  return (
    <div className="relative bg-surface">
      {/* History toggle button — floating over the chat */}
      <button
        onClick={() => setSidebarOpen(true)}
        className="absolute left-4 top-4 z-30 flex items-center gap-1.5 rounded-full border border-line bg-white px-3 py-2 text-xs font-semibold text-ink-600 shadow-sm transition-colors hover:border-ink-900 hover:text-ink-900"
      >
        <Clock className="h-3.5 w-3.5" />
        History
        {chatHistory.length > 0 && (
          <span className="flex h-4 w-4 items-center justify-center rounded-full bg-brand-500 text-[9px] font-bold text-white">
            {chatHistory.length > 9 ? '9+' : chatHistory.length}
          </span>
        )}
      </button>

      <ChatWindow reopenEntry={reopenEntry} onReopenHandled={() => setReopenEntry(null)} />

      <HistorySidebar
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onReopen={handleReopen}
      />
    </div>
  )
}
