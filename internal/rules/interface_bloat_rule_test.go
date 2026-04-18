package rules

import "testing"

func TestInterfaceBloatRule_FlagsLargeInterface(t *testing.T) {
	rule := NewInterfaceBloatRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{
		Path:    "pkg/api/contracts.go",
		Content: "package api\ntype Gateway interface {\nA()\nB()\nC()\nD()\nE()\nF()\nG()\nH()\nI()\nJ()\nK()\n}\n",
	}}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 1 {
		t.Fatalf("expected one interface-bloat violation, got %d", len(violations))
	}
	if violations[0].RuleID != "rule.interface-bloat" {
		t.Fatalf("unexpected rule id: %q", violations[0].RuleID)
	}
}

func TestInterfaceBloatRule_UsesConfigurableThreshold(t *testing.T) {
	rule := NewInterfaceBloatRule()
	ctx := AnalysisContext{
		RepositoryFiles: []RepositoryFile{{
			Path:    "pkg/api/contracts.go",
			Content: "package api\ntype Gateway interface {\nA()\nB()\nC()\nD()\nE()\nF()\n}\n",
		}},
		Configuration: Configuration{"interfaceBloatMaxMethods": 6},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no violations when threshold=6 and methods=6, got %d", len(violations))
	}
}
