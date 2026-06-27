package logger

import (
	"os"
	"strings"
)

// Config controls logger output destinations.
type Config struct {
	// Dir is the directory for log files. Falls back to IDENTITY_LOGS_DIR, then "logs".
	Dir string
	// MirrorToStdout mirrors every log line to stdout in addition to log files.
	MirrorToStdout bool
	// ServiceName is included in JSON logs (useful for Fluent Bit routing).
	ServiceName string
}

// ConfigFromEnv builds Config from environment variables.
//
//   - LOG_DIR or IDENTITY_LOGS_DIR
//   - LOG_MIRROR_STDOUT (default: true)
//   - LOG_SERVICE_NAME or SERVICE_NAME
func ConfigFromEnv() Config {
	return Config{
		Dir:            firstNonEmpty(os.Getenv("LOG_DIR"), os.Getenv("IDENTITY_LOGS_DIR")),
		MirrorToStdout: envBool("LOG_MIRROR_STDOUT", true),
		ServiceName:    firstNonEmpty(os.Getenv("LOG_SERVICE_NAME"), os.Getenv("SERVICE_NAME")),
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func envBool(key string, defaultValue bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultValue
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}
