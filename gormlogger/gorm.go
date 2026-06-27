package gormlogger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Fortuna-Lab/ihy-loglib/logger"
	gormlogger "gorm.io/gorm/logger"
)

// Logger implements gorm.io/gorm/logger.Interface using ihy-loglib.
type Logger struct {
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

// New returns a GORM logger with sensible defaults.
func New() gormlogger.Interface {
	return &Logger{
		LogLevel:                  gormlogger.Info,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: false,
	}
}

func (l *Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	next := *l
	next.LogLevel = level
	return &next
}

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		logger.Info("GORM info", "source", "gorm", "message", fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		logger.Warn("GORM warning", "source", "gorm", "message", fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		logger.Error("GORM error", "source", "gorm", "message", fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	elapsedMS := float64(elapsed.Nanoseconds()) / 1e6

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		logger.Error("GORM SQL error", "source", "gorm", "elapsed_ms", elapsedMS, "rows", rows, "sql", sql, "error", err)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		logger.Warn("GORM slow SQL", "source", "gorm", "elapsed_ms", elapsedMS, "rows", rows, "sql", sql, "slow_threshold", l.SlowThreshold.String())
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		logger.Info("GORM SQL", "source", "gorm", "elapsed_ms", elapsedMS, "rows", rows, "sql", sql)
	}
}
