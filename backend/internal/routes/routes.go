// Package routes wires every handler onto the router, under the middleware
// chain each group needs. This file is the only place that should know the
// full URL surface of the API — handlers know their own path only through the
// swagger annotations above each method, and main.go knows none of it at all.
package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/config"
	"github.com/myntra/bharat-ai-backend/internal/handlers"
	"github.com/myntra/bharat-ai-backend/internal/middleware"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// Handlers bundles every HTTP handler the router needs. Passed as one struct
// rather than seven parameters so that adding an eighth handler someday is a
// one-line change here instead of a signature change at every call site.
type Handlers struct {
	Health         *handlers.HealthHandler
	Auth           *handlers.AuthHandler
	Products       *handlers.ProductHandler
	Assistant      *handlers.AssistantHandler
	History        *handlers.HistoryHandler
	Recommendation *handlers.RecommendationHandler
	Favorite       *handlers.FavoriteHandler
	Dashboard      *handlers.DashboardHandler
	Profile        *handlers.ProfileHandler
	Cart           *handlers.CartHandler
	Persona        *handlers.PersonaHandler
	StarCloset     *handlers.StarClosetHandler
}

// Register mounts every route.
func Register(
	router *gin.Engine,
	h Handlers,
	authService *services.AuthService,
	tokens *utils.TokenManager,
	cfg *config.Config,
	log *zap.Logger,
) {
	// --- Health -----------------------------------------------------------
	// Unauthenticated and unrate-limited: an orchestrator's probe must never
	// be the thing that fails because the API is protecting itself from load.
	router.GET("/health", h.Health.Health)
	router.GET("/health/live", h.Health.Live)
	router.GET("/health/ready", h.Health.Ready)

	// --- Global middleware --------------------------------------------------
	// A coarse IP-keyed limit applies to every route below, including the
	// unauthenticated auth endpoints — nothing here should be usable to mount
	// a credential-stuffing run at unlimited speed.
	router.Use(middleware.RateLimit(cfg.RateLimit.RPS, cfg.RateLimit.Burst))

	api := router.Group("/api/v1")

	// --- Auth ---------------------------------------------------------------
	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
	}

	// --- Public catalogue browsing -------------------------------------------
	// Products are readable without an identity: a shopper should be able to
	// browse before the app has decided whether they are a guest or a member.
	products := api.Group("/products")
	{
		products.GET("", h.Products.List)
		products.GET("/search", h.Products.Search)
		products.GET("/category/:name", h.Products.ByCategory)
		products.GET("/brand/:name", h.Products.ByBrand)
		products.GET("/:id", h.Products.GetByID)
	}
	api.GET("/categories", h.Products.Categories)
	api.GET("/brands", h.Products.Brands)
	api.GET("/assistant/departments", h.Assistant.Departments)

	// --- StarCloset stylist personas -----------------------------------------
	// The shopper roster is public: a visitor should see and pick a stylist
	// before deciding to sign in, exactly like browsing the catalogue.
	api.GET("/assistant/personas", h.Persona.ListPublished)

	// The V2 department rosters (Women / Men / Kids and their stylists) are
	// public config, like the catalogue departments — a visitor sees the whole
	// journey before signing in. Per-department unlock *progress*, being
	// per-shopper, lives under the authed group below.
	api.GET("/starcloset/departments", h.StarCloset.Departments)

	// The partner onboarding portal. Every route here is gated by RBAC: the
	// group first resolves an identity (Auth) and then admits only PARTNER and
	// ADMIN roles (RequireRoles). A guest or an ordinary shopper is rejected with
	// a 403 "Access denied." — they can view and pick stylists via the public
	// roster above, but never create, edit, publish or delete one.
	partner := api.Group("/partner/personas")
	partner.Use(middleware.Auth(authService, tokens, *cfg, log))
	partner.Use(middleware.RequireRoles(models.RolePartner, models.RoleAdmin))
	{
		partner.GET("/schema", h.Persona.Schema)
		partner.GET("", h.Persona.ListAll)
		partner.GET("/:slug", h.Persona.Get)
		partner.POST("", h.Persona.Create)
		partner.PUT("/:slug", h.Persona.Update)
		partner.DELETE("/:slug", h.Persona.Delete)
		partner.POST("/:slug/publish", h.Persona.Publish)
	}

	// --- Identity-scoped routes -----------------------------------------------
	// Everything below resolves an identity — a real account or a minted guest
	// — and is where the product's per-shopper state lives.
	authed := api.Group("")
	authed.Use(middleware.Auth(authService, tokens, *cfg, log))
	{
		authed.GET("/profile", h.Auth.Profile)

		assistant := authed.Group("/assistant")
		// The assistant gets its own, tighter bucket: every call can cost a
		// Gemini request, so its budget is independent of the general API
		// limit above.
		assistant.Use(middleware.RateLimit(cfg.RateLimit.AssistantRPS, cfg.RateLimit.AssistantBurst))
		{
			assistant.POST("/chat", h.Assistant.Chat)
			assistant.GET("/recommendations/:id", h.Assistant.Replay)
			assistant.POST("/collection", h.Assistant.Collection)
			assistant.POST("/collections", h.Assistant.Collections)
		}

		history := authed.Group("/history")
		{
			history.GET("", h.History.List)
			history.POST("", h.History.Create)
			history.PATCH("/:id", h.History.UpdateStatus)
			history.DELETE("/:id", h.History.Delete)
		}

		recommendations := authed.Group("/recommendations")
		{
			recommendations.GET("", h.Recommendation.List)
			recommendations.GET("/:id", h.Recommendation.GetByID)
		}

		favorites := authed.Group("/favorites")
		{
			favorites.GET("", h.Favorite.List)
			favorites.POST("", h.Favorite.Add)
			favorites.DELETE("/:id", h.Favorite.Remove)
		}

		// Per-user cart and collection stores. These live under the authed
		// group so every request carries an identity; the handlers scope every
		// read and write to it, which is what keeps one shopper's cart and
		// wishlist invisible to another.
		cart := authed.Group("/cart")
		{
			cart.GET("", h.Cart.List)
			cart.POST("", h.Cart.Add)
			cart.DELETE("", h.Cart.Clear)
			cart.DELETE("/:productId", h.Cart.Remove)
		}

		savedCollections := authed.Group("/saved-collections")
		{
			savedCollections.GET("", h.Cart.ListSaved)
			savedCollections.POST("", h.Cart.Save)
			savedCollections.DELETE("/:id", h.Cart.DeleteSaved)
		}

		collectionCart := authed.Group("/collection-cart")
		{
			collectionCart.GET("", h.Cart.ListCollectionCart)
			collectionCart.POST("", h.Cart.AddCollection)
			collectionCart.DELETE("", h.Cart.ClearCollectionCart)
			collectionCart.PUT("/:id", h.Cart.UpdateCollection)
			collectionCart.DELETE("/:id", h.Cart.DeleteCollection)
		}

		// Personalization profile. Grouped under /profile so it sits alongside
		// the account profile (GET /profile) without colliding — these are the
		// shopping-preference sub-resources, all identity-scoped.
		profile := authed.Group("/profile")
		{
			profile.GET("/preferences", h.Profile.Get)
			profile.PUT("/preferences", h.Profile.Update)
			profile.GET("/status", h.Profile.Status)
			profile.GET("/orders", h.Profile.Orders)
		}

		authed.GET("/dashboard/stats", h.Dashboard.Stats)

		// StarCloset V2 per-department unlock progress. Identity-scoped: each
		// shopper advances and reads only their own three independent journeys.
		starcloset := authed.Group("/starcloset")
		{
			starcloset.GET("/progress", h.StarCloset.GetProgress)
			starcloset.POST("/progress", h.StarCloset.Increment)
		}
	}
}
