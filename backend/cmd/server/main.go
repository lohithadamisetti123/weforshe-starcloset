// Command server boots the Myntra Bharat AI backend.
//
// This file's only job is dependency injection: build each layer from the one
// below it — config, then infrastructure, then repositories, then the engine
// and AI extractors, then services, then handlers, then routes — and start
// listening. Nothing here decides behaviour; everything here decides wiring.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/database"
	"github.com/myntra/bharat-ai-backend/database/seed"
	"github.com/myntra/bharat-ai-backend/docs"
	"github.com/myntra/bharat-ai-backend/internal/ai"
	"github.com/myntra/bharat-ai-backend/internal/config"
	"github.com/myntra/bharat-ai-backend/internal/handlers"
	"github.com/myntra/bharat-ai-backend/internal/middleware"
	"github.com/myntra/bharat-ai-backend/internal/personalization"
	"github.com/myntra/bharat-ai-backend/internal/personas"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
	"github.com/myntra/bharat-ai-backend/internal/routes"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
	"github.com/myntra/bharat-ai-backend/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		// The logger may not exist yet if config.Load itself failed, so this
		// path goes to stderr directly rather than through zap.
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log, err := logger.New(cfg.Server.LogLevel, cfg.Server.IsProduction())
	if err != nil {
		return fmt.Errorf("building logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	log.Info("starting myntra bharat ai backend",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port))

	// A single boot-scoped context, distinct from any request context, so
	// startup work (migrating, seeding, warming the vocabulary) is not
	// accidentally tied to the lifetime of whichever request triggers it —
	// there is no request yet.
	bootCtx, bootCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer bootCancel()

	db, err := database.NewPostgres(bootCtx, cfg.Database, log, cfg.Server.IsProduction())
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}
	defer func() {
		if cerr := database.ClosePostgres(db); cerr != nil {
			log.Warn("closing postgres", zap.Error(cerr))
		}
	}()

	if cfg.Database.AutoMigrate {
		if err := database.Migrate(cfg.Database.URL, log); err != nil {
			return fmt.Errorf("migrating database: %w", err)
		}
	}

	// Seeding is gated on both the config flag and the environment, not just
	// the flag: DB_AUTO_SEED defaults to true so a fresh `docker compose up`
	// works with zero configuration, and that default must not survive a
	// misconfigured production deploy that forgot to turn it off. See the
	// seeder's own doc comment in database/seed/seed.go.
	if cfg.Database.AutoSeed && !cfg.Server.IsProduction() {
		if err := seed.NewSeeder(db, log).Run(bootCtx); err != nil {
			return fmt.Errorf("seeding database: %w", err)
		}
	}

	redisClient, err := database.NewRedis(bootCtx, cfg.Redis, log)
	if err != nil {
		// A malformed REDIS_URL is a config bug and should stop the boot; an
		// unreachable Redis is not — NewRedis already distinguishes the two,
		// returning (nil, nil) for the latter so caching degrades instead of
		// failing the whole service.
		return fmt.Errorf("configuring redis: %w", err)
	}
	defer func() {
		if cerr := database.CloseRedis(redisClient); cerr != nil {
			log.Warn("closing redis", zap.Error(cerr))
		}
	}()
	cache := repositories.NewCache(redisClient, cfg.Redis.CacheTTL, log)

	// --- Repositories -------------------------------------------------------
	productRepo := repositories.NewProductRepository(db, log)
	categoryRepo := repositories.NewCategoryRepository(db, log)
	userRepo := repositories.NewUserRepository(db, log)
	historyRepo := repositories.NewHistoryRepository(db, log)
	recommendationRepo := repositories.NewRecommendationRepository(db, log)
	favoriteRepo := repositories.NewFavoriteRepository(db, log)
	cartRepo := repositories.NewCartRepository(db, log)
	savedCollectionRepo := repositories.NewSavedCollectionRepository(db, log)
	collectionCartRepo := repositories.NewCollectionCartRepository(db, log)
	userProfileRepo := repositories.NewUserProfileRepository(db, log)
	pastOrderRepo := repositories.NewPastOrderRepository(db, log)
	starProgressRepo := repositories.NewStarProgressRepository(db, log)

	// --- Recommendation engine ------------------------------------------------
	// productRepo satisfies recommendation.CandidateSource directly — no
	// adapter needed, see product_repo.go.
	engine := recommendation.NewEngine(productRepo, cfg.Recommendation)

	// --- AI extractors --------------------------------------------------------
	vocab, err := buildVocabulary(bootCtx, productRepo)
	if err != nil {
		return fmt.Errorf("building extraction vocabulary: %w", err)
	}
	ruleExtractor := ai.NewRuleExtractor(vocab)

	// gemini stays a genuinely nil ai.Extractor, not a non-nil interface
	// wrapping a nil *GeminiExtractor — assigning the concrete type straight
	// into the interface on a failed construction is the classic Go trap that
	// would make AssistantService's `if s.gemini == nil` check always false.
	var (
		geminiExtractor ai.Extractor
		geminiConcrete  *ai.GeminiExtractor
	)
	if ge, gerr := ai.NewGeminiExtractor(bootCtx, cfg.Gemini, log); gerr != nil {
		if !errors.Is(gerr, ai.ErrGeminiUnavailable) {
			log.Warn("gemini extractor unavailable; falling back to rules only", zap.Error(gerr))
		} else {
			log.Info("gemini not configured; running on the rule-based extractor only")
		}
	} else {
		geminiExtractor = ge
		geminiConcrete = ge
		log.Info("gemini extractor ready", zap.String("model", cfg.Gemini.Model))
	}
	if geminiConcrete != nil {
		defer func() {
			if cerr := geminiConcrete.Close(); cerr != nil {
				log.Warn("closing gemini client", zap.Error(cerr))
			}
		}()
	}

	// --- StarCloset personas --------------------------------------------------
	// The stylist roster is file-backed (data/personas.json), seeded from the
	// embedded default on first boot. It is intentionally not in Postgres: a
	// persona is content, edited through the partner portal, not transactional
	// data — see internal/personas.
	personaStore, err := personas.DefaultStore()
	if err != nil {
		return fmt.Errorf("loading personas: %w", err)
	}

	tokens := utils.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	// --- Services -------------------------------------------------------------
	authService := services.NewAuthService(userRepo, tokens, log)
	productService := services.NewProductService(productRepo, categoryRepo, engine, cache, log)

	// The personalization builder is the assistant's optional personalizer: it
	// turns a shopper's stored profile + order history into the context the
	// engine's additive boost layer reads. Passing it here is the only wiring
	// the recommendation pipeline needs — swapping its data source for real
	// Myntra account data later happens entirely inside the builder.
	personalizer := personalization.NewBuilder(userProfileRepo, pastOrderRepo, log)

	assistantService := services.NewAssistantService(
		ruleExtractor, geminiExtractor, engine, recommendationRepo, historyRepo, cache, personalizer, personaStore, log)
	personaService := services.NewPersonaService(personaStore, log)
	starClosetService := services.NewStarClosetService(personaStore, starProgressRepo, log)
	profileService := services.NewProfileService(userProfileRepo, pastOrderRepo, productRepo, log)
	historyService := services.NewHistoryService(historyRepo, cache, log)
	recommendationService := services.NewRecommendationService(recommendationRepo, log)
	favoriteService := services.NewFavoriteService(favoriteRepo, log)
	cartService := services.NewCartService(cartRepo, savedCollectionRepo, collectionCartRepo, log)
	dashboardService := services.NewDashboardService(historyRepo, recommendationRepo, productRepo, cache, log)

	val := validator.New()

	h := routes.Handlers{
		Health:         handlers.NewHealthHandler(db, cache),
		Auth:           handlers.NewAuthHandler(authService, val),
		Products:       handlers.NewProductHandler(productService, val),
		Assistant:      handlers.NewAssistantHandler(assistantService, val),
		History:        handlers.NewHistoryHandler(historyService, val),
		Recommendation: handlers.NewRecommendationHandler(recommendationService, val),
		Favorite:       handlers.NewFavoriteHandler(favoriteService, val),
		Dashboard:      handlers.NewDashboardHandler(dashboardService),
		Profile:        handlers.NewProfileHandler(profileService, val),
		Cart:           handlers.NewCartHandler(cartService, val),
		Persona:        handlers.NewPersonaHandler(personaService, val),
		StarCloset:     handlers.NewStarClosetHandler(starClosetService, val),
	}

	router := buildRouter(cfg, log)
	docs.Register(router)
	routes.Register(router, h, authService, tokens, cfg, log)

	return serve(router, cfg, log)
}

// buildVocabulary loads the catalogue's distinct brands and colors so the
// rule-based extractor can match a shopper's mention of either against what
// the store actually carries. It fails open: an empty catalogue at boot
// (before the first seed) yields an empty vocabulary rather than an error, and
// the extractor still works on occasion, style, gender, category and budget.
func buildVocabulary(ctx context.Context, products repositories.ProductRepository) (ai.Vocabulary, error) {
	brands, err := products.Brands(ctx)
	if err != nil {
		return ai.Vocabulary{}, fmt.Errorf("loading brands: %w", err)
	}
	colors, err := products.Colors(ctx)
	if err != nil {
		return ai.Vocabulary{}, fmt.Errorf("loading colors: %w", err)
	}
	return ai.Vocabulary{Brands: brands, Colors: colors}, nil
}

// buildRouter assembles the global gin middleware chain.
//
// Order is load-bearing: RequestID must run before Logger and Recovery so both
// can attach the correlation ID; Recovery must wrap everything below it to
// catch a panic anywhere in the chain; ErrorHandler must run after the routes
// so it sees whatever the handler attached to the context.
func buildRouter(cfg *config.Config, log *zap.Logger) *gin.Engine {
	if cfg.Server.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.Recovery(log),
		middleware.Logger(log),
		middleware.CORS(cfg.CORS),
		middleware.ErrorHandler(log),
	)
	return router
}

// serve starts the HTTP server and blocks until a shutdown signal arrives,
// then drains in-flight requests before returning.
func serve(router *gin.Engine, cfg *config.Config, log *zap.Logger) error {
	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       cfg.Server.RequestTimeout,
		WriteTimeout:      cfg.Server.RequestTimeout + 5*time.Second,
		IdleTimeout:       90 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("server: %w", err)
		}
		return nil

	case <-ctx.Done():
		log.Info("shutdown signal received, draining in-flight requests")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		log.Info("shutdown complete")
		return nil
	}
}
