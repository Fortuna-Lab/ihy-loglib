package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	levelInfo   = "INFO"
	levelWarn   = "WARN"
	levelError  = "ERROR"
	levelAccess = "ACCESS"
)

// LogEntry represents a standard JSON log line.
type LogEntry struct {
	Time        string                 `json:"time"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	SessionID   string                 `json:"log_session_id,omitempty"`
	ServiceName string                 `json:"service,omitempty"`
}

// AccessLogEntry represents API access log data.
type AccessLogEntry struct {
	Time        string                 `json:"time"`
	Level       string                 `json:"level"`
	RequestID   string                 `json:"request_id,omitempty"`
	SessionID   string                 `json:"log_session_id,omitempty"`
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Status      int                    `json:"status"`
	Latency     string                 `json:"latency"`
	Body        string `json:"body,omitempty"`
	ServiceName string                 `json:"service,omitempty"`
}

type levelOutputs struct {
	info   io.Writer
	warn   io.Writer
	error  io.Writer
	access io.Writer
}

type state struct {
	mu sync.RWMutex

	files [4]*os.File
	outs  levelOutputs
	cfg   Config
}

var global = &state{}

// Init initializes logger output files and optional stdout mirroring.
func Init(cfg Config) error {
	dir := cfg.Dir
	if dir == "" {
		dir = os.Getenv("IDENTITY_LOGS_DIR")
	}
	if dir == "" {
		dir = "logs"
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	openFile := func(name string) (*os.File, error) {
		f, err := os.OpenFile(filepath.Join(dir, name), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", name, err)
		}
		return f, nil
	}

	fInfo, err := openFile("info.log")
	if err != nil {
		return err
	}
	fWarn, err := openFile("warning.log")
	if err != nil {
		_ = fInfo.Close()
		return err
	}
	fError, err := openFile("error.log")
	if err != nil {
		_ = fInfo.Close()
		_ = fWarn.Close()
		return err
	}
	fAccess, err := openFile("access.log")
	if err != nil {
		_ = fInfo.Close()
		_ = fWarn.Close()
		_ = fError.Close()
		return err
	}

	buildWriter := func(file *os.File) io.Writer {
		if cfg.MirrorToStdout {
			return io.MultiWriter(os.Stdout, file)
		}
		return file
	}

	global.mu.Lock()
	defer global.mu.Unlock()

	closeStateLocked(global)

	global.files = [4]*os.File{fInfo, fWarn, fError, fAccess}
	global.outs = levelOutputs{
		info:   buildWriter(fInfo),
		warn:   buildWriter(fWarn),
		error:  buildWriter(fError),
		access: buildWriter(fAccess),
	}
	global.cfg = cfg

	return nil
}

// InitFromEnv initializes logger using ConfigFromEnv().
func InitFromEnv() error {
	return Init(ConfigFromEnv())
}

// InitLogger is a compatibility wrapper for code that used the old API.
func InitLogger() {
	_ = Init(Config{MirrorToStdout: true})
}

// Close closes all log files.
func Close() error {
	global.mu.Lock()
	defer global.mu.Unlock()
	return closeStateLocked(global)
}

// CloseAll is a compatibility wrapper for code that used the old API.
func CloseAll() {
	_ = Close()
}

func closeStateLocked(s *state) error {
	errs := make([]error, 0, 4)
	for i, f := range s.files {
		if err := closeFile(f); err != nil {
			errs = append(errs, err)
		}
		s.files[i] = nil
	}
	s.outs = levelOutputs{}
	if len(errs) > 0 {
		return fmt.Errorf("close logger: %v", errs)
	}
	return nil
}

func closeFile(f *os.File) error {
	if f == nil {
		return nil
	}
	return f.Close()
}

// Info writes INFO level JSON log.
func Info(msg string, keysAndValues ...interface{}) {
	writeJSON(levelInfo, msg, keysAndValues...)
}

// Warn writes WARN level JSON log.
func Warn(msg string, keysAndValues ...interface{}) {
	writeJSON(levelWarn, msg, keysAndValues...)
}

// Error writes ERROR level JSON log.
func Error(msg string, keysAndValues ...interface{}) {
	writeJSON(levelError, msg, keysAndValues...)
}

// LogAccess writes access log JSON. body should be a JSON string or raw request body text.
func LogAccess(method, path string, status int, latency, body, requestID, sessionID string) {
	global.mu.RLock()
	target := global.outs.access
	serviceName := global.cfg.ServiceName
	global.mu.RUnlock()
	if target == nil {
		return
	}

	entry := AccessLogEntry{
		Time:        nowISO8601(),
		Level:       levelAccess,
		RequestID:   requestID,
		SessionID:   sessionID,
		Method:      method,
		Path:        path,
		Status:      status,
		Latency:     latency,
		Body:        body,
		ServiceName: serviceName,
	}

	writeLine(target, entry)
}

func writeJSON(level, msg string, keysAndValues ...interface{}) {
	global.mu.RLock()
	var target io.Writer
	switch level {
	case levelInfo:
		target = global.outs.info
	case levelWarn:
		target = global.outs.warn
	case levelError:
		target = global.outs.error
	default:
		target = global.outs.info
	}
	serviceName := global.cfg.ServiceName
	global.mu.RUnlock()

	if target == nil {
		return
	}

	fields := make(map[string]interface{})
	var sessionID string
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		if key == "log_session_id" {
			if v, ok := keysAndValues[i+1].(string); ok {
				sessionID = v
			}
			continue
		}
		fields[key] = normalizeValue(key, keysAndValues[i+1])
	}

	entry := LogEntry{
		Time:        nowISO8601(),
		Level:       level,
		Message:     msg,
		SessionID:   sessionID,
		ServiceName: serviceName,
	}
	if len(fields) > 0 {
		entry.Fields = fields
	}

	writeLine(target, entry)
}

func writeLine(target io.Writer, payload interface{}) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = target.Write(append(b, '\n'))
}

func nowISO8601() string {
	return time.Now().Format("2006-01-02T15:04:05.000Z07:00")
}
