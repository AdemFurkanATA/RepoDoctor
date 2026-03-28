package domain

import "testing"

func TestParseArchitectureProfile(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  ArchitectureProfile
		wantError bool
	}{
		{name: "default empty -> layered", input: "", expected: ArchitectureProfileLayered},
		{name: "clean", input: "clean", expected: ArchitectureProfileClean},
		{name: "layered", input: "layered", expected: ArchitectureProfileLayered},
		{name: "modular monolith", input: "modular-monolith", expected: ArchitectureProfileModularMonolith},
		{name: "invalid", input: "hexagonal", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArchitectureProfile(tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}
