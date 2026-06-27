# ihy-loglib

Shared JSON logging library for IHY Go services. Writes to local log files and optional stdout. Fluent Bit collects log files and ships them to OpenSearch.

## Features

- JSON logs for `INFO`, `WARN`, `ERROR`, and `ACCESS`
- Dedicated files per level: `info.log`, `warning.log`, `error.log`, `access.log`
- Optional mirrored output to stdout
- Sensitive field masking in structured logs
- **Log session ID** scoped to a request/task (auto-init or client-provided)
- Fiber middleware with request body masking
- GORM logger adapter

## Install

```bash
go get github.com/Fortuna-Lab/ihy-loglib
```

For GORM support:

```bash
go get github.com/Fortuna-Lab/ihy-loglib/gormlogger
```

## Quick Start

```go
package main

import (
	"log"

	"github.com/Fortuna-Lab/ihy-loglib/logger"
)

func main() {
	err := logger.Init(logger.Config{
		Dir:            "logs",
		MirrorToStdout: true,
		ServiceName:    "identity",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close()

	logger.Info("Server starting", "port", "8080", "env", "dev")
	logger.Warn("High latency detected", "latency_ms", 420)
	logger.Error("Failed to connect DB", "error", "timeout")
}
```

## Environment Variables

Use `InitFromEnv()` or `ConfigFromEnv()` for 12-factor configuration:

| Variable | Description |
|----------|-------------|
| `LOG_DIR` / `IDENTITY_LOGS_DIR` | Log file directory (default: `logs`) |
| `LOG_MIRROR_STDOUT` | Mirror to stdout (default: `true`) |
| `LOG_SERVICE_NAME` / `SERVICE_NAME` | Service name included in JSON logs |

```go
if err := logger.InitFromEnv(); err != nil {
	log.Fatal(err)
}
defer logger.Close()
```

## Log Session ID

Session ID ties all logs in one request/task together in OpenSearch.

### HTTP (Fiber)

Register `SessionID` middleware **before** handlers and API logger:

```go
app.Use(middleware.SessionID(middleware.SessionConfig{}))
app.Use(middleware.APILogger(middleware.FiberConfig{}))
```

- Client sends `X-Session-ID` → used for the whole request
- Header missing → lib generates UUID automatically
- Same ID is echoed in the response header

In handlers, use context-aware logging (no need to pass `log_session_id` every time):

```go
func handler(c *fiber.Ctx) error {
    logger.InfoCtx(middleware.Ctx(c), "AdminGetVideos start", "actor", actor)
    defer logger.InfoCtx(middleware.Ctx(c), "AdminGetVideos end", "actor", actor)
    // ...
}
```

Pass session to GORM via request context:

```go
db.WithContext(middleware.Ctx(c)).Find(&videos)
```

### Background worker / task

```go
logger.RunWithSession(context.Background(), incomingSessionID, func(ctx context.Context) {
    logger.InfoCtx(ctx, "batch received", "count", 10)
    logger.InfoCtx(ctx, "batch done")
})
```

### Plain Info/Warn/Error (no context)

Each goroutine auto-gets a `log_session_id` on the first log call and reuses it:

```go
logger.Info("worker loop started", "worker_id", workerID, "batch_size", batchSize)
// → {"message":"worker loop started","log_session_id":"...","fields":{"worker_id":1,...}}

logger.Info("batch received", "log_session_id", batchSessionID, "count", 10)
// → uses batchSessionID for this line only

logger.Warn("redis BLPOP error", "worker_id", workerID, "error", err)
// → continues using the worker goroutine session
```

Optional explicit bind at goroutine start:

```go
unbind := logger.BindSession("") // or pass known ID
defer unbind()
```

```go
import (
	"github.com/Fortuna-Lab/ihy-loglib/middleware"
)

app.Use(middleware.SessionID(middleware.SessionConfig{}))
app.Use(middleware.APILogger(middleware.FiberConfig{
	SensitiveFields:   []string{"password", "client_secret", "token"},
	RequestIDLocalKey: "request_id",
	SessionIDLocalKey: "log_session_id",
}))
```

## GORM Logger

```go
import (
	"github.com/Fortuna-Lab/ihy-loglib/gormlogger"
	"gorm.io/gorm"
)

db, err := gorm.Open(dialector, &gorm.Config{
	Logger: gormlogger.New(),
})
```

## Build & Test

```bash
make tidy
make test
make build
```

## API

- `logger.Init(cfg logger.Config) error`
- `logger.InitFromEnv() error`
- `logger.ConfigFromEnv() Config`
- `logger.Close() error`
- `logger.BeginSession(ctx, sessionID string) (context.Context, string)`
- `logger.SessionIDFromContext(ctx) string`
- `logger.RunWithSession(ctx, sessionID string, fn func(context.Context))`
- `logger.Info / Warn / Error(msg, keysAndValues...)`
- `logger.InfoCtx / WarnCtx / ErrorCtx(ctx, msg, keysAndValues...)`
- `logger.LogAccess(method, path, status, latency, body, requestID, sessionID string)`
- `middleware.SessionID(cfg SessionConfig) fiber.Handler`
- `middleware.Ctx(c *fiber.Ctx) context.Context`
- `middleware.APILogger(cfg FiberConfig) fiber.Handler`
- `gormlogger.New() gormlogger.Interface`

Compatibility helpers: `InitLogger()`, `CloseAll()`.
