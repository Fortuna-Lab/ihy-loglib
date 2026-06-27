package logger

import (
	"encoding/json"
	"fmt"
	"io"
)

var reservedLogKeys = map[string]struct{}{
	"time":           {},
	"level":          {},
	"message":        {},
	"log_session_id": {},
	"service":        {},
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

// flattenBody spreads JSON body keys to top-level string fields (body_<key>).
// Non-JSON body is stored as a plain "body" string field.
func flattenBody(body string) map[string]string {
	if body == "" {
		return nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return map[string]string{"body": body}
	}

	out := make(map[string]string, len(parsed))
	for k, v := range parsed {
		out["body_"+k] = toFieldString(v)
	}
	return out
}

func mergeFlatFields(base map[string]string, extra map[string]string) {
	for k, v := range extra {
		if _, reserved := reservedLogKeys[k]; reserved {
			continue
		}
		if v == "" {
			continue
		}
		base[k] = v
	}
}
