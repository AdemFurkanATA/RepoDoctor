package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Logger defines a minimal structured logging contract for orchestration/CLI layers.
type Logger interface {
	Debug(msg string, attrs ...any)
	Info(msg string, attrs ...any)
	Warn(msg string, attrs ...any)
	Error(msg string, attrs ...any)
}

type slogLogger struct {
	delegate *slog.Logger
}

func (l *slogLogger) Debug(msg string, attrs ...any) { l.delegate.Debug(msg, attrs...) }
func (l *slogLogger) Info(msg string, attrs ...any)  { l.delegate.Info(msg, attrs...) }
func (l *slogLogger) Warn(msg string, attrs ...any)  { l.delegate.Warn(msg, attrs...) }
func (l *slogLogger) Error(msg string, attrs ...any) { l.delegate.Error(msg, attrs...) }

type noopLogger struct{}

func (l *noopLogger) Debug(string, ...any) {}
func (l *noopLogger) Info(string, ...any)  {}
func (l *noopLogger) Warn(string, ...any)  {}
func (l *noopLogger) Error(string, ...any) {}

// NewNoop returns a logger implementation that never emits output.
func NewNoop() Logger {
	return &noopLogger{}
}

// NewSlog wraps a slog.Logger behind the repository logging abstraction.
func NewSlog(delegate *slog.Logger) Logger {
	if delegate == nil {
		return NewNoop()
	}
	return &slogLogger{delegate: delegate}
}

// NewFromEnv returns a structured logger controlled by REPODOCTOR_LOG_LEVEL.
// Default is disabled (no-op), preserving deterministic output behavior.
func NewFromEnv() Logger {
	levelText := strings.TrimSpace(strings.ToLower(os.Getenv("REPODOCTOR_LOG_LEVEL")))
	level, enabled := parseLevel(levelText)
	if !enabled {
		return NewNoop()
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return NewSlog(slog.New(handler))
}

func parseLevel(value string) (slog.Level, bool) {
	switch value {
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "warn", "warning":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	default:
		return slog.LevelInfo, false
	}
}

// NewTestLogger returns a structured logger writing to the provided writer.
// Intended for deterministic unit tests.
func NewTestLogger(writer io.Writer, level slog.Level) Logger {
	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})
	return NewSlog(slog.New(handler))
}
