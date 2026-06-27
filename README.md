# ihy-loglib

Shared JSON logging library for IHY Go services. Writes to local log files and optional stdout. Fluent Bit collects log files and ships them to OpenSearch.

## Features

- JSON logs for `INFO`, `WARN`, `ERROR`, and `ACCESS`
- Dedicated files per level: `info.log`, `warning.log`, `error.log`, `access.log`
- Optional mirrored output to stdout
- Sensitive field masking in structured logs
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

## Fiber Middleware

```go
import (
	"github.com/Fortuna-Lab/ihy-loglib/middleware"
)

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
- `logger.Info / Warn / Error(msg, keysAndValues...)`
- `logger.LogAccess(method, path, status, latency, body, requestID, sessionID string)`
- `middleware.APILogger(cfg FiberConfig) fiber.Handler`
- `gormlogger.New() gormlogger.Interface`

Compatibility helpers: `InitLogger()`, `CloseAll()`.
