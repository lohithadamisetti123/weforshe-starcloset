interface PromptChipProps {
  label: string
  onClick: () => void
}

export function PromptChip({ label, onClick }: PromptChipProps) {
  return (
    <button
      onClick={onClick}
      className="shrink-0 whitespace-nowrap rounded-full border border-line bg-white px-4 py-2 text-sm font-medium text-ink-700 transition-all hover:-translate-y-0.5 hover:border-brand-500 hover:text-brand-600 hover:shadow-sm"
    >
      {label}
    </button>
  )
}
