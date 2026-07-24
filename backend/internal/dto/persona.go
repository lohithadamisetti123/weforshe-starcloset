package dto

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/personas"
)

// This file is the wire contract for StarCloset personas. It mirrors the
// frontend's src/types/persona.ts (StylistPersona, PersonaDetail,
// OnboardPersonaPayload) exactly — the running JSON is the real contract.

// StylistPersona is the shopper-facing view of a stylist. It deliberately omits
// the Style DNA and Voice internals: a shopper needs to recognise, pick and
// unlock a stylist, not read its scoring configuration.
type StylistPersona struct {
	Slug           string   `json:"slug"`
	DisplayName    string   `json:"displayName"`
	Tagline        string   `json:"tagline"`
	Greeting       string   `json:"greeting"`
	PortraitURL    string   `json:"portraitUrl"`
	AccentFrom     string   `json:"accentFrom"`
	AccentTo       string   `json:"accentTo"`
	UnlockAtItems  int      `json:"unlockAtItems"`
	SamplePrompts  []string `json:"samplePrompts"`
	StyleSignature string   `json:"styleSignature"`
	// DefaultGender lets the studio scope the guided flow to this stylist's
	// department (Women → Women, Men → Men) without exposing the rest of the
	// Style DNA to shoppers.
	DefaultGender string `json:"defaultGender"`
	Status        string `json:"status"`
}

// PersonaStyleDNA is Layer 2 on the wire.
type PersonaStyleDNA struct {
	StyleSignature  string   `json:"styleSignature"`
	PreferredStyles []string `json:"preferredStyles"`
	PreferredBrands []string `json:"preferredBrands"`
	PreferredColors []string `json:"preferredColors"`
	StyleOverride   string   `json:"styleOverride"`
	DefaultGender   string   `json:"defaultGender"`
}

// PersonaVoice is Layer 3 on the wire.
type PersonaVoice struct {
	VoiceTone         string   `json:"voiceTone"`
	GreetingTemplates []string `json:"greetingTemplates"`
	Catchphrases      []string `json:"catchphrases"`
	StyleRules        []string `json:"styleRules"`
	Avoid             []string `json:"avoid"`
	ExampleResponses  []string `json:"exampleResponses"`
	ResponseTemplates []string `json:"responseTemplates"`
}

// PersonaDetail is the full partner-facing persona across all three layers.
type PersonaDetail struct {
	Slug          string          `json:"slug"`
	DisplayName   string          `json:"displayName"`
	PartnerName   string          `json:"partnerName"`
	Tagline       string          `json:"tagline"`
	Greeting      string          `json:"greeting"`
	PortraitURL   string          `json:"portraitUrl"`
	AccentFrom    string          `json:"accentFrom"`
	AccentTo      string          `json:"accentTo"`
	UnlockAtItems int             `json:"unlockAtItems"`
	SamplePrompts []string        `json:"samplePrompts"`
	Status        string          `json:"status"`
	StyleDNA      PersonaStyleDNA `json:"styleDna"`
	Voice         PersonaVoice    `json:"voice"`
}

// OnboardPersonaPayload is the body of POST/PUT /partner/personas. The slug is
// optional on create (derived from the display name) and ignored on update.
type OnboardPersonaPayload struct {
	Slug          string          `json:"slug,omitempty"`
	DisplayName   string          `json:"displayName" binding:"required" validate:"required,min=2,max=80"`
	PartnerName   string          `json:"partnerName,omitempty" validate:"omitempty,max=120"`
	Tagline       string          `json:"tagline,omitempty" validate:"omitempty,max=200"`
	Greeting      string          `json:"greeting,omitempty" validate:"omitempty,max=400"`
	PortraitURL   string          `json:"portraitUrl,omitempty" validate:"omitempty,max=500"`
	AccentFrom    string          `json:"accentFrom,omitempty" validate:"omitempty,max=32"`
	AccentTo      string          `json:"accentTo,omitempty" validate:"omitempty,max=32"`
	UnlockAtItems int             `json:"unlockAtItems" validate:"gte=0,lte=100"`
	SamplePrompts []string        `json:"samplePrompts,omitempty" validate:"omitempty,max=12,dive,max=200"`
	StyleDNA      PersonaStyleDNA `json:"styleDna"`
	Voice         PersonaVoice    `json:"voice"`
}

// PersonaSchemaField describes one input in the onboarding wizard.
type PersonaSchemaField struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"` // text | textarea | number | list | select
	Placeholder string   `json:"placeholder,omitempty"`
	Help        string   `json:"help,omitempty"`
	Options     []string `json:"options,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

// PersonaSchemaStep is one wizard step (Identity / Style DNA / Voice / Launch).
type PersonaSchemaStep struct {
	Key         string               `json:"key"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Fields      []PersonaSchemaField `json:"fields"`
}

// PersonaSchemaResponse is the whole onboarding form contract.
type PersonaSchemaResponse struct {
	Steps         []PersonaSchemaStep `json:"steps"`
	StyleOptions  []string            `json:"styleOptions"`
	GenderOptions []string            `json:"genderOptions"`
}

// ── Mappers ──────────────────────────────────────────────────────────────────

// NewStylistPersona projects a persona onto the shopper-facing wire type.
func NewStylistPersona(p personas.Persona) StylistPersona {
	return StylistPersona{
		Slug:           p.Slug,
		DisplayName:    p.DisplayName,
		Tagline:        p.Tagline,
		Greeting:       p.Greeting,
		PortraitURL:    p.PortraitURL,
		AccentFrom:     p.AccentFrom,
		AccentTo:       p.AccentTo,
		UnlockAtItems:  p.UnlockAtItems,
		SamplePrompts:  nonNil(p.SamplePrompts),
		StyleSignature: p.StyleSignature,
		DefaultGender:  p.DefaultGender,
		Status:         string(p.Status),
	}
}

// NewStylistPersonas maps a slice.
func NewStylistPersonas(list []personas.Persona) []StylistPersona {
	out := make([]StylistPersona, 0, len(list))
	for _, p := range list {
		out = append(out, NewStylistPersona(p))
	}
	return out
}

// NewPersonaDetail projects a persona onto the full partner-facing wire type.
func NewPersonaDetail(p personas.Persona) PersonaDetail {
	return PersonaDetail{
		Slug:          p.Slug,
		DisplayName:   p.DisplayName,
		PartnerName:   p.PartnerName,
		Tagline:       p.Tagline,
		Greeting:      p.Greeting,
		PortraitURL:   p.PortraitURL,
		AccentFrom:    p.AccentFrom,
		AccentTo:      p.AccentTo,
		UnlockAtItems: p.UnlockAtItems,
		SamplePrompts: nonNil(p.SamplePrompts),
		Status:        string(p.Status),
		StyleDNA: PersonaStyleDNA{
			StyleSignature:  p.StyleSignature,
			PreferredStyles: nonNil(p.PreferredStyles),
			PreferredBrands: nonNil(p.PreferredBrands),
			PreferredColors: nonNil(p.PreferredColors),
			StyleOverride:   p.StyleOverride,
			DefaultGender:   p.DefaultGender,
		},
		Voice: PersonaVoice{
			VoiceTone:         p.Voice.VoiceTone,
			GreetingTemplates: nonNil(p.Voice.GreetingTemplates),
			Catchphrases:      nonNil(p.Voice.Catchphrases),
			StyleRules:        nonNil(p.Voice.StyleRules),
			Avoid:             nonNil(p.Voice.Avoid),
			ExampleResponses:  nonNil(p.Voice.ExampleResponses),
			ResponseTemplates: nonNil(p.Voice.ResponseTemplates),
		},
	}
}

// NewPersonaDetails maps a slice.
func NewPersonaDetails(list []personas.Persona) []PersonaDetail {
	out := make([]PersonaDetail, 0, len(list))
	for _, p := range list {
		out = append(out, NewPersonaDetail(p))
	}
	return out
}

// ToPersona builds a domain persona from an onboarding payload. Status is not
// set here — creation defaults to draft, and publishing is a separate action.
func (r OnboardPersonaPayload) ToPersona() personas.Persona {
	return personas.Persona{
		Slug:            r.Slug,
		DisplayName:     r.DisplayName,
		PartnerName:     r.PartnerName,
		Tagline:         r.Tagline,
		Greeting:        r.Greeting,
		PortraitURL:     r.PortraitURL,
		AccentFrom:      r.AccentFrom,
		AccentTo:        r.AccentTo,
		UnlockAtItems:   r.UnlockAtItems,
		SamplePrompts:   r.SamplePrompts,
		StyleSignature:  r.StyleDNA.StyleSignature,
		PreferredStyles: r.StyleDNA.PreferredStyles,
		PreferredBrands: r.StyleDNA.PreferredBrands,
		PreferredColors: r.StyleDNA.PreferredColors,
		StyleOverride:   r.StyleDNA.StyleOverride,
		DefaultGender:   r.StyleDNA.DefaultGender,
		Voice: personas.Voice{
			VoiceTone:         r.Voice.VoiceTone,
			GreetingTemplates: r.Voice.GreetingTemplates,
			Catchphrases:      r.Voice.Catchphrases,
			StyleRules:        r.Voice.StyleRules,
			Avoid:             r.Voice.Avoid,
			ExampleResponses:  r.Voice.ExampleResponses,
			ResponseTemplates: r.Voice.ResponseTemplates,
		},
	}
}

// NewPersonaSchema returns the onboarding form contract the wizard renders.
func NewPersonaSchema() PersonaSchemaResponse {
	styles := make([]string, 0, len(models.AllStyles))
	for _, s := range models.AllStyles {
		styles = append(styles, string(s))
	}
	genders := make([]string, 0, len(models.AllGenders))
	for _, g := range models.AllGenders {
		genders = append(genders, string(g))
	}

	return PersonaSchemaResponse{
		StyleOptions:  styles,
		GenderOptions: genders,
		Steps: []PersonaSchemaStep{
			{
				Key:         "identity",
				Title:       "Celebrity Identity",
				Description: "Who is this stylist, and how do they present to shoppers?",
				Fields: []PersonaSchemaField{
					{Key: "displayName", Label: "Display name", Type: "text", Placeholder: "e.g. Alia Bhatt", Required: true},
					{Key: "partnerName", Label: "Partner / brand", Type: "text", Placeholder: "e.g. StarCloset Originals"},
					{Key: "tagline", Label: "Tagline", Type: "text", Placeholder: "One line that sells the vibe"},
					{Key: "greeting", Label: "Greeting", Type: "textarea", Placeholder: "The first thing the stylist says on open"},
					{Key: "portraitUrl", Label: "Portrait URL", Type: "text", Placeholder: "/images/stylists/slug.jpg"},
					{Key: "accentFrom", Label: "Accent (from)", Type: "text", Placeholder: "#FF7EB3"},
					{Key: "accentTo", Label: "Accent (to)", Type: "text", Placeholder: "#FF9A8B"},
					{Key: "unlockAtItems", Label: "Unlock at cart items", Type: "number", Help: "0 = available from the start"},
					{Key: "samplePrompts", Label: "Sample prompts", Type: "list", Help: "Prompt chips shown in the studio"},
				},
			},
			{
				Key:         "styleDna",
				Title:       "Style DNA",
				Description: "The signature the rule-based engine reads. No AI picks products — this only biases ranking.",
				Fields: []PersonaSchemaField{
					{Key: "styleSignature", Label: "Style signature", Type: "textarea", Placeholder: "Describe the aesthetic in a sentence or two"},
					{Key: "preferredStyles", Label: "Preferred styles", Type: "list", Help: "Matched against the catalogue's style vocabulary"},
					{Key: "preferredBrands", Label: "Preferred brands", Type: "list"},
					{Key: "preferredColors", Label: "Preferred colours", Type: "list"},
					{Key: "styleOverride", Label: "Style override", Type: "select", Help: "Used only when the shopper names no style"},
					{Key: "defaultGender", Label: "Default gender", Type: "select"},
				},
			},
			{
				Key:         "voice",
				Title:       "Persona Voice",
				Description: "Deterministic template material. Replies are assembled from these — never from an external model.",
				Fields: []PersonaSchemaField{
					{Key: "voiceTone", Label: "Voice tone", Type: "text", Placeholder: "e.g. warm, bubbly, playful"},
					{Key: "greetingTemplates", Label: "Greeting templates", Type: "list", Help: "One opens each reply, chosen deterministically"},
					{Key: "responseTemplates", Label: "Response templates", Type: "list", Help: "Use {occasion} {count} {budget} {highlight} {pieces}"},
					{Key: "catchphrases", Label: "Catchphrases", Type: "list"},
					{Key: "styleRules", Label: "Style rules", Type: "list"},
					{Key: "avoid", Label: "Avoid", Type: "list"},
					{Key: "exampleResponses", Label: "Example responses", Type: "list", Help: "Shown in the launch preview"},
				},
			},
			{
				Key:         "launch",
				Title:       "Launch",
				Description: "Preview the greeting and a sample conversation, set the unlock threshold, and publish.",
				Fields:      []PersonaSchemaField{},
			},
		},
	}
}

// nonNil returns a non-nil slice so the JSON encodes [] rather than null.
func nonNil(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
