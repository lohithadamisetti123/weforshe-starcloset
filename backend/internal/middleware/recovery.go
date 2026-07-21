package middleware

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// Recovery turns a panic into a logged 500 instead of a dead process.
//
// Gin ships its own, but it prints to stdout in a format nothing parses and it
// does not correlate the crash with a request ID. This one logs through zap with
// the stack and the request ID attached, so a panic is a searchable event rather
// than a wall of text someone scrolls past.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			// A client that hangs up mid-write makes the TCP stack panic on the
			// next write. That is the network being the network, not a bug: log
			// it quietly and do not attempt a response nobody is listening for.
			if isBrokenPipe(rec) {
				log.Warn("client disconnected mid-response",
					zap.String("path", c.Request.URL.Path),
					zap.Any("error", rec))
				c.Abort()
				return
			}

			dump, _ := httputil.DumpRequest(c.Request, false)
			log.Error("panic recovered",
				zap.Any("panic", rec),
				zap.String("requestId", utils.RequestIDFrom(c)),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("request", string(dump)),
				zap.Stack("stack"))

			c.AbortWithStatusJSON(http.StatusInternalServerError, utils.ErrorResponse{
				Success: false,
				Message: "Something went wrong on our end",
				Error: utils.ErrorDetail{
					Code:      utils.CodeInternal,
					RequestID: utils.RequestIDFrom(c),
				},
			})
		}()

		c.Next()
	}
}

// isBrokenPipe reports whether a panic value is a severed connection.
func isBrokenPipe(rec any) bool {
	ne, ok := rec.(*net.OpError)
	if !ok {
		return false
	}
	se, ok := ne.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	msg := strings.ToLower(se.Error())
	return strings.Contains(msg, "broken pipe") || strings.Contains(msg, "connection reset by peer")
}
