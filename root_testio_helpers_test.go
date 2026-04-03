package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed creating stdout pipe: %v", err)
	}
	os.Stdout = writer

	defer func() {
		os.Stdout = original
	}()

	fn()
	_ = writer.Close()

	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed reading stdout capture: %v", err)
	}
	return string(out)
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed creating stderr pipe: %v", err)
	}
	os.Stderr = writer

	defer func() {
		os.Stderr = original
	}()

	fn()
	_ = writer.Close()

	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed reading stderr capture: %v", err)
	}
	return string(out)
}

func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
