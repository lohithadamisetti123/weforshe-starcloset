package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/myntra/bharat-ai-backend/internal/config"
)

// CORS configures cross-origin access for the Vite dev server and any
// configured deployment origins.
//
// The origin list is explicit configuration, never "*". A wildcard would also
// be incompatible with credentialed requests, but the real reason is that the
// allowed origins are a deployment fact and belong in deployment config, where
// changing them is a decision someone makes rather than a default nobody
// noticed.
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			// X-Session-Id carries the guest identity. Without it in this list
			// the browser silently strips the header on cross-origin requests
			// and every guest looks like a brand-new shopper on each call.
			"X-Session-Id", "X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			// The client needs to read these to back off politely rather than
			// hammer a 429 in a retry loop.
			"X-RateLimit-Limit", "X-RateLimit-Remaining", "Retry-After",
		},
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           12 * time.Hour,
	})
}
