package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

var reservedLogKeys = map[string]struct{}{
	"time":           {},
	"level":          {},
	"message":        {},
	"fields":         {},
	"log_session_id": {},
	"service":        {},
}

type fieldPair struct {
	key   string
	value string
}

func writeFlatLine(w io.Writer, fields map[string]string) {
	if len(fields) == 0 {
		return
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return
	}
	_, _ = w.Write(append(b, '\n'))
}

func toFieldString(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return fmt.Sprint(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(b)
	}
}

// formatFieldPairs renders key-value pairs as plain text: "key: value, key: value".
func formatFieldPairs(pairs []fieldPair) string {
	if len(pairs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		if pair.key == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", pair.key, pair.value))
	}
	return strings.Join(parts, ", ")
}

// formatFieldsFromJSON renders JSON object text as plain "key: value, key: value".
func formatFieldsFromJSON(raw string) string {
	if raw == "" {
		return ""
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return raw
	}

	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]fieldPair, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, fieldPair{key: k, value: toFieldString(parsed[k])})
	}
	return formatFieldPairs(pairs)
}
