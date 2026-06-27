package middleware

import (
	"context"

	"github.com/Fortuna-Lab/ihy-loglib/logger"
	"github.com/gofiber/fiber/v2"
)

const (
	DefaultSessionHeader   = "X-Session-ID"
	DefaultSessionLocalKey = "log_session_id"
)

// SessionConfig defines behavior for log session middleware.
type SessionConfig struct {
	// HeaderName reads the incoming session ID (default: X-Session-ID).
	HeaderName string
	// LocalKey stores the session ID in c.Locals (default: log_session_id).
	LocalKey string
	// ResponseHeader echoes the session ID in the response (default: HeaderName).
	ResponseHeader string
}

// SessionID initializes a log session for the request lifecycle.
// Uses the client-provided header when present; otherwise generates a new UUID.
// The same ID is kept on ctx and locals until the request finishes.
func SessionID(cfg SessionConfig) fiber.Handler {
	headerName := cfg.HeaderName
	if headerName == "" {
		headerName = DefaultSessionHeader
	}
	localKey := cfg.LocalKey
	if localKey == "" {
		localKey = DefaultSessionLocalKey
	}
	responseHeader := cfg.ResponseHeader
	if responseHeader == "" {
		responseHeader = headerName
	}

	return func(c *fiber.Ctx) error {
		ctx, sessionID := logger.BeginSession(c.UserContext(), c.Get(headerName))
		c.SetUserContext(ctx)
		c.Locals(localKey, sessionID)
		c.Set(responseHeader, sessionID)
		return c.Next()
	}
}

// Ctx returns the request context carrying log_session_id for logger.InfoCtx and friends.
func Ctx(c *fiber.Ctx) context.Context {
	return c.UserContext()
}
