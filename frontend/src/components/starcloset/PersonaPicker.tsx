import { usePersona } from '@/context/PersonaContext'
import { PersonaCard } from '@/components/starcloset/PersonaCard'
import { Skeleton } from '@/components/ui/Skeleton'

// PersonaPicker renders the full stylist roster as a responsive grid. It is the
// shared surface used on the landing hero and inside the swap-stylist panel.
export function PersonaPicker({ onSelected }: { onSelected?: () => void }) {
  const { personas, loading } = usePersona()

  if (loading) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-56 rounded-3xl" />
        ))}
      </div>
    )
  }

  if (personas.length === 0) {
    return (
      <p className="rounded-2xl border border-dashed border-line p-8 text-center text-sm text-ink-500">
        No stylists are published yet. Onboard one from the partner portal.
      </p>
    )
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {personas.map((persona) => (
        <PersonaCard key={persona.slug} persona={persona} onSelected={onSelected} />
      ))}
    </div>
  )
}
