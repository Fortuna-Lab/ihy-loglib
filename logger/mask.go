package logger

import "strings"

func normalizeValue(key string, value interface{}) interface{} {
	if isSensitiveKey(key) {
		if strValue, ok := value.(string); ok {
			return maskString(strValue)
		}
		return "******"
	}
	if err, ok := value.(error); ok {
		return err.Error()
	}
	return value
}

func isSensitiveKey(key string) bool {
	switch strings.ToLower(key) {
	case "password", "new_password", "client_secret", "token",
		"access_token", "refresh_token", "id_token", "authorization":
		return true
	default:
		return false
	}
}

func maskString(value string) string {
	runes := []rune(value)
	length := len(runes)

	if length == 0 {
		return ""
	}
	if length <= 2 {
		return "******"
	}
	if length <= 5 {
		return string(runes[:1]) + "******"
	}
	return string(runes[:2]) + "******" + string(runes[length-2:])
}
