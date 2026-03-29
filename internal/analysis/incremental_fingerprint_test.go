package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestBuildGoFingerprintMap_OnlyGoFiles(t *testing.T) {
	repo := t.TempDir()
	goPath := filepath.Join(repo, "a.go")
	pyPath := filepath.Join(repo, "b.py")
	if err := os.WriteFile(goPath, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing go file: %v", err)
	}
	if err := os.WriteFile(pyPath, []byte("print('x')\n"), 0o644); err != nil {
		t.Fatalf("failed writing py file: %v", err)
	}

	state, err := BuildGoFingerprintMap(repo)
	if err != nil {
		t.Fatalf("BuildGoFingerprintMap failed: %v", err)
	}
	if len(state) != 1 {
		t.Fatalf("expected one go fingerprint entry, got %d", len(state))
	}
	if _, ok := state[filepath.Clean(goPath)]; !ok {
		t.Fatalf("expected fingerprint for go file path %s", goPath)
	}
}

func TestDiffFingerprintStates_DetectsChangedAddedRemoved(t *testing.T) {
	prev := map[string]string{"a.go": "h1", "b.go": "h2"}
	next := map[string]string{"a.go": "h1-new", "c.go": "h3"}

	changed, removed := DiffFingerprintStates(prev, next)
	if len(changed) != 2 {
		t.Fatalf("expected changed set size 2, got %d (%v)", len(changed), changed)
	}
	if len(removed) != 1 || removed[0] != "b.go" {
		t.Fatalf("expected removed [b.go], got %v", removed)
	}
}

func TestValidateFingerprintState(t *testing.T) {
	if err := ValidateFingerprintState(IncrementalFingerprintState{Files: nil}); err == nil {
		t.Fatal("expected nil files map validation failure")
	}
	if err := ValidateFingerprintState(IncrementalFingerprintState{Files: map[string]string{"": "h"}}); err == nil {
		t.Fatal("expected empty path validation failure")
	}
	if err := ValidateFingerprintState(IncrementalFingerprintState{Files: map[string]string{"a.go": "h"}}); err != nil {
		t.Fatalf("expected valid state, got %v", err)
	}
}

func TestBuildGoFingerprintMap_DeterministicAcrossRuns(t *testing.T) {
	repo := t.TempDir()
	for i := 0; i < 40; i++ {
		path := filepath.Join(repo, "pkg", "f"+strconv.Itoa(i)+".go")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed creating dir: %v", err)
		}
		content := []byte("package main\nfunc f" + strconv.Itoa(i) + "() int { return " + strconv.Itoa(i) + " }\n")
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("failed writing go file: %v", err)
		}
	}

	baseline, err := BuildGoFingerprintMap(repo)
	if err != nil {
		t.Fatalf("BuildGoFingerprintMap baseline failed: %v", err)
	}

	for i := 0; i < 20; i++ {
		next, nextErr := BuildGoFingerprintMap(repo)
		if nextErr != nil {
			t.Fatalf("BuildGoFingerprintMap failed at run %d: %v", i, nextErr)
		}
		if !reflect.DeepEqual(baseline, next) {
			t.Fatalf("fingerprint map must remain deterministic; baseline=%v next=%v", baseline, next)
		}
	}
}

func TestHashFileSHA256Streaming_MatchesReferenceHash(t *testing.T) {
	repo := t.TempDir()
	path := filepath.Join(repo, "streaming.go")
	content := []byte("package main\nfunc main(){ println(42) }\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed writing fixture file: %v", err)
	}

	actual, err := hashFileSHA256Streaming(path)
	if err != nil {
		t.Fatalf("hashFileSHA256Streaming failed: %v", err)
	}

	sum := sha256.Sum256(content)
	expected := hex.EncodeToString(sum[:])
	if actual != expected {
		t.Fatalf("streaming hash mismatch: expected %s got %s", expected, actual)
	}
}
