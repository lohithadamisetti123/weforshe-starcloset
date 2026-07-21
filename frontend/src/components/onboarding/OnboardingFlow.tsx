import { useEffect, useMemo, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { ArrowLeft, ArrowRight, Check, Sparkles, X } from 'lucide-react'
import { apiFetch } from '@/lib/api'
import {
  BUDGET_BANDS,
  CATEGORY_OPTIONS,
  CITY_OPTIONS,
  COLOR_OPTIONS,
  FALLBACK_BRANDS,
  GENDER_OPTIONS,
  STYLE_OPTIONS,
  type BudgetBand,
} from '@/data/personalizationOptions'
import type { Gender, UpdateProfileRequest, UserProfile } from '@/types'

interface OnboardingFlowProps {
  // initial pre-fills the wizard when editing an existing profile; omit for
  // first-time onboarding.
  initial?: UserProfile | null
  // mode changes the copy and controls only; the steps are identical, which is
  // the point — one flow serves onboarding and editing.
  mode?: 'onboarding' | 'edit'
  saving?: boolean
  onSave: (req: UpdateProfileRequest) => Promise<UserProfile | null>
  onDone: () => void
  // onSkip is offered only in onboarding mode — a shopper can always start
  // using the stylist and personalize later.
  onSkip?: () => void
}

// A single selectable pill used across the multi-select steps.
function Chip({
  label,
  selected,
  onToggle,
}: {
  label: string
  selected: boolean
  onToggle: () => void
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className={[
        'rounded-full border px-4 py-2 text-sm font-medium transition-all duration-150',
        selected
          ? 'border-brand-500 bg-brand-50 text-brand-600 shadow-sm'
          : 'border-line bg-white text-ink-600 hover:border-brand-300 hover:text-ink-900',
      ].join(' ')}
    >
      {label}
    </button>
  )
}

export function OnboardingFlow({
  initial,
  mode = 'onboarding',
  saving = false,
  onSave,
  onDone,
  onSkip,
}: OnboardingFlowProps) {
  const [step, setStep] = useState(0)

  // Form state, pre-filled from an existing profile when editing.
  const [name, setName] = useState(initial?.name ?? '')
  const [gender, setGender] = useState<Gender | ''>(initial?.gender ?? '')
  const [city, setCity] = useState(initial?.city ?? '')
  const [budget, setBudget] = useState<BudgetBand | null>(
    initial?.budgetMin != null && initial?.budgetMax != null
      ? BUDGET_BANDS.find((b) => b.min === initial.budgetMin && b.max === initial.budgetMax) ?? null
      : null,
  )
  const [styles, setStyles] = useState<string[]>(initial?.favoriteStyles ?? [])
  const [colors, setColors] = useState<string[]>(initial?.favoriteColors ?? [])
  const [categories, setCategories] = useState<string[]>(initial?.favoriteCategories ?? [])
  const [brands, setBrands] = useState<string[]>(initial?.favoriteBrands ?? [])

  // Live catalogue brands so a shopper's picks actually exist in the catalogue
  // (which is what makes the preferred-brand boost land). Falls back to a
  // curated list if the fetch fails, so the step is never empty.
  const [brandOptions, setBrandOptions] = useState<string[]>(FALLBACK_BRANDS)
  useEffect(() => {
    let alive = true
    apiFetch<string[]>('/brands')
      .then((list) => {
        if (alive && Array.isArray(list) && list.length > 0) {
          setBrandOptions(list.slice(0, 40))
        }
      })
      .catch(() => {
        /* keep the fallback list */
      })
    return () => {
      alive = false
    }
  }, [])

  const toggle = (list: string[], value: string, set: (v: string[]) => void) => {
    set(list.includes(value) ? list.filter((v) => v !== value) : [...list, value])
  }

  const buildRequest = (): UpdateProfileRequest => ({
    name: name.trim(),
    gender: gender || undefined,
    city: city.trim(),
    budgetMin: budget?.min,
    budgetMax: budget?.max,
    favoriteStyles: styles,
    favoriteColors: colors,
    favoriteBrands: brands,
    favoriteCategories: categories,
    onboarded: true,
  })

  const steps = useMemo(
    () => [
      {
        title: mode === 'edit' ? 'Your details' : "Let's get to know you",
        subtitle: 'A few quick taps and your stylist starts working for you.',
      },
      { title: "What's your everyday budget?", subtitle: 'Your prompts can always override this.' },
      { title: 'Which styles feel like you?', subtitle: 'Pick as many as you like.' },
      { title: 'Colours you reach for', subtitle: 'Pick as many as you like.' },
      { title: 'What do you shop for most?', subtitle: 'Pick as many as you like.' },
      { title: 'Any brands you love?', subtitle: 'These get a gentle boost — never a hard filter.' },
    ],
    [mode],
  )

  const lastStep = steps.length - 1
  const isLast = step === lastStep

  const next = async () => {
    if (!isLast) {
      setStep((s) => s + 1)
      return
    }
    const saved = await onSave(buildRequest())
    if (saved) onDone()
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="flex min-h-full items-center justify-center bg-surface px-4 py-10"
    >
      <div className="w-full max-w-2xl rounded-2xl border border-line bg-white p-6 shadow-lg sm:p-8">
        {/* Header */}
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <span className="flex h-11 w-11 items-center justify-center rounded-full bg-gradient-to-br from-brand-50 to-brand-100">
              <Sparkles className="h-5 w-5 text-brand-500" />
            </span>
            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-brand-500">
                {mode === 'edit' ? 'Edit preferences' : 'Personalize your stylist'}
              </p>
              <h2 className="font-display text-lg font-bold text-ink-900">{steps[step].title}</h2>
            </div>
          </div>
          {mode === 'onboarding' && onSkip && (
            <button
              onClick={onSkip}
              className="flex items-center gap-1 rounded-full px-3 py-1.5 text-xs font-medium text-ink-500 transition-colors hover:text-ink-900"
            >
              Skip for now
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {/* Progress */}
        <div className="mt-5 flex items-center gap-1.5">
          {steps.map((_, i) => (
            <div
              key={i}
              className={[
                'h-1.5 flex-1 rounded-full transition-colors duration-300',
                i <= step ? 'bg-brand-500' : 'bg-line',
              ].join(' ')}
            />
          ))}
        </div>

        <p className="mt-4 text-sm text-ink-500">{steps[step].subtitle}</p>

        {/* Step body */}
        <div className="mt-5 min-h-[220px]">
          <AnimatePresence mode="wait">
            <motion.div
              key={step}
              initial={{ opacity: 0, x: 12 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -12 }}
              transition={{ duration: 0.2 }}
            >
              {step === 0 && (
                <div className="space-y-5">
                  <label className="block">
                    <span className="mb-1.5 block text-sm font-medium text-ink-900">Name</span>
                    <input
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      placeholder="What should we call you?"
                      className="w-full rounded-xl border border-line bg-white px-4 py-2.5 text-sm text-ink-900 outline-none transition-colors focus:border-brand-500"
                    />
                  </label>

                  <div>
                    <span className="mb-1.5 block text-sm font-medium text-ink-900">Shopping for</span>
                    <div className="flex flex-wrap gap-2">
                      {GENDER_OPTIONS.map((g) => (
                        <Chip key={g} label={g} selected={gender === g} onToggle={() => setGender(gender === g ? '' : g)} />
                      ))}
                    </div>
                  </div>

                  <label className="block">
                    <span className="mb-1.5 block text-sm font-medium text-ink-900">City</span>
                    <input
                      list="city-options"
                      value={city}
                      onChange={(e) => setCity(e.target.value)}
                      placeholder="Where do you shop from?"
                      className="w-full rounded-xl border border-line bg-white px-4 py-2.5 text-sm text-ink-900 outline-none transition-colors focus:border-brand-500"
                    />
                    <datalist id="city-options">
                      {CITY_OPTIONS.map((c) => (
                        <option key={c} value={c} />
                      ))}
                    </datalist>
                  </label>
                </div>
              )}

              {step === 1 && (
                <div className="flex flex-col gap-2.5">
                  {BUDGET_BANDS.map((b) => {
                    const selected = budget?.label === b.label
                    return (
                      <button
                        key={b.label}
                        type="button"
                        onClick={() => setBudget(selected ? null : b)}
                        className={[
                          'flex items-center justify-between rounded-xl border px-4 py-3 text-left text-sm font-medium transition-all duration-150',
                          selected
                            ? 'border-brand-500 bg-brand-50 text-brand-600'
                            : 'border-line bg-white text-ink-700 hover:border-brand-300',
                        ].join(' ')}
                      >
                        {b.label}
                        {selected && <Check className="h-4 w-4" />}
                      </button>
                    )
                  })}
                </div>
              )}

              {step === 2 && (
                <div className="flex flex-wrap gap-2">
                  {STYLE_OPTIONS.map((s) => (
                    <Chip key={s} label={s} selected={styles.includes(s)} onToggle={() => toggle(styles, s, setStyles)} />
                  ))}
                </div>
              )}

              {step === 3 && (
                <div className="flex flex-wrap gap-2">
                  {COLOR_OPTIONS.map((c) => (
                    <Chip key={c} label={c} selected={colors.includes(c)} onToggle={() => toggle(colors, c, setColors)} />
                  ))}
                </div>
              )}

              {step === 4 && (
                <div className="flex flex-wrap gap-2">
                  {CATEGORY_OPTIONS.map((c) => (
                    <Chip
                      key={c}
                      label={c}
                      selected={categories.includes(c)}
                      onToggle={() => toggle(categories, c, setCategories)}
                    />
                  ))}
                </div>
              )}

              {step === 5 && (
                <div className="flex max-h-[240px] flex-wrap gap-2 overflow-y-auto">
                  {brandOptions.map((b) => (
                    <Chip key={b} label={b} selected={brands.includes(b)} onToggle={() => toggle(brands, b, setBrands)} />
                  ))}
                </div>
              )}
            </motion.div>
          </AnimatePresence>
        </div>

        {/* Footer controls */}
        <div className="mt-6 flex items-center justify-between border-t border-line pt-5">
          <button
            onClick={() => setStep((s) => Math.max(0, s - 1))}
            disabled={step === 0}
            className={[
              'flex items-center gap-1.5 rounded-full px-4 py-2 text-sm font-medium transition-colors',
              step === 0 ? 'cursor-not-allowed text-ink-300' : 'text-ink-600 hover:text-ink-900',
            ].join(' ')}
          >
            <ArrowLeft className="h-4 w-4" />
            Back
          </button>

          <span className="text-xs text-ink-400">
            {step + 1} / {steps.length}
          </span>

          <button
            onClick={next}
            disabled={saving}
            className="flex items-center gap-1.5 rounded-full bg-brand-500 px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition-all hover:bg-brand-600 disabled:opacity-60"
          >
            {isLast ? (saving ? 'Saving…' : mode === 'edit' ? 'Save' : 'Finish') : 'Next'}
            {!isLast && <ArrowRight className="h-4 w-4" />}
            {isLast && !saving && <Check className="h-4 w-4" />}
          </button>
        </div>
      </div>
    </motion.div>
  )
}
