package main

import (
	"os"
	"path/filepath"
	"testing"

	"RepoDoctor/internal/model"
)

func TestCollectHotspotSummary_DeterministicRiskOrdering(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package demo\nfunc A(){ if true {} }\n"), 0o644); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte("package demo\nfunc B(){ if true {} }\n"), 0o644); err != nil {
		t.Fatalf("write b.go: %v", err)
	}

	churn := model.GitChurnSummary{Files: []model.GitChurnFile{
		{Path: "a.go", CommitTouches: 10, RecentTouches: 3},
		{Path: "b.go", CommitTouches: 10, RecentTouches: 2},
	}}

	computed := collectHotspotSummary(root, churn)
	if len(computed) != 2 {
		t.Fatalf("expected two hotspot entries, got %d (%v)", len(computed), computed)
	}
	if computed[0].File != "a.go" || computed[1].File != "b.go" {
		t.Fatalf("expected deterministic file asc tie-break ordering, got %v", computed)
	}
	if computed[0].RiskScore <= 0 || computed[0].Explanation == "" {
		t.Fatalf("expected explainable positive risk score, got %+v", computed[0])
	}
}
