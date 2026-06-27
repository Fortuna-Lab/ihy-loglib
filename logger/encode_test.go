package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLogEntryUsesLogFieldsAndTopLevelKeys(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
		ReleaseGoroutineSession()
	})

	Info("translation history worker started", "batch_size", 100, "concurrency", 5, "key", "ihy:log:translation_history")

	raw := buffer.Bytes()
	if bytes.Contains(raw, []byte(`"fields"`)) {
		t.Fatalf("legacy fields key must not exist, got: %s", raw)
	}

	var doc map[string]string
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	logFields := doc["log_fields"]
	if !strings.Contains(logFields, "batch_size: 100") {
		t.Fatalf("log_fields = %q", logFields)
	}
	if doc["batch_size"] != "100" {
		t.Fatalf("batch_size = %q, want indexed top-level string", doc["batch_size"])
	}
	if doc["concurrency"] != "5" {
		t.Fatalf("concurrency = %q, want 5", doc["concurrency"])
	}
}

func TestAccessLogUsesLogFieldsAndTopLevelKeys(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.access
	global.outs.access = &buffer
	t.Cleanup(func() {
		global.outs.access = original
	})

	LogAccess("POST", "/api/login", 200, "12ms", `{"username":"user@example.com","password":"******"}`, "req-1", "sess-1")

	var doc map[string]string
	if err := json.Unmarshal(buffer.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["fields"]; ok {
		t.Fatal("legacy fields key must not exist")
	}
	if !strings.Contains(doc["log_fields"], "username: user@example.com") {
		t.Fatalf("log_fields = %q", doc["log_fields"])
	}
	if doc["username"] != "user@example.com" {
		t.Fatalf("username = %q, want indexed top-level string", doc["username"])
	}
}

func TestAccessLogRawPayloadUsesLogFields(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.access
	global.outs.access = &buffer
	t.Cleanup(func() {
		global.outs.access = original
	})

	LogAccess("POST", "/upload", 400, "3ms", "not-json-body", "req-2", "sess-2")

	var doc map[string]string
	if err := json.Unmarshal(buffer.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc["log_fields"] != "not-json-body" {
		t.Fatalf("log_fields = %q, want not-json-body", doc["log_fields"])
	}
}

func TestFormatFieldPairs(t *testing.T) {
	got := formatFieldPairs([]fieldPair{
		{key: "worker_id", value: "1"},
		{key: "batch_size", value: "100"},
	})
	want := "worker_id: 1, batch_size: 100"
	if got != want {
		t.Fatalf("formatFieldPairs = %q, want %q", got, want)
	}
}
