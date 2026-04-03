package main

import (
	"os"
	"strings"
	"testing"
)

func TestColorFormatter_BasicFormatting(t *testing.T) {
	formatter := NewColorFormatter(false)

	if got := formatter.Color("x", ColorRed); got != "x" {
		t.Fatalf("expected disabled formatter to return plain text, got %q", got)
	}
	if got := formatter.Success("ok"); got != "ok" {
		t.Fatalf("expected success to be plain text when disabled, got %q", got)
	}
	if got := formatter.Bold("b"); got != "b" {
		t.Fatalf("expected bold to be plain text when disabled, got %q", got)
	}
}

func TestGlobalColorHelpers_WithInitializedFormatter(t *testing.T) {
	InitColorFormatter(false)
	if GetColorFormatter() == nil {
		t.Fatal("expected initialized global color formatter")
	}

	if got := ColorInfo("info"); got != "info" {
		t.Fatalf("expected ColorInfo plain text, got %q", got)
	}
	if got := ColorWarn("warn"); got != "warn" {
		t.Fatalf("expected ColorWarn plain text, got %q", got)
	}
	if got := ColorError("err"); got != "err" {
		t.Fatalf("expected ColorError plain text, got %q", got)
	}
	if got := ColorSuccess("ok"); got != "ok" {
		t.Fatalf("expected ColorSuccess plain text, got %q", got)
	}
}

func TestColorPrintfAndFprintf(t *testing.T) {
	InitColorFormatter(false)

	stdout := captureStdout(t, func() {
		ColorPrintf(ColorBlue, "hello %s", "world")
	})
	if !strings.Contains(stdout, "hello world") {
		t.Fatalf("expected ColorPrintf output, got %q", stdout)
	}

	stderr := captureStderr(t, func() {
		ColorFprintf(os.Stderr, ColorGreen, "value=%d", 42)
	})
	if !strings.Contains(stderr, "value=42") {
		t.Fatalf("expected ColorFprintf output, got %q", stderr)
	}
}
