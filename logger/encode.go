package logger

import (
	"encoding/json"
	"fmt"
)

// logString always encodes as a JSON string (never object/number/boolean).
type logString string

func (s logString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func fieldsToJSONString(fields map[string]string) logString {
	if len(fields) == 0 {
		return ""
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return logString("{}")
	}
	return logString(b)
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
	default:
		return fmt.Sprint(v)
	}
}
