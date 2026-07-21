package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/myntra/bharat-ai-backend/database"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
)

// HealthHandler answers container-orchestration probes.
//
// Three distinct endpoints exist because "is the process up" and "can it serve
// traffic" are different questions with different consequences for a wrong
// answer. Collapsing them into one /health that checks the database means a
// slow Postgres failover restarts a perfectly healthy process — the restart
// does not fix the database and now there is a cold start on top of the
// outage.
type HealthHandler struct {
	db    *gorm.DB
	cache *repositories.Cache
}

// NewHealthHandler constructs a HealthHandler.
func NewHealthHandler(db *gorm.DB, cache *repositories.Cache) *HealthHandler {
	return &HealthHandler{db: db, cache: cache}
}

// Health handles GET /health — a liveness probe with no dependency checks.
//
//	@Summary		Liveness probe
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Live handles GET /health/live — identical to Health, named for orchestrators
// that expect the /live convention (Kubernetes-style).
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready handles GET /health/ready — a readiness probe that verifies every
// dependency the service actually needs to serve a request.
//
// Redis is checked but never fails readiness: every repository degrades to a
// database read when the cache is unavailable (see repositories.Cache), so a
// dead cache is a latency regression, not an outage. Reporting it in the body
// means the on-call engineer sees it without it paging anyone.
//
//	@Summary		Readiness probe
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Failure		503	{object}	map[string]any
//	@Router			/health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	dbErr := database.Health(ctx, h.db)

	cacheStatus := "disabled"
	if h.cache.Enabled() {
		if err := h.cache.Ping(ctx); err != nil {
			cacheStatus = "unreachable"
		} else {
			cacheStatus = "ok"
		}
	}

	if dbErr != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unavailable",
			"database": "unreachable",
			"cache":    cacheStatus,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": "ok",
		"cache":    cacheStatus,
	})
}
