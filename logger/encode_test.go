package logger

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLogEntryEncodesMessageAndFieldsAsJSONStrings(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.info
	global.outs.info = &buffer
	t.Cleanup(func() {
		global.outs.info = original
		ReleaseGoroutineSession()
	})

	Info("translation history worker started", "batch_size", 100, "concurrency", 5, "key", "ihy:log:translation_history")

	raw := buffer.Bytes()
	if bytes.Contains(raw, []byte(`"fields":{`)) {
		t.Fatalf("fields must be JSON string, got: %s", raw)
	}
	if !bytes.Contains(raw, []byte(`"message":"translation history worker started"`)) {
		t.Fatalf("message must be JSON string, got: %s", raw)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["message"].(string); !ok {
		t.Fatalf("message type = %T, want string", doc["message"])
	}
	fields, ok := doc["fields"].(string)
	if !ok {
		t.Fatalf("fields type = %T, want string", doc["fields"])
	}

	var parsedFields map[string]string
	if err := json.Unmarshal([]byte(fields), &parsedFields); err != nil {
		t.Fatalf("fields content is not JSON object string: %v", err)
	}
	if parsedFields["batch_size"] != "100" {
		t.Fatalf("batch_size = %q, want %q", parsedFields["batch_size"], "100")
	}
	if parsedFields["concurrency"] != "5" {
		t.Fatalf("concurrency = %q, want %q", parsedFields["concurrency"], "5")
	}
}

func TestAccessLogEncodesBodyAsJSONString(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.access
	global.outs.access = &buffer
	t.Cleanup(func() {
		global.outs.access = original
	})

	LogAccess("POST", "/api/login", 200, "12ms", `{"username":"user@example.com"}`, "req-1", "sess-1")

	raw := buffer.Bytes()
	if bytes.Contains(raw, []byte(`"body":{`)) {
		t.Fatalf("body must be JSON string, got: %s", raw)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	body, ok := doc["body"].(string)
	if !ok {
		t.Fatalf("body type = %T, want string", doc["body"])
	}
	if body != `{"username":"user@example.com"}` {
		t.Fatalf("body = %q", body)
	}
}