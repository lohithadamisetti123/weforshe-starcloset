import { useMemo, useState } from 'react'
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion'
import { ArrowLeft, ArrowRight, Check } from 'lucide-react'
import { LuxIcon, type LuxIconName } from '@/components/starcloset/LuxuryIcons'
import { LocationSearch } from '@/components/onboarding/LocationSearch'
import { cn } from '@/lib/utils'
import type { IndianCity } from '@/data/indianCities'
import type { Gender, UpdateProfileRequest, UserProfile } from '@/types'

// ProfileSetup — the mandatory post-signup onboarding: a premium five-step
// wizard that collects only what personalization needs (gender, styles,
// colours, occasions, categories — deliberately NO budget; that belongs to the
// stylist conversation). Purely a client flow over the existing
// PUT /profile/preferences contract: no backend, auth or schema change.

// UI gender choices → wire values. "Kids" and "Prefer not to say" send no
// gender signal at all — a wrong signal is worse than none, and the Kids
// journey picks Girls/Boys through its own department flow.
const GENDER_CHOICES: { label: string; sub: string; icon: LuxIconName; wire: Gender | '' }[] = [
  { label: 'Women', sub: 'Womenswear first', icon: 'heel', wire: 'Women' },
  { label: 'Men', sub: 'Menswear first', icon: 'hanger', wire: 'Men' },
  { label: 'Kids', sub: 'Shopping for little ones', icon: 'star', wire: '' },
  { label: 'Prefer not to say', sub: 'Show me everything', icon: 'sparkle', wire: '' },
]

const STYLES = ['Ethnic', 'Western', 'Casual', 'Formal', 'Streetwear', 'Minimal', 'Bohemian', 'Indo-Western', 'Athleisure', 'Party Wear']
const COLORS: { name: string; swatch: string; ring?: boolean }[] = [
  { name: 'Black', swatch: '#16161a' },
  { name: 'White', swatch: '#ffffff', ring: true },
  { name: 'Blue', swatch: '#2f5fe3' },
  { name: 'Pink', swatch: '#ec5f9e' },
  { name: 'Red', swatch: '#d33434' },
  { name: 'Green', swatch: '#2e9e5b' },
  { name: 'Yellow', swatch: '#e8c33c' },
  { name: 'Beige', swatch: '#d9c8a9' },
  { name: 'Brown', swatch: '#8b5e3c' },
  { name: 'Purple', swatch: '#7c4fd0' },
]
const OCCASIONS = ['Casual', 'Office', 'Party', 'Wedding', 'Festival', 'Travel', 'College', 'Date Night', 'Weekend', 'Traditional Events']
const CATEGORIES = ['Dresses', 'Kurta Sets', 'Sarees', 'Jeans', 'Tops', 'Shirts', 'Footwear', 'Accessories', 'Bags', 'Jewellery']

const CHAMPAGNE = 'linear-gradient(135deg,#F4E4B8 0%,#C9A24B 55%,#A9822F 100%)'

const STEPS = [
  { key: 'gender', title: 'Who are you shopping for?', hint: 'Choose one' },
  { key: 'location', title: 'Where are you located?', hint: 'Required' },
  { key: 'styles', title: 'Which styles feel like you?', hint: 'Pick as many as you like' },
  { key: 'colors', title: 'Colours you reach for', hint: 'Pick as many as you like' },
  { key: 'occasions', title: 'What do you dress for?', hint: 'Pick as many as you like' },
  { key: 'categories', title: 'What do you shop for most?', hint: 'Pick as many as you like' },
] as const

function SelectChip({ label, selected, onToggle, leading }: { label: string; selected: boolean; onToggle: () => void; leading?: React.ReactNode }) {
  return (
    <motion.button
      type="button"
      onClick={onToggle}
      whileHover={{ y: -2, scale: 1.03 }}
      whileTap={{ scale: 0.96 }}
      transition={{ type: 'spring', stiffness: 380, damping: 22 }}
      className={cn(
        'flex items-center gap-2 rounded-full border px-4 py-2.5 text-[13px] font-semibold tracking-tight backdrop-blur transition-colors',
        selected ? 'border-transparent text-ink-900' : 'border-line bg-white/75 text-ink-600 hover:border-[#C9A24B]/60 hover:text-ink-900',
      )}
      style={selected ? { background: 'linear-gradient(135deg,#FBF3DD,#EBD9A6)', boxShadow: 'inset 0 0 0 1px rgba(201,162,75,0.55), 0 10px 22px -12px rgba(169,130,47,0.6)' } : { boxShadow: '0 6px 16px -12px rgba(17,24,39,0.35)' }}
    >
      {leading}
      {label}
      <AnimatePresence>
        {selected && (
          <motion.span initial={{ scale: 0 }} animate={{ scale: 1 }} exit={{ scale: 0 }} transition={{ type: 'spring', stiffness: 400, damping: 18 }} className="grid h-4 w-4 place-items-center rounded-full bg-ink-900 text-white">
            <Check className="h-2.5 w-2.5" />
          </motion.span>
        )}
      </AnimatePresence>
    </motion.button>
  )
}

export function ProfileSetup({
  userName, saving, onSave, onComplete,
}: {
  userName?: string
  saving: boolean
  onSave: (req: UpdateProfileRequest) => Promise<UserProfile | null>
  onComplete: () => void
}) {
  const reduce = useReducedMotion() ?? false
  const [step, setStep] = useState(0)
  const [genderIdx, setGenderIdx] = useState<number | null>(null)
  const [city, setCity] = useState<IndianCity | null>(null)
  const [styles, setStyles] = useState<string[]>([])
  const [colors, setColors] = useState<string[]>([])
  const [occasions, setOccasions] = useState<string[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [phase, setPhase] = useState<'form' | 'saving' | 'done'>('form')
  const [error, setError] = useState(false)

  const isLast = step === STEPS.length - 1
  // Gender and location are required; the taste steps are free-form.
  const canContinue = step === 0 ? genderIdx !== null : step === 1 ? city !== null : true

  const toggle = (list: string[], v: string, set: (x: string[]) => void) =>
    set(list.includes(v) ? list.filter((x) => x !== v) : [...list, v])

  const firstName = useMemo(() => (userName ?? '').trim().split(/\s+/)[0] || null, [userName])

  const finish = async () => {
    setError(false)
    setPhase('saving')
    const saved = await onSave({
      // Preserve the account name — the profile update is a full replace.
      name: userName?.trim() || undefined,
      gender: genderIdx !== null ? GENDER_CHOICES[genderIdx].wire : '',
      // Plain city name: that's what the engine's location signal recognises.
      city: city?.city,
      favoriteStyles: styles,
      favoriteColors: colors,
      favoriteOccasions: occasions,
      favoriteCategories: categories,
      onboarded: true,
    })
    if (!saved) {
      setPhase('form')
      setError(true)
      return
    }
    setPhase('done')
    setTimeout(onComplete, reduce ? 200 : 1400)
  }

  return (
    <div className="relative flex min-h-full items-center justify-center px-4 py-10">
      {/* ambient blooms */}
      <div aria-hidden className="pointer-events-none fixed inset-0">
        <div className="absolute -left-24 top-16 h-96 w-96 rounded-full" style={{ background: 'radial-gradient(closest-side, rgba(255,63,108,0.12), transparent)' }} />
        <div className="absolute -right-24 bottom-10 h-[26rem] w-[26rem] rounded-full" style={{ background: 'radial-gradient(closest-side, rgba(201,162,75,0.14), transparent)' }} />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 26, scale: 0.97 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ type: 'spring', stiffness: 120, damping: 19 }}
        className="relative w-full max-w-2xl overflow-hidden rounded-[28px] border border-white/70 bg-white/80 backdrop-blur-2xl"
        style={{ boxShadow: '0 34px 80px -28px rgba(11,13,20,0.35), inset 0 0 0 1px rgba(255,255,255,0.6)' }}
      >
        {/* ── Header ── */}
        <div className="px-7 pt-7 text-center sm:px-10">
          <span className="mx-auto grid h-12 w-12 place-items-center rounded-full" style={{ background: CHAMPAGNE, boxShadow: '0 14px 30px -12px rgba(169,130,47,0.7), inset 0 0 0 2px rgba(255,255,255,0.5)' }}>
            <LuxIcon name="sparkle" className="h-5 w-5 text-white" strokeWidth={1.5} />
          </span>
          <h1 className="mt-3 font-display text-[22px] font-semibold tracking-tight text-ink-900 sm:text-2xl">
            {firstName ? `${firstName}, let's personalize your fashion journey` : "Let's Personalize Your Fashion Journey"}
          </h1>
          <p className="mt-1 text-[13px] text-ink-500">Help us understand your style to deliver better recommendations.</p>

          {/* ── Progress ── */}
          <div className="mx-auto mt-5 max-w-sm">
            <div className="flex items-center justify-between text-[10px] font-semibold uppercase tracking-[0.18em] text-ink-400">
              <span>Step {step + 1} of {STEPS.length}</span>
              <span className="text-[#A9822F]">{STEPS[step].hint}</span>
            </div>
            <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-ink-900/8">
              <motion.div className="h-full rounded-full" style={{ background: 'linear-gradient(90deg,#FF9EC5,#FF3F6C 45%,#E7B84B)' }} animate={{ width: `${((step + 1) / STEPS.length) * 100}%` }} transition={{ type: 'spring', stiffness: 140, damping: 22 }} />
            </div>
          </div>
        </div>

        {/* ── Step body ── */}
        <div className="min-h-[300px] px-7 pb-2 pt-6 sm:px-10">
          <AnimatePresence mode="wait">
            <motion.div
              key={step}
              initial={reduce ? { opacity: 0 } : { opacity: 0, x: 34 }}
              animate={{ opacity: 1, x: 0 }}
              exit={reduce ? { opacity: 0 } : { opacity: 0, x: -34 }}
              transition={{ type: 'spring', stiffness: 160, damping: 22 }}
            >
              <h2 className="font-display text-[16px] font-semibold tracking-tight text-ink-900">{STEPS[step].title}</h2>

              {step === 0 && (
                <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
                  {GENDER_CHOICES.map((g, i) => {
                    const selected = genderIdx === i
                    return (
                      <motion.button
                        key={g.label}
                        type="button"
                        onClick={() => setGenderIdx(i)}
                        whileHover={{ y: -3 }}
                        whileTap={{ scale: 0.97 }}
                        transition={{ type: 'spring', stiffness: 320, damping: 20 }}
                        className={cn('flex items-center gap-3.5 rounded-2xl border px-4 py-4 text-left backdrop-blur transition-colors', selected ? 'border-transparent' : 'border-line bg-white/75 hover:border-[#C9A24B]/60')}
                        style={selected ? { background: 'linear-gradient(135deg,#FBF3DD,#EBD9A6)', boxShadow: 'inset 0 0 0 1.5px rgba(201,162,75,0.6), 0 14px 30px -14px rgba(169,130,47,0.6)' } : { boxShadow: '0 8px 20px -14px rgba(17,24,39,0.4)' }}
                      >
                        <span className={cn('grid h-11 w-11 shrink-0 place-items-center rounded-full', selected ? 'text-white' : 'text-[#A9822F]')} style={{ background: selected ? CHAMPAGNE : 'rgba(201,162,75,0.12)' }}>
                          <LuxIcon name={g.icon} className="h-5 w-5" strokeWidth={1.5} />
                        </span>
                        <span className="flex-1">
                          <span className="block text-[14px] font-semibold tracking-tight text-ink-900">{g.label}</span>
                          <span className="block text-[11px] font-medium text-ink-500">{g.sub}</span>
                        </span>
                        <AnimatePresence>
                          {selected && (
                            <motion.span initial={{ scale: 0 }} animate={{ scale: 1 }} exit={{ scale: 0 }} transition={{ type: 'spring', stiffness: 400, damping: 18 }} className="grid h-5 w-5 place-items-center rounded-full bg-ink-900 text-white">
                              <Check className="h-3 w-3" />
                            </motion.span>
                          )}
                        </AnimatePresence>
                      </motion.button>
                    )
                  })}
                </div>
              )}

              {step === 1 && (
                <div className="mt-2">
                  <p className="mb-4 text-[12px] leading-relaxed text-ink-500">
                    This helps us personalize fashion recommendations, seasonal trends, delivery estimates, and regional styles.
                  </p>
                  <LocationSearch value={city} onChange={setCity} />
                </div>
              )}

              {step === 2 && (
                <div className="mt-4 flex flex-wrap gap-2.5">
                  {STYLES.map((s) => (
                    <SelectChip key={s} label={s} selected={styles.includes(s)} onToggle={() => toggle(styles, s, setStyles)} />
                  ))}
                </div>
              )}

              {step === 3 && (
                <div className="mt-4 flex flex-wrap gap-2.5">
                  {COLORS.map((c) => (
                    <SelectChip
                      key={c.name}
                      label={c.name}
                      selected={colors.includes(c.name)}
                      onToggle={() => toggle(colors, c.name, setColors)}
                      leading={<span className={cn('h-3.5 w-3.5 rounded-full', c.ring && 'ring-1 ring-ink-300')} style={{ background: c.swatch }} />}
                    />
                  ))}
                </div>
              )}

              {step === 4 && (
                <div className="mt-4 flex flex-wrap gap-2.5">
                  {OCCASIONS.map((o) => (
                    <SelectChip key={o} label={o} selected={occasions.includes(o)} onToggle={() => toggle(occasions, o, setOccasions)} />
                  ))}
                </div>
              )}

              {step === 5 && (
                <div className="mt-4 flex flex-wrap gap-2.5">
                  {CATEGORIES.map((c) => (
                    <SelectChip key={c} label={c} selected={categories.includes(c)} onToggle={() => toggle(categories, c, setCategories)} />
                  ))}
                </div>
              )}

              {error && <p className="mt-4 text-[12px] font-medium text-brand-600">We couldn’t save your preferences — please try again.</p>}
            </motion.div>
          </AnimatePresence>
        </div>

        {/* ── Footer ── */}
        <div className="flex items-center justify-between border-t border-line/70 px-7 py-4 sm:px-10">
          <button
            onClick={() => setStep((s) => Math.max(0, s - 1))}
            disabled={step === 0 || phase !== 'form'}
            className={cn('flex items-center gap-1.5 rounded-full px-4 py-2 text-[13px] font-semibold transition-colors', step === 0 ? 'cursor-not-allowed text-ink-300' : 'text-ink-600 hover:text-ink-900')}
          >
            <ArrowLeft className="h-4 w-4" /> Back
          </button>
          <div className="flex items-center gap-1.5" aria-hidden>
            {STEPS.map((_, i) => (
              <span key={i} className="h-1.5 rounded-full transition-all duration-300" style={{ width: i === step ? 18 : 6, background: i <= step ? '#C9A24B' : 'rgba(17,24,39,0.12)' }} />
            ))}
          </div>
          <motion.button
            onClick={() => (isLast ? finish() : setStep((s) => s + 1))}
            disabled={!canContinue || phase !== 'form'}
            whileHover={canContinue ? { scale: 1.04, y: -1 } : {}}
            whileTap={canContinue ? { scale: 0.96 } : {}}
            className={cn('flex items-center gap-1.5 rounded-full px-5 py-2.5 text-[13px] font-semibold tracking-tight text-white transition-opacity', !canContinue && 'opacity-40')}
            style={{ background: 'linear-gradient(135deg,#FF3F6C,#E7B84B)', boxShadow: canContinue ? '0 12px 26px -10px rgba(255,63,108,0.6)' : 'none' }}
          >
            {isLast ? 'Complete Setup' : 'Continue'}
            {isLast ? <Check className="h-4 w-4" /> : <ArrowRight className="h-4 w-4" />}
          </motion.button>
        </div>

        {/* ── Saving / done overlay ── */}
        <AnimatePresence>
          {phase !== 'form' && (
            <motion.div className="absolute inset-0 z-10 flex flex-col items-center justify-center bg-white/85 backdrop-blur-xl" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}>
              {phase === 'saving' || saving ? (
                <>
                  <span className="relative grid h-16 w-16 place-items-center rounded-full" style={{ background: CHAMPAGNE, boxShadow: '0 18px 40px -14px rgba(169,130,47,0.7)' }}>
                    <motion.span animate={reduce ? {} : { rotate: 360 }} transition={{ duration: 2.6, repeat: Infinity, ease: 'linear' }}>
                      <LuxIcon name="sparkle" className="h-7 w-7 text-white" strokeWidth={1.5} />
                    </motion.span>
                    {!reduce && <motion.span className="absolute inset-0 rounded-full" style={{ boxShadow: '0 0 0 3px rgba(201,162,75,0.5)' }} animate={{ scale: [1, 1.4], opacity: [0.7, 0] }} transition={{ duration: 1.6, repeat: Infinity, ease: 'easeOut' }} />}
                  </span>
                  <p className="mt-4 font-display text-[15px] font-semibold tracking-tight text-ink-900">Tailoring your experience…</p>
                  <p className="text-[11.5px] font-medium text-ink-500">Your stylists are taking notes</p>
                </>
              ) : (
                <>
                  <motion.span initial={{ scale: 0.4 }} animate={{ scale: 1 }} transition={{ type: 'spring', stiffness: 260, damping: 14 }} className="grid h-16 w-16 place-items-center rounded-full" style={{ background: CHAMPAGNE, boxShadow: '0 18px 40px -14px rgba(169,130,47,0.7)' }}>
                    <Check className="h-7 w-7 text-white" />
                  </motion.span>
                  <p className="mt-4 font-display text-[15px] font-semibold tracking-tight text-ink-900">You’re all set</p>
                  <p className="text-[11.5px] font-medium text-ink-500">Welcome to StarCloset</p>
                </>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </div>
  )
}
