package utils

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a domain error carrying the HTTP status and client-safe message
// that should accompany it.
//
// The point is that only the code that detects a problem knows what it means.
// A handler catching a bare `gorm.ErrRecordNotFound` has to guess whether that
// is a 404, a 403, or a bug. Wrapping it at the source in an AppError makes the
// intent explicit and lets one middleware render every failure consistently.
type AppError struct {
	// Status is the HTTP status code to return.
	Status int
	// Code is a stable machine-readable identifier for clients to branch on.
	Code string
	// Message is safe to show a user. It must never leak internals.
	Message string
	// Details carries structured context, such as per-field validation errors.
	Details any
	// Err is the wrapped cause. It is logged, never serialised.
	Err error
}

// Error implements error.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap supports errors.Is and errors.As.
func (e *AppError) Unwrap() error { return e.Err }

// Error codes. Stable identifiers the frontend may switch on.
const (
	CodeBadRequest   = "BAD_REQUEST"
	CodeValidation   = "VALIDATION_ERROR"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeNotFound     = "NOT_FOUND"
	CodeConflict     = "CONFLICT"
	CodeRateLimited  = "RATE_LIMITED"
	CodeInternal     = "INTERNAL_ERROR"
	CodeUnavailable  = "SERVICE_UNAVAILABLE"
)

// NewBadRequest returns a 400.
func NewBadRequest(msg string, err error) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: CodeBadRequest, Message: msg, Err: err}
}

// NewValidation returns a 422 carrying per-field details.
func NewValidation(msg string, details any) *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Code: CodeValidation, Message: msg, Details: details}
}

// NewUnauthorized returns a 401.
func NewUnauthorized(msg string) *AppError {
	if msg == "" {
		msg = "Authentication required"
	}
	return &AppError{Status: http.StatusUnauthorized, Code: CodeUnauthorized, Message: msg}
}

// NewForbidden returns a 403.
func NewForbidden(msg string) *AppError {
	if msg == "" {
		msg = "You do not have permission to perform this action"
	}
	return &AppError{Status: http.StatusForbidden, Code: CodeForbidden, Message: msg}
}

// NewNotFound returns a 404 phrased around the missing resource.
func NewNotFound(resource string) *AppError {
	return &AppError{
		Status:  http.StatusNotFound,
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewConflict returns a 409.
func NewConflict(msg string) *AppError {
	return &AppError{Status: http.StatusConflict, Code: CodeConflict, Message: msg}
}

// NewInternal returns a 500 with a deliberately generic message.
//
// The cause is preserved in Err for the logs but never reaches the client:
// database errors leak schema, and a stack trace on the wire is a gift to
// whoever is probing the service.
func NewInternal(err error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternal,
		Message: "Something went wrong on our end",
		Err:     err,
	}
}

// NewUnavailable returns a 503.
func NewUnavailable(msg string, err error) *AppError {
	if msg == "" {
		msg = "Service temporarily unavailable"
	}
	return &AppError{Status: http.StatusServiceUnavailable, Code: CodeUnavailable, Message: msg, Err: err}
}

// AsAppError extracts an *AppError from err, converting anything unrecognised
// into a 500 so that callers always have a status to work with.
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return NewInternal(err)
}
