package middleware

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Fortuna-Lab/ihy-loglib/logger"
	"github.com/gofiber/fiber/v2"
)

// FiberConfig defines behavior for API logger middleware.
type FiberConfig struct {
	SensitiveFields   []string
	RequestIDLocalKey string
	SessionIDLocalKey string
}

// APILogger logs request metadata and masked request body to logger.LogAccess.
func APILogger(cfg FiberConfig) fiber.Handler {
	sensitiveSet := make(map[string]struct{}, len(cfg.SensitiveFields))
	for _, k := range cfg.SensitiveFields {
		sensitiveSet[strings.ToLower(k)] = struct{}{}
	}
	if len(sensitiveSet) == 0 {
		for _, k := range []string{"password", "client_secret", "token"} {
			sensitiveSet[k] = struct{}{}
		}
	}

	requestIDKey := cfg.RequestIDLocalKey
	if requestIDKey == "" {
		requestIDKey = "request_id"
	}
	sessionIDKey := cfg.SessionIDLocalKey
	if sessionIDKey == "" {
		sessionIDKey = "log_session_id"
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
			}
		}

		var bodyStr string
		bodyBytes := c.Body()
		if len(bodyBytes) > 0 {
			var body map[string]interface{}
			if json.Unmarshal(bodyBytes, &body) == nil {
				maskFields(body, sensitiveSet)
				if b, err := json.Marshal(body); err == nil {
					bodyStr = string(b)
				}
			} else {
				bodyStr = string(bodyBytes)
			}
		}

		requestID, _ := c.Locals(requestIDKey).(string)
		sessionID, _ := c.Locals(sessionIDKey).(string)
		logger.LogAccess(c.Method(), c.Path(), status, time.Since(start).String(), bodyStr, requestID, sessionID)
		return err
	}
}

func maskFields(v interface{}, sensitiveSet map[string]struct{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, item := range val {
			if _, ok := sensitiveSet[strings.ToLower(k)]; ok {
				if strVal, ok := item.(string); ok {
					val[k] = maskString(strVal)
				} else {
					val[k] = "******"
				}
				continue
			}
			maskFields(item, sensitiveSet)
		}
	case []interface{}:
		for _, item := range val {
			maskFields(item, sensitiveSet)
		}
	}
}

func maskString(s string) string {
	runes := []rune(s)
	l := len(runes)
	if l == 0 {
		return ""
	}
	if l <= 2 {
		return "******"
	}
	if l <= 5 {
		return string(runes[:1]) + "******"
	}
	return string(runes[:2]) + "******" + string(runes[l-2:])
}
