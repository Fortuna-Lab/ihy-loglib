package logger

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var goroutineSessions sync.Map

func goid() uint64 {
	var buf [32]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseUint(idField, 10, 64)
	return id
}

// BindSession binds a log session to the current goroutine until cleanup is called.
// If sessionID is empty, a new UUID is generated.
func BindSession(sessionID string) func() {
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	id := goid()
	goroutineSessions.Store(id, sessionID)
	return func() {
		goroutineSessions.Delete(id)
	}
}

// ReleaseGoroutineSession removes the log session bound to the current goroutine.
func ReleaseGoroutineSession() {
	goroutineSessions.Delete(goid())
}

func ensureGoroutineSession() string {
	id := goid()
	if v, ok := goroutineSessions.Load(id); ok {
		if sessionID, ok := v.(string); ok && sessionID != "" {
			return sessionID
		}
	}
	sessionID := uuid.New().String()
	goroutineSessions.Store(id, sessionID)
	return sessionID
}

func sessionIDFromGoroutine() string {
	v, ok := goroutineSessions.Load(goid())
	if !ok {
		return ""
	}
	sessionID, _ := v.(string)
	return sessionID
}

func resolveSessionID(ctx context.Context, keysAndValues ...interface{}) string {
	var explicitID string
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok || key != "log_session_id" {
			continue
		}
		if v, ok := keysAndValues[i+1].(string); ok && v != "" {
			explicitID = v
			break
		}
	}
	if explicitID != "" {
		return explicitID
	}
	if sessionID := SessionIDFromContext(ctx); sessionID != "" {
		return sessionID
	}
	if sessionID := sessionIDFromGoroutine(); sessionID != "" {
		return sessionID
	}
	return ensureGoroutineSession()
}
