import { motion } from 'framer-motion'
import { Avatar } from '@/components/ui/Avatar'

export function AILoader() {
  return (
    <div className="flex items-start gap-3">
      <Avatar variant="ai" />
      <div className="flex items-center gap-2 rounded-xl rounded-tl-sm border border-line bg-white px-4 py-3.5 shadow-sm">
        <span className="text-xs font-medium text-ink-500">Styling your look</span>
        <div className="flex gap-1">
          {[0, 1, 2].map((i) => (
            <motion.span
              key={i}
              className="h-1.5 w-1.5 rounded-full bg-brand-500"
              animate={{ y: [0, -5, 0], opacity: [0.4, 1, 0.4] }}
              transition={{ duration: 0.9, repeat: Infinity, delay: i * 0.15 }}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
