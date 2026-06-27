package logger

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestInfoWritesSessionIDAndFields(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
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
	if got := entry.Fields["username"]; got != "user@example.com" {
		t.Fatalf("Fields[username] = %v, want %q", got, "user@example.com")
	}
	if _, ok := entry.Fields["log_session_id"]; ok {
		t.Fatalf("log_session_id should be top-level, not inside fields")
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

	var entry LogEntry
	if err := json.Unmarshal(buffer.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Fields["password"] == "secret123" {
		t.Fatalf("password should be masked, got %v", entry.Fields["password"])
	}
	if entry.Fields["token"] == "abc" {
		t.Fatalf("token should be masked, got %v", entry.Fields["token"])
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
