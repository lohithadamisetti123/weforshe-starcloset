package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// Logger writes one structured line per request.
//
// The level is chosen from the status: a 500 is an error, a 4xx is a warning
// worth seeing but not paging on, and a 2xx is info. Logging everything at the
// same level is how a log becomes something people filter out entirely.
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		// Health checks are called every few seconds by the container runtime
		// and say nothing. Logging them buries the requests that matter under
		// a stream of 200s.
		if path == "/health" || path == "/health/live" || path == "/health/ready" {
			return
		}

		status := c.Writer.Status()
		fields := []zap.Field{
			zap.String("requestId", utils.RequestIDFrom(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		}
		if query != "" {
			fields = append(fields, zap.String("query", query))
		}
		if userID := UserIDString(c); userID != "" {
			fields = append(fields, zap.String("userId", userID))
		}

		msg := "request"
		switch {
		case status >= 500:
			log.Log(zapcore.ErrorLevel, msg, fields...)
		case status >= 400:
			log.Log(zapcore.WarnLevel, msg, fields...)
		default:
			log.Log(zapcore.InfoLevel, msg, fields...)
		}
	}
}
