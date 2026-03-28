package main

import "testing"

func TestFormatCycle_UsesDecimalIndex(t *testing.T) {
	formatted := formatCycle(12, []string{"a", "b"})
	if formatted[:4] != "[12]" {
		t.Fatalf("expected decimal index prefix [12], got %q", formatted)
	}
}
