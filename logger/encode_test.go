package logger

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLogEntryFlattensFieldsToTopLevel(t *testing.T) {
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
		t.Fatalf("fields wrapper must not exist, got: %s", raw)
	}
	if bytes.Contains(raw, []byte(`{`)) && bytes.Contains(raw, []byte(`"batch_size"`)) {
		// only allowed braces are the outer JSON object
		var doc map[string]interface{}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		for k, v := range doc {
			if k == "time" || k == "level" || k == "message" || k == "log_session_id" {
				continue
			}
			if _, ok := v.(string); !ok {
				t.Fatalf("field %q type = %T, want string", k, v)
			}
		}
	}

	var doc map[string]string
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["message"] != "translation history worker started" {
		t.Fatalf("message = %q", doc["message"])
	}
	if doc["batch_size"] != "100" {
		t.Fatalf("batch_size = %q, want 100", doc["batch_size"])
	}
	if doc["concurrency"] != "5" {
		t.Fatalf("concurrency = %q, want 5", doc["concurrency"])
	}
	if doc["key"] != "ihy:log:translation_history" {
		t.Fatalf("key = %q", doc["key"])
	}
}

func TestAccessLogFlattensBodyToTopLevel(t *testing.T) {
	var buffer bytes.Buffer
	original := global.outs.access
	global.outs.access = &buffer
	t.Cleanup(func() {
		global.outs.access = original
	})

	LogAccess("POST", "/api/login", 200, "12ms", `{"username":"user@example.com","password":"******"}`, "req-1", "sess-1")

	raw := buffer.Bytes()
	if bytes.Contains(raw, []byte(`"body":`)) {
		t.Fatalf("body wrapper must not exist, got: %s", raw)
	}
	if bytes.Contains(raw, []byte(`"body":{`)) {
		t.Fatalf("body must not be object, got: %s", raw)
	}

	var doc map[string]string
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["body_username"] != "user@example.com" {
		t.Fatalf("body_username = %q", doc["body_username"])
	}
	if doc["body_password"] != "******" {
		t.Fatalf("body_password = %q", doc["body_password"])
	}
	if doc["status"] != "200" {
		t.Fatalf("status = %q, want 200", doc["status"])
	}
}

func TestAccessLogRawBodyUsesPlainBodyField(t *testing.T) {
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
	if doc["body"] != "not-json-body" {
		t.Fatalf("body = %q, want not-json-body", doc["body"])
	}
}
