package logger

import (
	"context"

	"github.com/google/uuid"
)

type sessionContextKey struct{}

// BeginSession attaches a log session ID to ctx for the lifetime of the task.
// If sessionID is empty, a new UUID is generated.
func BeginSession(ctx context.Context, sessionID string) (context.Context, string) {
	if ctx == nil {
		ctx = context.Background()
	}
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	return context.WithValue(ctx, sessionContextKey{}, sessionID), sessionID
}

// SessionIDFromContext returns the log session ID from ctx, or empty string.
func SessionIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(sessionContextKey{}).(string)
	return v
}

// RunWithSession runs fn with a scoped log session on ctx.
// If sessionID is empty, a new UUID is generated for the scope.
func RunWithSession(ctx context.Context, sessionID string, fn func(context.Context)) {
	ctx, _ = BeginSession(ctx, sessionID)
	fn(ctx)
}
