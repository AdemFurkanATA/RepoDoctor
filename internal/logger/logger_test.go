package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNewFromEnv_DefaultIsNoop(t *testing.T) {
	t.Setenv("REPODOCTOR_LOG_LEVEL", "")

	log := NewFromEnv()
	log.Debug("debug", "k", "v")
	log.Info("info", "k", "v")
	log.Warn("warn", "k", "v")
	log.Error("error", "k", "v")
}

func TestNewFromEnv_InvalidValueFallsBackToNoop(t *testing.T) {
	t.Setenv("REPODOCTOR_LOG_LEVEL", "invalid")

	log := NewFromEnv()
	log.Info("ignored")
}

func TestNewTestLogger_EmitsStructuredOutput(t *testing.T) {
	var buf bytes.Buffer
	log := NewTestLogger(&buf, slog.LevelDebug)

	log.Info("pipeline.start", "adapter", "Go", "files", 3)

	out := buf.String()
	if !strings.Contains(out, "pipeline.start") {
		t.Fatalf("expected message in output, got %q", out)
	}
	if !strings.Contains(out, "adapter=Go") || !strings.Contains(out, "files=3") {
		t.Fatalf("expected key-value attrs in output, got %q", out)
	}
}

func TestParseLevel(t *testing.T) {
	cases := []struct {
		input   string
		enabled bool
	}{
		{input: "debug", enabled: true},
		{input: "info", enabled: true},
		{input: "warn", enabled: true},
		{input: "error", enabled: true},
		{input: "", enabled: false},
		{input: "off", enabled: false},
	}

	for _, tc := range cases {
		_, enabled := parseLevel(tc.input)
		if enabled != tc.enabled {
			t.Fatalf("parseLevel(%q) enabled expected %v, got %v", tc.input, tc.enabled, enabled)
		}
	}
}

func TestNewSlog_NilDelegateReturnsNoop(t *testing.T) {
	log := NewSlog(nil)
	log.Info("safe")

	if _, ok := log.(*noopLogger); !ok {
		t.Fatal("expected noop logger when delegate is nil")
	}
}

func TestNewFromEnv_RecognizesConfiguredLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			t.Setenv("REPODOCTOR_LOG_LEVEL", level)
			log := NewFromEnv()
			if _, ok := log.(*noopLogger); ok {
				t.Fatalf("expected configured level %q to enable logger", level)
			}
		})
	}
}
