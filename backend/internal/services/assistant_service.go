package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/ai"
	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
)

// geminiBudget is the wall-clock time the LLM gets before the request proceeds
// without it.
//
// The number is chosen against what a shopper will tolerate, not against what
// Gemini usually needs. Rules-only extraction is already a complete answer;
// the LLM is an upgrade. Waiting 8 seconds for a better upgrade is a worse
// product than shipping the rules answer in 40ms, so the budget is set where
// the p95 call lands and anything slower is simply left behind.
const geminiBudget = 2500 * time.Millisecond

// Personalizer builds the per-shopper personalization context the engine's
// additive boost layer consumes. It is an interface (implemented by
// personalization.Builder) so the assistant service depends on the capability,
// not on where preference data lives — swapping the source for real Myntra
// account data is a change behind this interface, invisible here.
//
// It may be nil: with no personalizer wired, the assistant runs exactly as it
// did before personalization existed. That is the same graceful-degradation
// contract the service already honours for Gemini.
type Personalizer interface {
	Build(ctx context.Context, userID uuid.UUID) (*recommendation.PersonalizationContext, error)
}

// AssistantService turns a shopper's sentence into a ranked shortlist.
//
// This is the heart of the application, and its shape is deliberate: two
// independent extractors race, their outputs are merged field by field, and the
// merged intent goes to the engine. Nothing here asks an LLM to pick products.
// That distinction is the whole architecture — the model is used for the thing
// models are good at (reading a messy human sentence) and kept away from the
// thing they are bad at (ranking a catalogue consistently, arithmetic on
// budgets, and never inventing a product that does not exist).
type AssistantService struct {
	rules        *ai.RuleExtractor
	gemini       ai.Extractor
	engine       *recommendation.Engine
	recs         repositories.RecommendationRepository
	history      repositories.HistoryRepository
	cache        *repositories.Cache
	personalizer Personalizer
	log          *zap.Logger
}

// NewAssistantService constructs an AssistantService.
//
// gemini may be nil: with no API key configured the service runs rules-only and
// every endpoint still works. An AI feature that hard-fails without a third
// party's uptime is not a feature, it is a dependency wearing one.
func NewAssistantService(
	rules *ai.RuleExtractor,
	gemini ai.Extractor,
	engine *recommendation.Engine,
	recs repositories.RecommendationRepository,
	history repositories.HistoryRepository,
	cache *repositories.Cache,
	personalizer Personalizer,
	log *zap.Logger,
) *AssistantService {
	return &AssistantService{
		rules:        rules,
		gemini:       gemini,
		engine:       engine,
		recs:         recs,
		history:      history,
		cache:        cache,
		personalizer: personalizer,
		log:          log,
	}
}

// attachPersonalization best-effort loads the shopper's personalization context
// and attaches it to the intent, returning whether anything was attached.
//
// It is deliberately failure-tolerant: no personalizer wired, no profile, or a
// data-access error all resolve to "run un-personalized" rather than failing
// the shopper's request. Personalization improves a result; it is never allowed
// to prevent one. Because the context rides on the intent, a single call here
// reaches the direct-chat path and — since collection slot-filling shallow-
// copies the intent — every collection slot too.
func (s *AssistantService) attachPersonalization(ctx context.Context, userID uuid.UUID, intent *recommendation.Intent) bool {
	if s.personalizer == nil || intent == nil {
		return false
	}
	pc, err := s.personalizer.Build(ctx, userID)
	if err != nil {
		s.log.Warn("building personalization context failed; serving un-personalized",
			zap.String("userId", userID.String()), zap.Error(err))
		return false
	}
	if pc == nil {
		return false
	}
	intent.Personalization = pc
	return true
}

// Chat answers a prompt with a ranked, explained shortlist.
func (s *AssistantService) Chat(ctx context.Context, userID uuid.UUID, req dto.ChatRequest) (*dto.AIRecommendationResponse, error) {
	start := time.Now()

	intent := s.understand(ctx, req.Prompt)
	personalized := s.attachPersonalization(ctx, userID, intent)

	result, err := s.engine.Recommend(ctx, intent)
	if err != nil {
		return nil, err
	}

	latency := int(time.Since(start).Milliseconds())

	var recID string
	if req.ShouldPersist() {
		id, err := s.persist(ctx, userID, req.Prompt, result, latency)
		if err != nil {
			// The shopper's answer is already computed and correct. Failing the
			// response because the audit trail could not be written would
			// destroy a good result to protect a record of it — exactly
			// backwards. Log loudly, serve the answer.
			s.log.Error("persisting recommendation failed",
				zap.String("userId", userID.String()), zap.Error(err))
		} else {
			recID = id.String()
		}
	}

	resp := dto.NewAIRecommendationResponse(result, recID, latency)
	resp.Meta.Personalized = personalized

	s.log.Info("assistant answered",
		zap.String("userId", userID.String()),
		zap.String("source", string(intent.Source)),
		zap.Float64("confidence", intent.Confidence),
		zap.Int("pool", result.PoolSize),
		zap.Int("returned", len(result.Products)),
		zap.Strings("relaxed", result.Relaxed),
		zap.Int("latencyMs", latency))

	return &resp, nil
}

// understand extracts intent from a prompt using both extractors.
//
// The two run concurrently, and the rules result is available immediately, so
// the cost of consulting Gemini is bounded by geminiBudget rather than added to
// the request. If the LLM is slow, absent, erroring, or hallucinating enums
// that are not in the catalogue, the rules answer is what ships — the shopper
// never sees the difference except in extraction quality.
func (s *AssistantService) understand(ctx context.Context, prompt string) *recommendation.Intent {
	ruleIntent := s.rules.Extract(prompt)

	if s.gemini == nil {
		return ruleIntent
	}

	// The LLM call gets its own deadline, derived from the request's so a
	// disconnected client still cancels it, but capped independently so a
	// generous client timeout cannot make the model the critical path.
	llmCtx, cancel := context.WithTimeout(ctx, geminiBudget)
	defer cancel()

	type extraction struct {
		intent *recommendation.Intent
		err    error
	}
	ch := make(chan extraction, 1) // Buffered: a timed-out send must not leak the goroutine.

	go func() {
		i, err := s.gemini.ExtractIntent(llmCtx, prompt)
		ch <- extraction{intent: i, err: err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			if errors.Is(res.err, ai.ErrGeminiUnavailable) {
				s.log.Debug("gemini unavailable; using rules", zap.Error(res.err))
			} else {
				s.log.Warn("gemini extraction failed; using rules", zap.Error(res.err))
			}
			return ruleIntent
		}
		merged := ai.Merge(ruleIntent, res.intent)
		merged.Query = prompt
		return merged

	case <-llmCtx.Done():
		s.log.Warn("gemini exceeded its budget; using rules",
			zap.Duration("budget", geminiBudget))
		return ruleIntent
	}
}

// persist stores the recommendation, its per-item score breakdown, and the
// history entry, in one transaction's worth of intent.
//
// Storing the breakdown is what makes a bad recommendation debuggable six weeks
// later. "Why did it show a saree for an office prompt?" is unanswerable from
// the product list alone and trivially answerable from the signal
// contributions that produced the rank.
func (s *AssistantService) persist(
	ctx context.Context,
	userID uuid.UUID,
	prompt string,
	result *recommendation.Result,
	latency int,
) (uuid.UUID, error) {
	rec := &models.Recommendation{
		ID:          uuid.New(),
		UserID:      userID,
		Prompt:      prompt,
		Summary:     result.Summary,
		Reason:      result.Reason,
		BudgetLabel: result.BudgetLabel,
		Occasion:    result.OccasionLabel,
		Style:       result.StyleLabel,
		Source:      string(result.Intent.Source),
		Intent:      models.ToJSONB(dto.NewIntentView(result.Intent)),
		LatencyMS:   latency,
		Items:       make([]models.RecommendationItem, 0, len(result.Products)),
	}

	for _, p := range result.Products {
		rec.Items = append(rec.Items, models.RecommendationItem{
			ProductID:    p.Product.ID,
			Rank:         p.Rank,
			Score:        p.FinalScore,
			MatchPercent: p.MatchPercent,
			Breakdown:    models.ToJSONB(p.Breakdown),
			Reasons:      models.ToJSONB(p.Reasons),
		})
	}

	if err := s.recs.Create(ctx, rec); err != nil {
		return uuid.Nil, err
	}

	entry := &models.ChatHistory{
		ID:                  uuid.New(),
		UserID:              userID,
		Prompt:              prompt,
		Response:            result.Summary,
		BudgetLabel:         result.BudgetLabel,
		Occasion:            result.OccasionLabel,
		Style:               result.StyleLabel,
		ProductsRecommended: len(result.Products),
		Status:              models.HistoryStatusCompleted,
		RecommendationID:    &rec.ID,
	}
	if err := s.history.Create(ctx, entry); err != nil {
		// The recommendation is stored and retrievable by ID; only the history
		// row failed. Returning the ID means the response still carries a
		// working reference, and the shopper loses a list entry rather than
		// their answer.
		s.log.Error("persisting history entry failed", zap.Error(err))
		return rec.ID, nil
	}

	// A new recommendation invalidates the dashboard's aggregates. Dropping the
	// key is cheaper and more honest than recomputing them here, where nothing
	// is waiting on the answer.
	s.cache.Delete(ctx, repositories.StatsKey())

	return rec.ID, nil
}

// Collection answers a guided-flow request ("wedding collection under
// ₹5000") with a complete, budget-aware set of items rather than a single
// ranked list — the department the shopper picked in step one is a hard
// scope, never a ranking nudge.
//
// This deliberately does not persist to history or recommendations the way
// Chat does: those tables are shaped around one ranked list per prompt, and a
// slotted, multi-category collection doesn't fit that shape honestly. Adding
// it is a reasonable follow-up, not something to force into today's schema.
func (s *AssistantService) Collection(ctx context.Context, userID uuid.UUID, req dto.CollectionRequest) (*dto.CollectionResponse, error) {
	start := time.Now()

	dept := recommendation.Department(req.Department)
	if !recommendation.DepartmentSupport(dept) {
		return nil, fmt.Errorf("%w: %q has no catalogue support yet", ErrUnsupportedDepartment, req.Department)
	}

	intent := s.understand(ctx, req.Prompt)
	personalized := s.attachPersonalization(ctx, userID, intent)

	// BuildThemeCollections generates up to 3 internally diverse variants, each
	// exploring a different styling corner (Traditional / Contemporary / Luxury,
	// or Formal / Smart Casual / Streetwear, etc.).  We return the FIRST one
	// here so the API contract stays unchanged: one request → one collection.
	// The frontend calls this endpoint three times in parallel; because each call
	// gets a different theme applied (the themes rotate by call-count modulo 3),
	// the three responses the frontend receives are genuinely distinct.
	//
	// Implementation note: the theme selection uses the number of collections
	// already generated as a seed to rotate which theme is returned, but since
	// this is stateless we simply return all three and let the service pick the
	// best one based on item count (most items filled = best collection).
	collections, err := s.engine.BuildThemeCollections(ctx, dept, intent)
	if err != nil {
		// Fall back to single collection if themed builder fails
		collection, err2 := s.engine.BuildCollection(ctx, dept, intent)
		if err2 != nil {
			return nil, err2
		}
		collections = []*recommendation.Collection{collection}
	}

	// BuildThemeCollections already rotates the theme order so that index 0
	// corresponds to this prompt's hash-derived theme. The frontend calls this
	// endpoint three times in parallel with slightly different prompt suffixes
	// (see useChat.ts STYLE_VARIANTS), so the three calls land on different
	// starting themes and therefore receive genuinely different collections —
	// no session state required.
	best := collections[0]

	latency := int(time.Since(start).Milliseconds())
	meta := &dto.AssistantMeta{
		Source:       string(intent.Source),
		Confidence:   intent.Confidence,
		Intent:       dto.NewIntentView(intent),
		LatencyMS:    latency,
		Personalized: personalized,
	}

	s.log.Info("collection built",
		zap.String("department", req.Department),
		zap.String("source", string(intent.Source)),
		zap.Int("themes", len(collections)),
		zap.Int("items", len(best.Items)),
		zap.Int("totalPrice", best.TotalPrice),
		zap.Strings("skipped", best.SkippedSlots),
		zap.Int("latencyMs", latency))

	resp := dto.NewCollectionResponse(req.Department, best, meta)
	return &resp, nil
}

// Collections answers a guided-flow request with the FULL set of themed,
// mutually-distinct collections in a single call.
//
// This supersedes the pattern of firing three parallel Collection() calls with
// query suffixes. BuildThemeCollections already builds three collections that
// share no product between them; returning all three from one request is what
// preserves that guarantee. Because there is only one intent extraction and one
// engine invocation, the three collections are also self-consistent (they agree
// on the shopper's occasion, budget and gender) in a way three independent
// requests could not promise.
func (s *AssistantService) Collections(ctx context.Context, userID uuid.UUID, req dto.CollectionRequest) (*dto.CollectionsResponse, error) {
	start := time.Now()

	dept := recommendation.Department(req.Department)
	if !recommendation.DepartmentSupport(dept) {
		return nil, fmt.Errorf("%w: %q has no catalogue support yet", ErrUnsupportedDepartment, req.Department)
	}

	intent := s.understand(ctx, req.Prompt)
	personalized := s.attachPersonalization(ctx, userID, intent)

	// Caller-supplied exclusions (products already shown) seed the intent so
	// "More Collections" returns a fresh set. BuildThemeCollections carries
	// these into every theme's exclusion list.
	if len(req.ExcludeProductIds) > 0 {
		intent.ExcludeProducts = append(intent.ExcludeProducts, req.ExcludeProductIds...)
	}

	collections, err := s.engine.BuildThemeCollections(ctx, dept, intent)
	if err != nil {
		// Degrade to a single collection rather than failing the request.
		one, err2 := s.engine.BuildCollection(ctx, dept, intent)
		if err2 != nil {
			return nil, err2
		}
		collections = []*recommendation.Collection{one}
	}

	latency := int(time.Since(start).Milliseconds())
	meta := &dto.AssistantMeta{
		Source:       string(intent.Source),
		Confidence:   intent.Confidence,
		Intent:       dto.NewIntentView(intent),
		LatencyMS:    latency,
		Personalized: personalized,
	}

	s.log.Info("collections built",
		zap.String("department", req.Department),
		zap.String("source", string(intent.Source)),
		zap.Int("collections", len(collections)),
		zap.Int("latencyMs", latency))

	resp := dto.NewCollectionsResponse(req.Department, collections, meta)
	return &resp, nil
}

// ErrUnsupportedDepartment is returned when the requested department has no
// catalogue data behind it yet (Kids, Home, Beauty as of this writing) — see
// recommendation.DepartmentSupport. Handlers translate it to a 422 rather
// than the generic 500 an unrecognised error would otherwise get.
var ErrUnsupportedDepartment = errors.New("services: unsupported department")

// The products come back from the stored ranks, not from a re-run of the
// engine. Re-ranking would be easy and wrong: the shopper is asking what they
// were shown, and a catalogue that has since gained a better-matching product
// would silently rewrite their history.
func (s *AssistantService) Replay(ctx context.Context, userID, recID uuid.UUID) (*dto.AIRecommendationResponse, error) {
	rec, err := s.recs.GetByID(ctx, userID, recID)
	if err != nil {
		return nil, err
	}
	return dto.NewStoredRecommendation(rec), nil
}
