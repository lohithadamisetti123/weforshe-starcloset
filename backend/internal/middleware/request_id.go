// Package middleware holds the HTTP cross-cutting concerns.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// RequestID assigns every request a correlation ID and echoes it back.
//
// An inbound X-Request-ID is honoured rather than overwritten so a trace
// started at the browser or a proxy survives into these logs. When one shopper
// reports "the assistant showed me sarees for an office prompt", the ID in
// their network tab is the thing that finds the extraction and the score
// breakdown for that exact call.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(utils.RequestIDKey)
		if id == "" {
			id = uuid.NewString()
		}

		c.Set(utils.RequestIDKey, id)
		c.Header(utils.RequestIDKey, id)
		c.Next()
	}
}
