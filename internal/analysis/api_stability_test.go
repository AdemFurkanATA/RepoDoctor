package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"RepoDoctor/internal/model"
)

func TestComputeAPIBreakingChanges_DetectsRemovedAndSignatureChangesDeterministically(t *testing.T) {
	repo := t.TempDir()
	goFile := filepath.Join(repo, "service.go")
	if err := os.WriteFile(goFile, []byte("package demo\n\nfunc Ping(v int) int { return v }\n"), 0o644); err != nil {
		t.Fatalf("write go file: %v", err)
	}

	baseline := model.APISnapshot{
		SchemaVersion: "v1",
		Language:      "Go",
		Symbols: []model.APISymbol{
			{ID: "demo|function|Legacy", Package: "demo", Name: "Legacy", Kind: "function", Signature: "func()", File: "legacy.go", Line: 3},
			{ID: "demo|function|Ping", Package: "demo", Name: "Ping", Kind: "function", Signature: "func(string) int", File: "service.go", Line: 3},
		},
	}
	baselinePath := filepath.Join(repo, "baseline.json")
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		t.Fatalf("marshal baseline: %v", err)
	}
	if err := os.WriteFile(baselinePath, data, 0o644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	t.Setenv(apiStabilityBaselineEnv, baselinePath)
	changes, warnings := ComputeAPIBreakingChanges(repo, "Go", []string{goFile})
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 breaking changes, got %d (%v)", len(changes), changes)
	}
	if changes[0].ChangeType != apiBreakTypeRemoved {
		t.Fatalf("expected first change to be removed, got %s", changes[0].ChangeType)
	}
	if changes[1].ChangeType != apiBreakTypeSignatureChanged {
		t.Fatalf("expected second change to be signature-changed, got %s", changes[1].ChangeType)
	}
}

func TestComputeAPIBreakingChanges_MissingBaselineIsFailSoft(t *testing.T) {
	repo := t.TempDir()
	t.Setenv(apiStabilityBaselineEnv, filepath.Join(repo, "does-not-exist.json"))

	changes, warnings := ComputeAPIBreakingChanges(repo, "Go", nil)
	if len(changes) != 0 {
		t.Fatalf("expected no changes without baseline, got %v", changes)
	}
	if len(warnings) == 0 {
		t.Fatal("expected fail-soft warning when baseline is missing")
	}
}
