package languages

import "testing"

func TestGoAdapter_CapabilitiesAndNormalizeImport(t *testing.T) {
	adapter := NewGoAdapter()
	caps := adapter.Capabilities()

	if !caps.SupportsDependencyGraph || !caps.SupportsMetrics {
		t.Fatal("expected Go adapter to support graph and metrics")
	}

	if got := adapter.NormalizeImport("  github.com/foo/bar "); got != "github.com/foo/bar" {
		t.Fatalf("unexpected normalized import: %q", got)
	}
}

func TestPythonAdapter_CapabilitiesAndNormalizeImport(t *testing.T) {
	adapter := NewPythonAdapter()
	caps := adapter.Capabilities()

	if !caps.SupportsDependencyGraph || !caps.SupportsMetrics {
		t.Fatal("expected Python adapter to support graph and metrics")
	}

	if got := adapter.NormalizeImport(" requests.sessions "); got != "requests" {
		t.Fatalf("unexpected normalized import: %q", got)
	}
}
