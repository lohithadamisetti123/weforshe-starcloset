import { motion } from 'framer-motion'
import type { ChatMessage } from '@/types'
import { Avatar } from '@/components/ui/Avatar'
import { ResponseCard } from '@/components/assistant/ResponseCard'
import { CollectionsView } from '@/components/assistant/CollectionsView'
import { cn } from '@/lib/utils'

interface ChatBubbleProps {
  message: ChatMessage
  isThinking: boolean
  onLoadMore: (messageId: string, prompt: string) => void
  onBrowseAll: (messageId: string, prompt: string) => void
  onCartAdd?: () => void
}

export function ChatBubble({ message, isThinking, onLoadMore, onBrowseAll, onCartAdd }: ChatBubbleProps) {
  const isUser = message.role === 'user'
  const time = new Date(message.timestamp).toLocaleTimeString('en-IN', {
    hour: '2-digit',
    minute: '2-digit',
  })

  // Extract the original prompt from content for "More Collections" / "Browse All"
  const prompt = isUser
    ? message.content
    : message.content.replace(/^\d+ collections for "|"$/g, '')

  return (
    <motion.div
      initial={{ opacity: 0, y: 14 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35 }}
      className={cn('flex items-start gap-3', isUser && 'flex-row-reverse')}
    >
      <Avatar variant={isUser ? 'user' : 'ai'} />
      <div className={cn('flex max-w-[92%] flex-col gap-2', isUser && 'items-end')}>
        {/* The stylist's spoken line, in the persona's voice, above their picks. */}
        {!isUser && message.stylistReply && (
          <div className="rounded-xl rounded-tl-sm border border-brand-100 bg-gradient-to-br from-brand-50/70 to-white px-4 py-3 text-sm font-medium leading-relaxed text-ink-800 shadow-sm">
            {message.stylistReply}
          </div>
        )}

        {/* Multiple named collections */}
        {message.collections && message.collections.length > 0 ? (
          <CollectionsView
            collections={message.collections}
            response={message.response}
            messageId={message.id}
            prompt={prompt}
            isThinking={isThinking}
            onLoadMore={onLoadMore}
            onBrowseAll={onBrowseAll}
            onCartAdd={onCartAdd}
          />
        ) : message.response ? (
          <ResponseCard response={message.response} />
        ) : (
          <div
            className={cn(
              'rounded-xl px-4 py-3 text-sm leading-relaxed shadow-sm',
              isUser
                ? 'rounded-tr-sm bg-ink-900 text-white'
                : 'rounded-tl-sm border border-line bg-white text-ink-900',
            )}
          >
            {message.content}
          </div>
        )}
        <span className="px-1 text-[10px] text-ink-300">{time}</span>
      </div>
    </motion.div>
  )
}
