// Package handlers is the HTTP boundary. Handlers do exactly three things:
// bind and validate the request, call one service method, and render the
// result through the response envelope. Nothing that decides application
// behaviour belongs here — that discipline is what keeps a handler readable in
// one screen and keeps the business logic reachable from a test that never
// imports "net/http".
package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
)

// bindJSON decodes a JSON body and runs full validation.
//
// Two passes look redundant but are not: gin's ShouldBindJSON only enforces
// `binding:"required"` (presence), which is exactly enough to catch a
// malformed request before an empty string reaches business logic. The
// `validator` pass catches everything with a shape opinion — email format,
// length bounds, oneof membership — and renders it as the per-field message a
// form can attach to an input, which gin's own binding error does not give.
func bindJSON(c *gin.Context, val *validator.Validator, dest any) bool {
	if err := c.ShouldBindJSON(dest); err != nil {
		utils.Fail(c, utils.NewBadRequest("The request body could not be understood", err))
		return false
	}
	if err := val.Struct(dest); err != nil {
		utils.Fail(c, err)
		return false
	}
	return true
}

// bindQuery decodes a query string and runs full validation.
func bindQuery(c *gin.Context, val *validator.Validator, dest any) bool {
	if err := c.ShouldBindQuery(dest); err != nil {
		utils.Fail(c, utils.NewBadRequest("The query parameters could not be understood", err))
		return false
	}
	if err := val.Struct(dest); err != nil {
		utils.Fail(c, err)
		return false
	}
	return true
}
