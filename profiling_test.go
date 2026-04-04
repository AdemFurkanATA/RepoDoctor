package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildProfilingRequest_DefaultOff(t *testing.T) {
	root := t.TempDir()
	request, err := buildProfilingRequest(root, "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request.enabled() {
		t.Fatal("profiling should be disabled by default")
	}
}

func TestBuildProfilingRequest_WatchModeRejected(t *testing.T) {
	root := t.TempDir()
	_, err := buildProfilingRequest(root, "cpu.pprof", "", true)
	if err == nil {
		t.Fatal("expected watch mode + profiling to be rejected")
	}
}

func TestBuildProfilingRequest_ProfilePathRootBounded(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(root, "..", "escape", "cpu.pprof")

	_, err := buildProfilingRequest(root, outside, "", false)
	if err == nil {
		t.Fatal("expected path outside root to be rejected")
	}
}

func TestBuildProfilingRequest_ProfilePathSanitizedUnderRoot(t *testing.T) {
	root := t.TempDir()
	request, err := buildProfilingRequest(root, filepath.Join(".repodoctor", "profiles", "cpu.pprof"), "", false)
	if err != nil {
		t.Fatalf("unexpected sanitize error: %v", err)
	}

	if !pathWithinRoot(root, request.cpuProfilePath) {
		t.Fatalf("expected cpu profile path under root, got %s", request.cpuProfilePath)
	}
}

func TestStartProfiling_CreatesCpuAndMemArtifacts(t *testing.T) {
	root := t.TempDir()
	request, err := buildProfilingRequest(root, ".repodoctor/profiles/cpu.pprof", ".repodoctor/profiles/mem.pprof", false)
	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}

	session, err := startProfiling(request)
	if err != nil {
		t.Fatalf("startProfiling failed: %v", err)
	}

	for i := 0; i < 100000; i++ {
		_ = i * i
	}

	if err := session.Stop(); err != nil {
		t.Fatalf("profile stop failed: %v", err)
	}

	if info, statErr := os.Stat(request.cpuProfilePath); statErr != nil || info.Size() == 0 {
		t.Fatalf("expected non-empty cpu profile file at %s", request.cpuProfilePath)
	}
	if info, statErr := os.Stat(request.memProfilePath); statErr != nil || info.Size() == 0 {
		t.Fatalf("expected non-empty mem profile file at %s", request.memProfilePath)
	}
}

func TestComposeAnalyzeRequest_ParsesAndNormalizesProfileFlags(t *testing.T) {
	root := t.TempDir()
	req, err := composeAnalyzeRequest([]string{"-path", root, "-cpu-profile", "profiles/cpu.pprof", "-mem-profile", "profiles/mem.pprof"})
	if err != nil {
		t.Fatalf("unexpected compose error: %v", err)
	}

	if !req.profiling.enabled() {
		t.Fatal("expected profiling to be enabled")
	}
	if !pathWithinRoot(root, req.profiling.cpuProfilePath) {
		t.Fatalf("expected cpu profile under root, got %s", req.profiling.cpuProfilePath)
	}
	if !pathWithinRoot(root, req.profiling.memProfilePath) {
		t.Fatalf("expected mem profile under root, got %s", req.profiling.memProfilePath)
	}
}

func pathWithinRoot(root, target string) bool {
	resolvedRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return false
	}
	if next, resolveErr := filepath.EvalSymlinks(resolvedRoot); resolveErr == nil {
		resolvedRoot = filepath.Clean(next)
	}

	resolvedTarget, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return false
	}
	if next, resolveErr := filepath.EvalSymlinks(resolvedTarget); resolveErr == nil {
		resolvedTarget = filepath.Clean(next)
	}

	rel, relErr := filepath.Rel(resolvedRoot, resolvedTarget)
	if relErr != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
