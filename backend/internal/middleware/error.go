package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// ErrorHandler renders every error the handlers attach to the context.
//
// Handlers call utils.Fail(c, err) and return; nothing below this line writes an
// error body itself. Centralising it is what guarantees the envelope is
// identical for a validation failure and a database outage — a client that has
// to parse two error shapes will eventually parse one of them wrong.
//
// It also draws the line between what is logged and what is sent. A 5xx logs at
// error with its cause chain intact and returns a generic sentence; a 4xx logs
// at debug and returns the real message. The client learns what they can fix and
// nothing about our schema.
func ErrorHandler(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		// The last error is the one that aborted the chain; earlier entries are
		// context Gin collected on the way.
		err := c.Errors.Last().Err
		appErr := utils.AsAppError(err)
		requestID := utils.RequestIDFrom(c)

		fields := []zap.Field{
			zap.String("requestId", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", appErr.Status),
			zap.String("code", appErr.Code),
		}

		if appErr.Status >= 500 {
			// Unwrap() gives the wrapped cause, which for an AppError built by
			// NewInternal is the real database or network error. That is the
			// thing worth having in the log and the thing that must never be in
			// the response.
			fields = append(fields, zap.Error(err))
			log.Error("request failed", fields...)
		} else {
			log.Debug("request rejected", append(fields, zap.String("reason", appErr.Message))...)
		}

		c.JSON(appErr.Status, utils.ErrorResponse{
			Success: false,
			Message: appErr.Message,
			Error: utils.ErrorDetail{
				Code:      appErr.Code,
				Details:   appErr.Details,
				RequestID: requestID,
			},
		})
	}
}
