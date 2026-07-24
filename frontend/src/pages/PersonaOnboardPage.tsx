import { useState, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, ArrowRight, Check, Sparkles, Rocket } from 'lucide-react'
import { apiFetch } from '@/lib/api'
import { Button } from '@/components/ui/Button'
import { StylistAvatar } from '@/components/starcloset/StylistAvatar'
import type { PersonaSchema, PersonaDetail, OnboardPersonaPayload } from '@/types/persona'

// PersonaOnboardPage is the partner portal wizard (/partner/onboard). Partner
// authentication is out of scope for this MVP, so it is public. It walks a
// partner through the three persona layers and publishes the result — the same
// stylist a shopper can then pick in the studio.

interface FormState {
  displayName: string
  partnerName: string
  tagline: string
  greeting: string
  portraitUrl: string
  accentFrom: string
  accentTo: string
  unlockAtItems: number
  samplePrompts: string
  styleSignature: string
  preferredStyles: string
  preferredBrands: string
  preferredColors: string
  styleOverride: string
  defaultGender: string
  voiceTone: string
  greetingTemplates: string
  responseTemplates: string
  catchphrases: string
  styleRules: string
  avoid: string
  exampleResponses: string
}

const EMPTY: FormState = {
  displayName: '',
  partnerName: 'StarCloset Partner',
  tagline: '',
  greeting: '',
  portraitUrl: '',
  accentFrom: '#FF7EB3',
  accentTo: '#1F1C18',
  unlockAtItems: 0,
  samplePrompts: '',
  styleSignature: '',
  preferredStyles: '',
  preferredBrands: '',
  preferredColors: '',
  styleOverride: '',
  defaultGender: 'Women',
  voiceTone: '',
  greetingTemplates: '',
  responseTemplates: '',
  catchphrases: '',
  styleRules: '',
  avoid: '',
  exampleResponses: '',
}

const STEPS = ['Celebrity Identity', 'Style DNA', 'Persona Voice', 'Launch']

function lines(s: string): string[] {
  return s
    .split('\n')
    .map((l) => l.trim())
    .filter(Boolean)
}

function buildPayload(f: FormState): OnboardPersonaPayload {
  return {
    displayName: f.displayName.trim(),
    partnerName: f.partnerName.trim(),
    tagline: f.tagline.trim(),
    greeting: f.greeting.trim(),
    portraitUrl: f.portraitUrl.trim(),
    accentFrom: f.accentFrom,
    accentTo: f.accentTo,
    unlockAtItems: Number(f.unlockAtItems) || 0,
    samplePrompts: lines(f.samplePrompts),
    styleDna: {
      styleSignature: f.styleSignature.trim(),
      preferredStyles: lines(f.preferredStyles),
      preferredBrands: lines(f.preferredBrands),
      preferredColors: lines(f.preferredColors),
      styleOverride: f.styleOverride,
      defaultGender: f.defaultGender,
    },
    voice: {
      voiceTone: f.voiceTone.trim(),
      greetingTemplates: lines(f.greetingTemplates),
      responseTemplates: lines(f.responseTemplates),
      catchphrases: lines(f.catchphrases),
      styleRules: lines(f.styleRules),
      avoid: lines(f.avoid),
      exampleResponses: lines(f.exampleResponses),
    },
  }
}

// previewReply approximates the backend persona voice so the partner can see how
// a reply will read before publishing. It mirrors the server's deterministic
// assembly (greeting + response template with placeholders filled).
function previewReply(f: FormState): string {
  const greet = lines(f.greetingTemplates)[0] || f.greeting || ''
  const tmpl =
    lines(f.responseTemplates)[0] ||
    "I've put together a few looks for your {occasion}, all within {budget}."
  const body = tmpl
    .replace(/{occasion}/g, 'sangeet')
    .replace(/{count}/g, '3')
    .replace(/{budget}/g, '₹8,000')
    .replace(/{department}/g, f.defaultGender)
    .replace(/{highlight}/g, 'Maroon Velvet Lehenga')
    .replace(/{pieces}/g, 'Outfit, Footwear, Jewellery')
  return [greet, body].filter(Boolean).join(' ')
}

// ── small field primitives ───────────────────────────────────────────────────

function Field({ label, help, children }: { label: string; help?: string; children: ReactNode }) {
  return (
    <label className="block">
      <span className="text-sm font-semibold text-ink-800">{label}</span>
      {help && <span className="ml-2 text-xs text-ink-400">{help}</span>}
      <div className="mt-1.5">{children}</div>
    </label>
  )
}

const inputCls =
  'w-full rounded-xl border border-line bg-white px-3.5 py-2.5 text-sm text-ink-900 outline-none focus:border-brand-500'

export function PersonaOnboardPage() {
  const [step, setStep] = useState(0)
  const [form, setForm] = useState<FormState>(EMPTY)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [published, setPublished] = useState<PersonaDetail | null>(null)
  const queryClient = useQueryClient()

  const { data: schema } = useQuery<PersonaSchema>({
    queryKey: ['persona-schema'],
    queryFn: () => apiFetch<PersonaSchema>('/partner/personas/schema'),
    staleTime: 60 * 60 * 1000,
  })

  const styleOptions = schema?.styleOptions ?? ['Ethnic', 'Western', 'Formal', 'Casual', 'Indo-Western', 'Athleisure']
  const genderOptions = schema?.genderOptions ?? ['Women', 'Men', 'Girls', 'Boys', 'Unisex']

  const set = <K extends keyof FormState>(key: K, value: FormState[K]) =>
    setForm((f) => ({ ...f, [key]: value }))

  const canProceed = step > 0 || form.displayName.trim().length >= 2

  const handlePublish = async () => {
    setBusy(true)
    setError(null)
    try {
      const created = await apiFetch<PersonaDetail>('/partner/personas', {
        method: 'POST',
        body: JSON.stringify(buildPayload(form)),
      })
      const live = await apiFetch<PersonaDetail>(`/partner/personas/${created.slug}/publish`, {
        method: 'POST',
      })
      await queryClient.invalidateQueries({ queryKey: ['stylist-personas'] })
      setPublished(live)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not publish the stylist.')
    } finally {
      setBusy(false)
    }
  }

  // ── Success screen ─────────────────────────────────────────────────────────
  if (published) {
    return (
      <div className="mx-auto max-w-2xl px-5 py-20 text-center lg:px-8">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-emerald-50">
          <Check className="h-8 w-8 text-emerald-600" />
        </div>
        <h1 className="mt-6 font-display text-2xl font-bold text-ink-900">
          {published.displayName} is live!
        </h1>
        <p className="mt-2 text-sm text-ink-500">
          Your stylist is published and shoppers can pick them in the studio.
        </p>
        <div className="mt-8 flex justify-center gap-3">
          <Link to="/assistant">
            <Button size="lg">
              <Sparkles className="h-4 w-4" /> Open Stylist Studio
            </Button>
          </Link>
          <Button
            variant="outline"
            size="lg"
            onClick={() => {
              setPublished(null)
              setForm(EMPTY)
              setStep(0)
            }}
          >
            Onboard another
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-3xl px-5 py-10 lg:px-8">
      <div className="text-center">
        <span className="inline-flex items-center gap-1.5 rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-600">
          <Rocket className="h-3.5 w-3.5" /> Partner Portal
        </span>
        <h1 className="mt-4 font-display text-3xl font-bold text-ink-900">Onboard a Celebrity Stylist</h1>
        <p className="mt-2 text-sm text-ink-500">
          Configure identity, Style DNA and voice — the engine stays rule-based and explainable.
        </p>
      </div>

      {/* Stepper */}
      <div className="mt-8 flex items-center justify-center gap-2">
        {STEPS.map((label, i) => (
          <div key={label} className="flex items-center gap-2">
            <div
              className={`flex h-8 items-center gap-2 rounded-full px-3 text-xs font-semibold ${
                i === step ? 'bg-ink-900 text-white' : i < step ? 'bg-emerald-50 text-emerald-600' : 'bg-ink-100 text-ink-400'
              }`}
            >
              {i < step ? <Check className="h-3.5 w-3.5" /> : <span>{i + 1}</span>}
              <span className="hidden sm:inline">{label}</span>
            </div>
            {i < STEPS.length - 1 && <div className="h-px w-4 bg-line" />}
          </div>
        ))}
      </div>

      <div className="mt-8 rounded-3xl border border-line bg-white p-6 lg:p-8">
        {/* Step 1 — Identity */}
        {step === 0 && (
          <div className="grid gap-5">
            <Field label="Display name" help="required">
              <input className={inputCls} value={form.displayName} onChange={(e) => set('displayName', e.target.value)} placeholder="e.g. Alia Bhatt" />
            </Field>
            <div className="grid gap-5 sm:grid-cols-2">
              <Field label="Partner / brand">
                <input className={inputCls} value={form.partnerName} onChange={(e) => set('partnerName', e.target.value)} />
              </Field>
              <Field label="Unlock at cart items" help="0 = from the start">
                <input type="number" min={0} className={inputCls} value={form.unlockAtItems} onChange={(e) => set('unlockAtItems', Number(e.target.value))} />
              </Field>
            </div>
            <Field label="Tagline">
              <input className={inputCls} value={form.tagline} onChange={(e) => set('tagline', e.target.value)} placeholder="One line that sells the vibe" />
            </Field>
            <Field label="Greeting">
              <textarea className={inputCls} rows={2} value={form.greeting} onChange={(e) => set('greeting', e.target.value)} placeholder="The first thing the stylist says" />
            </Field>
            <div className="grid gap-5 sm:grid-cols-3">
              <Field label="Portrait URL">
                <input className={inputCls} value={form.portraitUrl} onChange={(e) => set('portraitUrl', e.target.value)} placeholder="/images/stylists/…" />
              </Field>
              <Field label="Accent from">
                <input type="color" className="h-10 w-full rounded-xl border border-line" value={form.accentFrom} onChange={(e) => set('accentFrom', e.target.value)} />
              </Field>
              <Field label="Accent to">
                <input type="color" className="h-10 w-full rounded-xl border border-line" value={form.accentTo} onChange={(e) => set('accentTo', e.target.value)} />
              </Field>
            </div>
            <Field label="Sample prompts" help="one per line">
              <textarea className={inputCls} rows={4} value={form.samplePrompts} onChange={(e) => set('samplePrompts', e.target.value)} placeholder={'Sangeet look under ₹8000\nBrunch outfit for the weekend'} />
            </Field>
          </div>
        )}

        {/* Step 2 — Style DNA */}
        {step === 1 && (
          <div className="grid gap-5">
            <p className="rounded-xl bg-brand-50/60 px-4 py-3 text-xs text-ink-600">
              Style DNA biases the rule-based ranking engine. No AI ever picks products — this only nudges what surfaces.
            </p>
            <Field label="Style signature">
              <textarea className={inputCls} rows={2} value={form.styleSignature} onChange={(e) => set('styleSignature', e.target.value)} placeholder="Describe the aesthetic in a sentence" />
            </Field>
            <div className="grid gap-5 sm:grid-cols-2">
              <Field label="Style override" help="used when the shopper names no style">
                <select className={inputCls} value={form.styleOverride} onChange={(e) => set('styleOverride', e.target.value)}>
                  <option value="">— none —</option>
                  {styleOptions.map((s) => (
                    <option key={s} value={s}>{s}</option>
                  ))}
                </select>
              </Field>
              <Field label="Default gender">
                <select className={inputCls} value={form.defaultGender} onChange={(e) => set('defaultGender', e.target.value)}>
                  {genderOptions.map((g) => (
                    <option key={g} value={g}>{g}</option>
                  ))}
                </select>
              </Field>
            </div>
            <div className="grid gap-5 sm:grid-cols-3">
              <Field label="Preferred styles" help="one per line">
                <textarea className={inputCls} rows={4} value={form.preferredStyles} onChange={(e) => set('preferredStyles', e.target.value)} placeholder={'Ethnic\nElegant'} />
              </Field>
              <Field label="Preferred brands" help="one per line">
                <textarea className={inputCls} rows={4} value={form.preferredBrands} onChange={(e) => set('preferredBrands', e.target.value)} />
              </Field>
              <Field label="Preferred colours" help="one per line">
                <textarea className={inputCls} rows={4} value={form.preferredColors} onChange={(e) => set('preferredColors', e.target.value)} placeholder={'maroon\nred'} />
              </Field>
            </div>
          </div>
        )}

        {/* Step 3 — Voice */}
        {step === 2 && (
          <div className="grid gap-5">
            <Field label="Voice tone">
              <input className={inputCls} value={form.voiceTone} onChange={(e) => set('voiceTone', e.target.value)} placeholder="e.g. warm, bubbly, playful" />
            </Field>
            <Field label="Greeting templates" help="one per line — opens each reply">
              <textarea className={inputCls} rows={3} value={form.greetingTemplates} onChange={(e) => set('greetingTemplates', e.target.value)} placeholder={'Ooh, I love this brief!\nOkay, this is going to be fun —'} />
            </Field>
            <Field label="Response templates" help="{occasion} {count} {budget} {highlight} {pieces}">
              <textarea className={inputCls} rows={4} value={form.responseTemplates} onChange={(e) => set('responseTemplates', e.target.value)} placeholder={'For your {occasion}, I pulled together {count} looks — the {highlight} is my favourite. All within {budget}.'} />
            </Field>
            <div className="grid gap-5 sm:grid-cols-2">
              <Field label="Catchphrases" help="one per line">
                <textarea className={inputCls} rows={3} value={form.catchphrases} onChange={(e) => set('catchphrases', e.target.value)} />
              </Field>
              <Field label="Style rules" help="one per line">
                <textarea className={inputCls} rows={3} value={form.styleRules} onChange={(e) => set('styleRules', e.target.value)} />
              </Field>
            </div>
            <div className="grid gap-5 sm:grid-cols-2">
              <Field label="Avoid" help="one per line">
                <textarea className={inputCls} rows={2} value={form.avoid} onChange={(e) => set('avoid', e.target.value)} />
              </Field>
              <Field label="Example responses" help="shown in preview">
                <textarea className={inputCls} rows={2} value={form.exampleResponses} onChange={(e) => set('exampleResponses', e.target.value)} />
              </Field>
            </div>
          </div>
        )}

        {/* Step 4 — Launch / preview */}
        {step === 3 && (
          <div className="grid gap-6">
            <div className="flex items-center gap-4">
              <StylistAvatar
                persona={{ displayName: form.displayName || 'New Stylist', portraitUrl: form.portraitUrl, accentFrom: form.accentFrom, accentTo: form.accentTo }}
                size="xl"
              />
              <div>
                <h3 className="font-display text-xl font-bold text-ink-900">{form.displayName || 'New Stylist'}</h3>
                <p className="text-sm text-ink-500">{form.tagline || 'No tagline yet'}</p>
                <p className="mt-1 text-xs font-semibold text-brand-600">
                  Unlocks at {form.unlockAtItems} cart item{form.unlockAtItems === 1 ? '' : 's'} · styling {form.defaultGender}
                </p>
              </div>
            </div>

            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-ink-400">Greeting preview</p>
              <div className="mt-2 rounded-xl border border-line bg-surface px-4 py-3 text-sm text-ink-800">
                {form.greeting || 'No greeting set.'}
              </div>
            </div>

            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-ink-400">Sample conversation</p>
              <div className="mt-2 space-y-2">
                <div className="ml-auto w-fit max-w-[80%] rounded-xl rounded-tr-sm bg-ink-900 px-4 py-2.5 text-sm text-white">
                  Sangeet look under ₹8000
                </div>
                <div className="w-fit max-w-[85%] rounded-xl rounded-tl-sm border border-brand-100 bg-brand-50/60 px-4 py-2.5 text-sm text-ink-800">
                  {previewReply(form)}
                </div>
              </div>
            </div>

            {error && (
              <div className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{error}</div>
            )}
          </div>
        )}
      </div>

      {/* Nav buttons */}
      <div className="mt-6 flex items-center justify-between">
        <Button
          variant="ghost"
          onClick={() => (step === 0 ? undefined : setStep((s) => s - 1))}
          disabled={step === 0}
        >
          <ArrowLeft className="h-4 w-4" /> Back
        </Button>

        {step < STEPS.length - 1 ? (
          <Button onClick={() => setStep((s) => s + 1)} disabled={!canProceed}>
            Next <ArrowRight className="h-4 w-4" />
          </Button>
        ) : (
          <Button onClick={handlePublish} disabled={busy || form.displayName.trim().length < 2}>
            <Rocket className="h-4 w-4" /> {busy ? 'Publishing…' : 'Publish stylist'}
          </Button>
        )}
      </div>
    </div>
  )
}
