package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// The response envelope.
//
// Every endpoint returns the same outer shape, so the frontend can write one
// unwrap helper and one error path instead of special-casing each route. The
// payload *inside* `data` is what the frontend's TypeScript types describe, and
// it is kept byte-identical to those types — see internal/dto.

// Response is the success envelope.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// ErrorResponse is the failure envelope. Its shape matches the brief:
//
//	{"success": false, "message": "Product not found", "error": {}}
type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail carries the machine-readable part of a failure.
type ErrorDetail struct {
	Code      string `json:"code,omitempty"`
	Details   any    `json:"details,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// PageMeta describes a paginated collection.
type PageMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// NewPageMeta computes pagination metadata.
func NewPageMeta(page, pageSize int, total int64) PageMeta {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return PageMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// OK writes a 200 with data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

// OKWithMeta writes a 200 with data and metadata.
func OKWithMeta(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data, Meta: meta})
}

// OKMessage writes a 200 with a message and optional data.
func OKMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{Success: true, Message: message, Data: data})
}

// Created writes a 201.
func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, Response{Success: true, Message: message, Data: data})
}

// NoContent writes a 204.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Fail attaches an error to the Gin context for the error middleware to render.
//
// Handlers call this and return; they never write an error body themselves.
// Centralising the rendering is what guarantees a 404 from the favourites
// handler looks exactly like a 404 from the products handler.
func Fail(c *gin.Context, err error) {
	//nolint:errcheck // Gin's Error only fails on a nil argument.
	_ = c.Error(err)
	c.Abort()
}

// RequestIDKey is the context key and header name for the correlation ID.
const RequestIDKey = "X-Request-ID"

// RequestIDFrom pulls the correlation ID off the context.
func RequestIDFrom(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
