package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"go.uber.org/zap"
	"google.golang.org/api/option"

	"github.com/myntra/bharat-ai-backend/internal/config"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
)

// Extractor is the interface both the rule-based and LLM extractors satisfy.
// The assistant service depends on this, never on a concrete type, which is why
// the service's tests need neither an API key nor a network.
type Extractor interface {
	// ExtractIntent reads structured intent from a prompt. Implementations must
	// treat ctx cancellation as authoritative and must never block indefinitely.
	ExtractIntent(ctx context.Context, prompt string) (*recommendation.Intent, error)
	// Name identifies the extractor in logs and in the persisted Source field.
	Name() string
}

// ErrGeminiUnavailable signals that the LLM path could not be used and the
// caller should proceed on rules alone. It is a normal condition, not a fault:
// no API key in a hackathon demo, a cold-start quota error, an outage.
var ErrGeminiUnavailable = errors.New("ai: gemini unavailable")

// geminiResponse is the exact JSON shape Gemini is instructed to return.
//
// Every field is a pointer or a slice so that "omitted" and "empty" stay
// distinguishable. This matters more than it looks: if budget were a plain int,
// a model that declined to guess a budget would return 0, and 0 is
// indistinguishable from "free". Pointers make the model's silence legible.
type geminiResponse struct {
	Budget struct {
		Min *int `json:"min"`
		Max *int `json:"max"`
	} `json:"budget"`
	Occasion      *string  `json:"occasion"`
	Style         *string  `json:"style"`
	Gender        *string  `json:"gender"`
	Category      *string  `json:"category"`
	Colors        []string `json:"colors"`
	Brands        []string `json:"brands"`
	ExcludeColors []string `json:"excludeColors"`
	ExcludeStyles []string `json:"excludeStyles"`
	ExcludeBrands []string `json:"excludeBrands"`
	Sort          *string  `json:"sort"`
	Confidence    *float64 `json:"confidence"`
	Reasoning     *string  `json:"reasoning"`
}

// GeminiExtractor calls Google's Generative AI API for intent extraction.
type GeminiExtractor struct {
	client *genai.Client
	model  *genai.GenerativeModel
	cfg    config.GeminiConfig
	log    *zap.Logger
}

// NewGeminiExtractor constructs a Gemini-backed extractor.
//
// It returns (nil, ErrGeminiUnavailable) rather than an error when Gemini is
// simply not configured. Booting the whole service because an optional
// enhancement is switched off would be the wrong failure mode — the assistant
// degrades to rules and keeps working.
func NewGeminiExtractor(ctx context.Context, cfg config.GeminiConfig, log *zap.Logger) (*GeminiExtractor, error) {
	if !cfg.Usable() {
		return nil, ErrGeminiUnavailable
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("ai: creating gemini client: %w", err)
	}

	model := client.GenerativeModel(cfg.Model)

	// Force JSON out of the model rather than parsing prose. Without this the
	// model wraps its answer in ```json fences and commentary roughly half the
	// time, and the parser becomes a scraper.
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = intentSchema()

	// Near-zero temperature: this is an extraction task with a right answer,
	// not a creative one. Sampling variety here would mean the same prompt
	// yields different shortlists on refresh.
	model.SetTemperature(0.1)
	model.SetTopP(0.8)
	model.SetMaxOutputTokens(512)

	// The safety filters exist for open-ended chat. A blocked response here
	// means a silent failure on a benign fashion prompt, so they are relaxed to
	// the lowest setting the API allows for this narrow extraction task.
	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockOnlyHigh},
	}

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}

	return &GeminiExtractor{client: client, model: model, cfg: cfg, log: log}, nil
}

// Name implements Extractor.
func (g *GeminiExtractor) Name() string { return "gemini" }

// Close releases the underlying client.
func (g *GeminiExtractor) Close() error {
	if g == nil || g.client == nil {
		return nil
	}
	return g.client.Close()
}

// ExtractIntent asks Gemini to read the prompt into structured intent.
func (g *GeminiExtractor) ExtractIntent(ctx context.Context, prompt string) (*recommendation.Intent, error) {
	if g == nil || g.model == nil {
		return nil, ErrGeminiUnavailable
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, errors.New("ai: empty prompt")
	}

	// A hard deadline, independent of the request's own. The assistant endpoint
	// has a latency budget and a slow LLM must not consume all of it: the rules
	// have already produced a usable Intent, so waiting is a choice with a
	// perfectly good alternative.
	ctx, cancel := context.WithTimeout(ctx, g.cfg.Timeout)
	defer cancel()

	start := time.Now()
	resp, err := g.model.GenerateContent(ctx, genai.Text(buildUserPrompt(prompt)))
	elapsed := time.Since(start)

	if err != nil {
		g.log.Warn("gemini extraction failed; falling back to rules",
			zap.Error(err), zap.Duration("elapsed", elapsed))
		return nil, fmt.Errorf("%w: %v", ErrGeminiUnavailable, err)
	}

	raw, err := firstTextPart(resp)
	if err != nil {
		g.log.Warn("gemini returned no usable content", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrGeminiUnavailable, err)
	}

	intent, err := parseGeminiIntent(raw, prompt)
	if err != nil {
		g.log.Warn("gemini returned unparseable json",
			zap.Error(err), zap.String("raw", truncateForLog(raw)))
		return nil, fmt.Errorf("%w: %v", ErrGeminiUnavailable, err)
	}

	g.log.Debug("gemini extraction ok",
		zap.Duration("elapsed", elapsed),
		zap.String("occasion", intent.OccasionOrEmpty()),
		zap.Float64("confidence", intent.Confidence))

	return intent, nil
}

// firstTextPart pulls the text out of a Gemini response, distinguishing the
// several distinct ways it can be empty.
func firstTextPart(resp *genai.GenerateContentResponse) (string, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		if resp != nil && resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != genai.BlockReasonUnspecified {
			return "", fmt.Errorf("prompt blocked: %v", resp.PromptFeedback.BlockReason)
		}
		return "", errors.New("no candidates returned")
	}

	cand := resp.Candidates[0]
	if cand.Content == nil || len(cand.Content.Parts) == 0 {
		return "", fmt.Errorf("empty candidate content (finish reason: %v)", cand.FinishReason)
	}

	var b strings.Builder
	for _, part := range cand.Content.Parts {
		if t, ok := part.(genai.Text); ok {
			b.WriteString(string(t))
		}
	}
	if b.Len() == 0 {
		return "", errors.New("candidate contained no text parts")
	}
	return b.String(), nil
}

// parseGeminiIntent converts the model's JSON into a validated Intent.
//
// Nothing the model says is trusted directly. Every enum is re-parsed against
// the catalogue vocabulary and silently dropped if it does not resolve, because
// LLMs invent plausible-looking members ("Formalwear", "Semi-Formal",
// "Traditional") that would sail through a naive cast and then match no product
// at all. The vocabulary is the authority; the model is a suggestion.
func parseGeminiIntent(raw, originalPrompt string) (*recommendation.Intent, error) {
	cleaned := stripCodeFences(raw)

	var gr geminiResponse
	if err := json.Unmarshal([]byte(cleaned), &gr); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	intent := recommendation.NewIntent(originalPrompt)
	intent.Source = recommendation.SourceGemini

	if gr.Budget.Min != nil || gr.Budget.Max != nil {
		br := &recommendation.BudgetRange{}
		if gr.Budget.Min != nil && *gr.Budget.Min > 0 {
			v := *gr.Budget.Min
			br.Min = &v
		}
		if gr.Budget.Max != nil && *gr.Budget.Max > 0 {
			v := *gr.Budget.Max
			br.Max = &v
		}
		if !br.IsZero() {
			if br.HasMin() && br.HasMax() && *br.Min > *br.Max {
				br.Min, br.Max = br.Max, br.Min
			}
			intent.Budget = br
		}
	}

	if gr.Occasion != nil {
		if o, ok := models.ParseOccasion(*gr.Occasion); ok {
			intent.Occasion = &o
		}
	}
	if gr.Style != nil {
		if s, ok := models.ParseStyle(*gr.Style); ok {
			intent.Style = &s
		}
	}
	if gr.Gender != nil {
		if g, ok := models.ParseGender(*gr.Gender); ok {
			intent.Gender = &g
		}
	}
	if gr.Category != nil {
		if c, ok := models.ParseCategory(*gr.Category); ok {
			intent.Category = &c
		}
	}

	intent.Colors = cleanStrings(gr.Colors)
	intent.Brands = cleanStrings(gr.Brands)
	intent.ExcludeColors = cleanStrings(gr.ExcludeColors)
	intent.ExcludeBrands = cleanStrings(gr.ExcludeBrands)

	for _, s := range cleanStrings(gr.ExcludeStyles) {
		if st, ok := models.ParseStyle(s); ok {
			intent.ExcludeStyles = append(intent.ExcludeStyles, st)
		}
	}

	intent.Sort = recommendation.SortRelevance
	if gr.Sort != nil {
		if s, ok := parseSortHint(*gr.Sort); ok {
			intent.Sort = s
		}
	}

	// A model asked to score its own confidence will happily answer 0.95 to
	// everything, so the value is clamped into a band that cannot let the LLM
	// overrule the merge logic on the strength of its own self-regard.
	intent.Confidence = 0.7
	if gr.Confidence != nil {
		c := *gr.Confidence
		if c < 0.25 {
			c = 0.25
		}
		if c > 0.95 {
			c = 0.95
		}
		intent.Confidence = c
	}

	if gr.Reasoning != nil && *gr.Reasoning != "" {
		intent.Notes = append(intent.Notes, "gemini: "+truncateForLog(*gr.Reasoning))
	}

	intent.Keywords = recommendation.ExtractKeywords(originalPrompt)
	return intent, nil
}

// stripCodeFences removes markdown fencing. ResponseMIMEType should make this
// unnecessary, but models regress and a one-line guard is cheaper than an
// incident.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	if i := strings.LastIndex(s, "```"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func parseSortHint(s string) (recommendation.SortHint, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "price_asc", "price asc", "cheapest":
		return recommendation.SortPriceAsc, true
	case "price_desc", "price desc", "expensive":
		return recommendation.SortPriceDesc, true
	case "rating", "top_rated", "best_rated":
		return recommendation.SortRating, true
	case "discount", "deals":
		return recommendation.SortDiscount, true
	case "newest", "latest":
		return recommendation.SortNewest, true
	case "relevance", "":
		return recommendation.SortRelevance, true
	default:
		return recommendation.SortRelevance, false
	}
}

func cleanStrings(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, s := range in {
		t := strings.TrimSpace(s)
		if t == "" || len(t) > 40 {
			continue
		}
		key := strings.ToLower(t)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func truncateForLog(s string) string {
	const max = 240
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
