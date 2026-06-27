package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLogEntryUsesPlainTextFields(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
		ReleaseGoroutineSession()
	})

	Info("translation history worker started", "batch_size", 100, "concurrency", 5, "key", "ihy:log:translation_history")

	raw := buffer.Bytes()
	if bytes.Contains(raw, []byte(`"batch_size"`)) && !bytes.Contains(raw, []byte(`"fields"`)) {
		t.Fatalf("extra data should be in fields string, got: %s", raw)
	}

	var doc map[string]string
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["message"] != "translation history worker started" {
		t.Fatalf("message = %q", doc["message"])
	}
	fields := doc["fields"]
	if !strings.Contains(fields, "batch_size: 100") {
		t.Fatalf("fields = %q, want batch_size: 100", fields)
	}
	if !strings.Contains(fields, "concurrency: 5") {
		t.Fatalf("fields = %q, want concurrency: 5", fields)
	}
	if !strings.Contains(fields, "key: ihy:log:translation_history") {
		t.Fatalf("fields = %q", fields)
	}
	if strings.Contains(fields, "{") {
		t.Fatalf("fields must not contain JSON object braces: %q", fields)
	}
}

func TestAccessLogUsesFieldsLikeInfoLog(t *testing.T) {
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
	if _, ok := doc["body"]; ok {
		t.Fatal("access log should use fields, not body")
	}
	fields := doc["fields"]
	if !strings.Contains(fields, "password: ******") {
		t.Fatalf("fields = %q", fields)
	}
	if !strings.Contains(fields, "username: user@example.com") {
		t.Fatalf("fields = %q", fields)
	}
	if strings.Contains(fields, "{") {
		t.Fatalf("fields must not contain JSON object braces: %q", fields)
	}
}

func TestAccessLogRawPayloadUsesFields(t *testing.T) {
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
	if doc["fields"] != "not-json-body" {
		t.Fatalf("fields = %q, want not-json-body", doc["fields"])
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
