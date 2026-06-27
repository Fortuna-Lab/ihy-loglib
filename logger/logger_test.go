package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestInfoAutoInitGoroutineSession(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
		ReleaseGoroutineSession()
	})

	Info("worker loop started", "worker_id", 1)
	Info("worker loop tick", "worker_id", 1)

	lines := bytes.Split(bytes.TrimSpace(buffer.Bytes()), []byte("\n"))
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}

	var first, second LogEntry
	if err := json.Unmarshal(lines[0], &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(lines[1], &second); err != nil {
		t.Fatal(err)
	}
	if first.SessionID == "" {
		t.Fatal("expected auto-generated log_session_id")
	}
	if first.SessionID != second.SessionID {
		t.Fatalf("session IDs differ: %q vs %q", first.SessionID, second.SessionID)
	}
	if first.Message != "worker loop started" {
		t.Fatalf("Message = %q", first.Message)
	}
}

func TestExplicitSessionOverridesSingleLogOnly(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
		ReleaseGoroutineSession()
	})

	Info("worker loop started", "worker_id", 1)
	Info("batch received", "log_session_id", "batch-session", "count", 10)
	Info("worker idle", "worker_id", 1)

	lines := bytes.Split(bytes.TrimSpace(buffer.Bytes()), []byte("\n"))
	var entries [3]LogEntry
	for i, line := range lines {
		if err := json.Unmarshal(line, &entries[i]); err != nil {
			t.Fatal(err)
		}
	}

	if entries[1].SessionID != "batch-session" {
		t.Fatalf("batch log session = %q, want batch-session", entries[1].SessionID)
	}
	if entries[2].SessionID != entries[0].SessionID {
		t.Fatalf("worker session should continue after batch log: %q vs %q", entries[2].SessionID, entries[0].SessionID)
	}
}

func TestBindSessionUsesProvidedID(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
	})

	unbind := BindSession("worker-session-42")
	defer unbind()

	Info("worker loop started", "worker_id", 1)

	var entry LogEntry
	if err := json.Unmarshal(buffer.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.SessionID != "worker-session-42" {
		t.Fatalf("SessionID = %q, want worker-session-42", entry.SessionID)
	}
}

func TestInfoWritesSessionIDAndFields(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
		ReleaseGoroutineSession()
	})

	Info("test message", "log_session_id", "session-abc", "username", "user@example.com")

	var entry LogEntry
	if err := json.Unmarshal(buffer.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Level != "INFO" {
		t.Fatalf("Level = %q, want %q", entry.Level, "INFO")
	}
	if entry.Message != "test message" {
		t.Fatalf("Message = %q, want %q", entry.Message, "test message")
	}
	if entry.SessionID != "session-abc" {
		t.Fatalf("SessionID = %q, want %q", entry.SessionID, "session-abc")
	}
	if entry.Fields != `{"username":"user@example.com"}` {
		t.Fatalf("Fields = %q, want %q", entry.Fields, `{"username":"user@example.com"}`)
	}
}

func TestInfoCtxUsesSessionFromContext(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
	})

	ctx, sessionID := BeginSession(context.Background(), "session-from-ctx")
	InfoCtx(ctx, "task started", "step", 1)

	var entry LogEntry
	if err := json.Unmarshal(buffer.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.SessionID != sessionID {
		t.Fatalf("SessionID = %q, want %q", entry.SessionID, sessionID)
	}
}

func TestBeginSessionGeneratesIDWhenEmpty(t *testing.T) {
	ctx, sessionID := BeginSession(context.Background(), "")
	if sessionID == "" {
		t.Fatal("expected generated session ID")
	}
	if SessionIDFromContext(ctx) != sessionID {
		t.Fatalf("SessionIDFromContext = %q, want %q", SessionIDFromContext(ctx), sessionID)
	}
}

func TestBeginSessionPreservesProvidedID(t *testing.T) {
	ctx, sessionID := BeginSession(context.Background(), "client-session-123")
	if sessionID != "client-session-123" {
		t.Fatalf("sessionID = %q, want client-session-123", sessionID)
	}
	if SessionIDFromContext(ctx) != "client-session-123" {
		t.Fatalf("SessionIDFromContext = %q", SessionIDFromContext(ctx))
	}
}

func TestInfoCtxExplicitSessionOverridesContext(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
	})

	ctx, _ := BeginSession(context.Background(), "ctx-session")
	InfoCtx(ctx, "override test", "log_session_id", "explicit-session")

	var entry LogEntry
	if err := json.Unmarshal(buffer.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.SessionID != "explicit-session" {
		t.Fatalf("SessionID = %q, want explicit-session", entry.SessionID)
	}
}

func TestRunWithSession(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
	})

	RunWithSession(context.Background(), "worker-session", func(ctx context.Context) {
		InfoCtx(ctx, "processing")
		InfoCtx(ctx, "done")
	})

	lines := bytes.Split(bytes.TrimSpace(buffer.Bytes()), []byte("\n"))
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}
	for _, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Fatal(err)
		}
		if entry.SessionID != "worker-session" {
			t.Fatalf("SessionID = %q, want worker-session", entry.SessionID)
		}
	}
}

func TestInfoMasksSensitiveFields(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
	})

	Info("login", "password", "secret123", "token", "abc")

	var raw map[string]interface{}
	if err := json.Unmarshal(buffer.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	fields, ok := raw["fields"].(string)
	if !ok {
		t.Fatalf("fields should be a JSON string, got %T: %v", raw["fields"], raw["fields"])
	}
	if !strings.Contains(fields, `"password"`) || strings.Contains(fields, "secret123") {
		t.Fatalf("password should be masked in fields string: %s", fields)
	}
	if strings.Contains(fields, `"abc"`) {
		t.Fatalf("token should be masked in fields string: %s", fields)
	}
}

func TestLogAccessWritesRequestAndSessionID(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.access
	global.outs.access = &buffer
	t.Cleanup(func() {
		global.outs.access = original
	})

	LogAccess("GET", "/health", 200, "1ms", `{"ok":true}`, "request-123", "session-abc")

	var raw map[string]interface{}
	if err := json.Unmarshal(buffer.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	body, ok := raw["body"].(string)
	if !ok {
		t.Fatalf("body should be a JSON string, got %T: %v", raw["body"], raw["body"])
	}
	if body != `{"ok":true}` {
		t.Fatalf("body = %q, want %q", body, `{"ok":true}`)
	}

	var entry AccessLogEntry
	if err := json.Unmarshal(buffer.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.RequestID != "request-123" {
		t.Fatalf("RequestID = %q, want %q", entry.RequestID, "request-123")
	}
	if entry.SessionID != "session-abc" {
		t.Fatalf("SessionID = %q, want %q", entry.SessionID, "session-abc")
	}
	if entry.Method != "GET" || entry.Path != "/health" || entry.Status != 200 {
		t.Fatalf("access entry = %#v", entry)
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("LOG_DIR", "/tmp/logs")
	t.Setenv("LOG_SERVICE_NAME", "identity")
	t.Setenv("LOG_MIRROR_STDOUT", "false")

	cfg := ConfigFromEnv()
	if cfg.Dir != "/tmp/logs" {
		t.Fatalf("Dir = %q, want /tmp/logs", cfg.Dir)
	}
	if cfg.ServiceName != "identity" {
		t.Fatalf("ServiceName = %q, want identity", cfg.ServiceName)
	}
	if cfg.MirrorToStdout {
		t.Fatal("expected MirrorToStdout=false")
	}
}
