import { useEffect, useMemo, useRef, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Check, MapPin, Search, X } from 'lucide-react'
import { INDIAN_CITIES, cityDisplay, type IndianCity } from '@/data/indianCities'
import { cn } from '@/lib/utils'

// LocationSearch — the onboarding location selector: a fast, keyboard-friendly
// autocomplete over the India city dataset (search by city OR state), with
// recent picks and a clear button. Pure UI: the chosen city is handed to the
// caller, which stores it through the existing profile `city` field.

const RECENTS_KEY = 'sc-recent-cities'
const MAX_RESULTS = 8

function readRecents(): IndianCity[] {
  try {
    const raw = JSON.parse(localStorage.getItem(RECENTS_KEY) || '[]') as IndianCity[]
    return Array.isArray(raw) ? raw.filter((r) => r && typeof r.city === 'string' && typeof r.state === 'string').slice(0, 5) : []
  } catch {
    return []
  }
}
function pushRecent(c: IndianCity) {
  const next = [c, ...readRecents().filter((r) => r.city !== c.city || r.state !== c.state)].slice(0, 5)
  try {
    localStorage.setItem(RECENTS_KEY, JSON.stringify(next))
  } catch {
    /* storage full/blocked — recents are a nicety */
  }
}

export function LocationSearch({ value, onChange }: { value: IndianCity | null; onChange: (c: IndianCity | null) => void }) {
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const [highlight, setHighlight] = useState(0)
  const [recents, setRecents] = useState<IndianCity[]>(() => readRecents())
  const inputRef = useRef<HTMLInputElement>(null)
  const rootRef = useRef<HTMLDivElement>(null)

  // Close on outside click.
  useEffect(() => {
    const onDoc = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [])

  // Fast in-memory filter: prefix matches on city rank first, then substring
  // matches on city, then state matches — searchable by either part.
  const results = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return []
    const starts: IndianCity[] = []
    const contains: IndianCity[] = []
    const byState: IndianCity[] = []
    for (const c of INDIAN_CITIES) {
      const city = c.city.toLowerCase()
      const state = c.state.toLowerCase()
      if (city.startsWith(q)) starts.push(c)
      else if (city.includes(q)) contains.push(c)
      else if (state.startsWith(q) || state.includes(q)) byState.push(c)
      if (starts.length >= MAX_RESULTS) break
    }
    return [...starts, ...contains, ...byState].slice(0, MAX_RESULTS)
  }, [query])

  const showRecents = query.trim() === '' && recents.length > 0
  const listed = showRecents ? recents : results

  useEffect(() => setHighlight(0), [query, open])

  const select = (c: IndianCity) => {
    onChange(c)
    pushRecent(c)
    setRecents(readRecents())
    setQuery('')
    setOpen(false)
  }

  const clear = () => {
    onChange(null)
    setQuery('')
    inputRef.current?.focus()
    setOpen(true)
  }

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (!open && (e.key === 'ArrowDown' || e.key === 'Enter')) {
      setOpen(true)
      return
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setHighlight((h) => Math.min(listed.length - 1, h + 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setHighlight((h) => Math.max(0, h - 1))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (listed[highlight]) select(listed[highlight])
    } else if (e.key === 'Escape') {
      setOpen(false)
    }
  }

  return (
    <div ref={rootRef} className="relative">
      {/* ── Search field / selected chip ── */}
      {value ? (
        <motion.div initial={{ opacity: 0, scale: 0.97 }} animate={{ opacity: 1, scale: 1 }} className="flex items-center justify-between rounded-full border border-transparent px-4 py-3" style={{ background: 'linear-gradient(135deg,#FBF3DD,#EBD9A6)', boxShadow: 'inset 0 0 0 1.5px rgba(201,162,75,0.6), 0 12px 26px -14px rgba(169,130,47,0.6)' }}>
          <span className="flex items-center gap-2.5">
            <span className="grid h-8 w-8 place-items-center rounded-full bg-white/70 text-[#A9822F]">
              <MapPin className="h-4 w-4" />
            </span>
            <span>
              <span className="block text-[13.5px] font-semibold tracking-tight text-ink-900">{value.city}</span>
              <span className="block text-[10.5px] font-medium text-ink-500">{value.state}</span>
            </span>
            <span className="ml-1 grid h-5 w-5 place-items-center rounded-full bg-ink-900 text-white"><Check className="h-3 w-3" /></span>
          </span>
          <button type="button" onClick={clear} aria-label="Change location" className="grid h-8 w-8 place-items-center rounded-full text-ink-500 transition-colors hover:bg-white/70 hover:text-ink-900">
            <X className="h-4 w-4" />
          </button>
        </motion.div>
      ) : (
        <div className="relative">
          <Search className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-ink-400" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => { setQuery(e.target.value); setOpen(true) }}
            onFocus={() => setOpen(true)}
            onKeyDown={onKeyDown}
            placeholder="Search your city or state…"
            role="combobox"
            aria-expanded={open}
            aria-autocomplete="list"
            className="w-full rounded-full border border-line bg-white/80 py-3 pl-11 pr-10 text-sm text-ink-900 shadow-[0_8px_22px_-14px_rgba(17,24,39,0.4)] backdrop-blur outline-none transition-all placeholder:text-ink-300 focus:border-brand-300 focus:shadow-[0_10px_28px_-12px_rgba(255,63,108,0.3)]"
          />
          <AnimatePresence>
            {query && (
              <motion.button initial={{ opacity: 0, scale: 0.6 }} animate={{ opacity: 1, scale: 1 }} exit={{ opacity: 0, scale: 0.6 }} type="button" onClick={() => { setQuery(''); inputRef.current?.focus() }} aria-label="Clear search" className="absolute right-3 top-1/2 grid h-6 w-6 -translate-y-1/2 place-items-center rounded-full bg-ink-900/6 text-ink-500 hover:text-ink-900">
                <X className="h-3.5 w-3.5" />
              </motion.button>
            )}
          </AnimatePresence>
        </div>
      )}

      {/* ── Dropdown ── */}
      <AnimatePresence>
        {open && !value && listed.length > 0 && (
          <motion.ul
            initial={{ opacity: 0, y: -6, scale: 0.985 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -6, scale: 0.985 }}
            transition={{ type: 'spring', stiffness: 320, damping: 26 }}
            role="listbox"
            className="absolute inset-x-0 z-30 mt-2 max-h-72 overflow-y-auto rounded-2xl border border-white/70 bg-white/90 p-1.5 backdrop-blur-2xl"
            style={{ boxShadow: '0 24px 60px -20px rgba(11,13,20,0.35), inset 0 0 0 1px rgba(255,255,255,0.6)' }}
          >
            {showRecents && <li className="px-3 pb-1 pt-1.5 text-[9px] font-bold uppercase tracking-[0.2em] text-ink-400">Recent</li>}
            {listed.map((c, i) => (
              <li key={`${c.city}|${c.state}`} role="option" aria-selected={i === highlight}>
                <button
                  type="button"
                  onMouseEnter={() => setHighlight(i)}
                  onClick={() => select(c)}
                  className={cn('flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left transition-colors', i === highlight ? 'bg-brand-50' : 'hover:bg-ink-900/4')}
                >
                  <span className={cn('grid h-7 w-7 shrink-0 place-items-center rounded-full', i === highlight ? 'bg-white text-brand-600' : 'bg-ink-900/5 text-ink-500')}>
                    <MapPin className="h-3.5 w-3.5" />
                  </span>
                  <span className="min-w-0">
                    <span className="block truncate text-[13px] font-semibold tracking-tight text-ink-900">{c.city}</span>
                    <span className="block truncate text-[10.5px] font-medium text-ink-500">{c.state}</span>
                  </span>
                </button>
              </li>
            ))}
          </motion.ul>
        )}
        {open && !value && !showRecents && query.trim() !== '' && results.length === 0 && (
          <motion.div initial={{ opacity: 0, y: -6 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -6 }} className="absolute inset-x-0 z-30 mt-2 rounded-2xl border border-white/70 bg-white/90 px-4 py-3.5 text-[12.5px] font-medium text-ink-500 backdrop-blur-2xl" style={{ boxShadow: '0 24px 60px -20px rgba(11,13,20,0.35)' }}>
            No matching city — try another spelling or a nearby metro.
          </motion.div>
        )}
      </AnimatePresence>
      <span className="sr-only" aria-live="polite">{value ? `Selected ${cityDisplay(value)}` : ''}</span>
    </div>
  )
}
