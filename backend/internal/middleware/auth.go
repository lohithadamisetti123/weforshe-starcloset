package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/config"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// Context keys for the authenticated identity.
const (
	ctxUserID  = "auth.userId"
	ctxEmail   = "auth.email"
	ctxRole    = "auth.role"
	ctxIsGuest = "auth.isGuest"
)

// sessionHeader carries the browser-scoped guest identity.
const sessionHeader = "X-Session-Id"

// Auth resolves the caller's identity, minting a guest when permitted.
//
// The awkward fact this middleware exists to handle: the shipped frontend has
// no sign-in screen. It was built against mocks and every page assumes it can
// just ask. Three options were available — bolt a login UI onto a finished
// frontend, let the assistant write to a shared anonymous bucket, or give each
// browser a real but unregistered account.
//
// The first modifies a frontend that is explicitly out of scope. The second is
// worse than it sounds: one shopper's history and favourites would be visible
// to every other visitor, which is a data leak dressed up as a demo shortcut.
// The third is what runs here. A guest is a real user row keyed off a stable
// browser-supplied session ID, so every downstream query is ordinary
// user-scoped code and the isolation is enforced by the same WHERE clause that
// protects registered accounts.
//
// A bearer token always wins when present, so the moment a login UI does appear,
// this path stops being reached without a line changing.
func Auth(authService *services.AuthService, tokens *utils.TokenManager, cfg config.Config, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token := bearerToken(c); token != "" {
			claims, err := tokens.Verify(token, utils.TokenAccess)
			if err != nil {
				utils.Fail(c, authError(err))
				return
			}

			c.Set(ctxUserID, claims.UserID)
			c.Set(ctxEmail, claims.Email)
			c.Set(ctxRole, claims.Role)
			c.Set(ctxIsGuest, claims.IsGuest)
			c.Next()
			return
		}

		if !cfg.Server.AllowGuest {
			utils.Fail(c, utils.NewUnauthorized("Authentication required"))
			return
		}

		sessionID := strings.TrimSpace(c.GetHeader(sessionHeader))
		if sessionID == "" {
			// Naming the header in the message matters. Without it the client
			// author sees a bare 401 on an endpoint that has no login form and
			// concludes the API is broken.
			utils.Fail(c, utils.NewUnauthorized(
				"Provide a bearer token, or an "+sessionHeader+" header to continue as a guest"))
			return
		}
		if len(sessionID) > 128 {
			utils.Fail(c, utils.NewBadRequest(sessionHeader+" is too long", nil))
			return
		}

		user, _, err := authService.EnsureGuest(c.Request.Context(), sessionID)
		if err != nil {
			utils.Fail(c, err)
			return
		}

		c.Set(ctxUserID, user.ID)
		c.Set(ctxEmail, user.Email)
		c.Set(ctxRole, string(user.Role))
		c.Set(ctxIsGuest, true)
		c.Next()
	}
}

// RequireRegistered rejects guests.
//
// Guests can shop, ask and save; they cannot do anything that presumes a real
// account. Nothing in the MVP needs it, but the routes that would — changing a
// password, deleting an account — should not have to remember to check.
func RequireRegistered() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsGuest(c) {
			utils.Fail(c, utils.NewForbidden("Create an account to use this feature"))
			return
		}
		c.Next()
	}
}

// RequireAdmin rejects non-admins.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, _ := c.Get(ctxRole); role != "Admin" {
			utils.Fail(c, utils.NewForbidden("Administrator access required"))
			return
		}
		c.Next()
	}
}

// bearerToken extracts a token from the Authorization header.
func bearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	// EqualFold, not a "Bearer " prefix comparison: the scheme is
	// case-insensitive per RFC 7235, and a client sending "bearer" is correct
	// even though it is unusual.
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// authError maps a token failure to a message the client can act on.
//
// "Expired" and "invalid" lead to different client behaviour — refresh versus
// re-authenticate — so collapsing them into one 401 would force the client to
// guess, and it would guess wrong by looping on a refresh that cannot succeed.
func authError(err error) error {
	switch {
	case errors.Is(err, utils.ErrTokenExpired):
		return utils.NewUnauthorized("Your session has expired. Please sign in again")
	case errors.Is(err, utils.ErrTokenWrongUse):
		return utils.NewUnauthorized("That token cannot be used to authenticate requests")
	default:
		return utils.NewUnauthorized("Invalid authentication token")
	}
}

// UserID returns the authenticated shopper's ID.
//
// Handlers call this after Auth has run, so the identity is always present. The
// ok result exists for the one caller that runs on an optionally-authenticated
// route; everything else treats absence as a bug rather than a case.
func UserID(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get(ctxUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// MustUserID returns the shopper's ID or fails the request.
func MustUserID(c *gin.Context) (uuid.UUID, bool) {
	id, ok := UserID(c)
	if !ok {
		utils.Fail(c, utils.NewUnauthorized("Authentication required"))
		return uuid.Nil, false
	}
	return id, true
}

// UserIDString returns the shopper's ID for logging, or "".
func UserIDString(c *gin.Context) string {
	if id, ok := UserID(c); ok {
		return id.String()
	}
	return ""
}

// IsGuest reports whether the caller is an unregistered browser session.
func IsGuest(c *gin.Context) bool {
	v, _ := c.Get(ctxIsGuest)
	guest, _ := v.(bool)
	return guest
}
