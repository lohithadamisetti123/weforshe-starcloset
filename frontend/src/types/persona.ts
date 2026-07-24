// StarCloset persona types. These mirror the backend dto/persona.go wire shapes
// exactly (the running JSON is the contract): StylistPersona is the shopper
// view, PersonaDetail is the full partner view, OnboardPersonaPayload is what
// the onboarding wizard POSTs, and PersonaSchema drives the wizard's fields.

// StylistPersona is the shopper-facing view returned by GET /assistant/personas.
export interface StylistPersona {
  slug: string
  displayName: string
  tagline: string
  greeting: string
  portraitUrl: string
  accentFrom: string
  accentTo: string
  unlockAtItems: number
  samplePrompts: string[]
  styleSignature: string
  defaultGender: string
  status: string
}

// PersonaStyleDNA is Layer 2 — the signature the rule-based engine reads.
export interface PersonaStyleDNA {
  styleSignature: string
  preferredStyles: string[]
  preferredBrands: string[]
  preferredColors: string[]
  styleOverride: string
  defaultGender: string
}

// PersonaVoice is Layer 3 — the deterministic template material.
export interface PersonaVoice {
  voiceTone: string
  greetingTemplates: string[]
  catchphrases: string[]
  styleRules: string[]
  avoid: string[]
  exampleResponses: string[]
  responseTemplates: string[]
}

// PersonaDetail is the full partner-facing persona (GET /partner/personas/:slug).
export interface PersonaDetail {
  slug: string
  displayName: string
  partnerName: string
  tagline: string
  greeting: string
  portraitUrl: string
  accentFrom: string
  accentTo: string
  unlockAtItems: number
  samplePrompts: string[]
  status: string
  styleDna: PersonaStyleDNA
  voice: PersonaVoice
}

// OnboardPersonaPayload is the body of POST/PUT /partner/personas.
export interface OnboardPersonaPayload {
  slug?: string
  displayName: string
  partnerName?: string
  tagline?: string
  greeting?: string
  portraitUrl?: string
  accentFrom?: string
  accentTo?: string
  unlockAtItems: number
  samplePrompts?: string[]
  styleDna: PersonaStyleDNA
  voice: PersonaVoice
}

// ── Onboarding schema (GET /partner/personas/schema) ─────────────────────────

export interface PersonaSchemaField {
  key: string
  label: string
  type: 'text' | 'textarea' | 'number' | 'list' | 'select'
  placeholder?: string
  help?: string
  options?: string[]
  required?: boolean
}

export interface PersonaSchemaStep {
  key: string
  title: string
  description: string
  fields: PersonaSchemaField[]
}

export interface PersonaSchema {
  steps: PersonaSchemaStep[]
  styleOptions: string[]
  genderOptions: string[]
}
