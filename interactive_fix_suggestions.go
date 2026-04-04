package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const interactiveFixSuggestionsEnv = "REPODOCTOR_INTERACTIVE_FIX_SUGGESTIONS"

type InteractiveFixSuggestionAdvisor struct {
	Enabled bool
	io      *interactiveIO
}

func NewInteractiveFixSuggestionAdvisorFromEnv(io *interactiveIO) *InteractiveFixSuggestionAdvisor {
	return &InteractiveFixSuggestionAdvisor{
		Enabled: parseBoolEnv(interactiveFixSuggestionsEnv),
		io:      io,
	}
}

func (a *InteractiveFixSuggestionAdvisor) MaybePrint(report *StructuralReport) {
	if a == nil || !a.Enabled || report == nil || !report.HasViolations {
		return
	}

	suggestions := buildInteractiveFixSuggestions(report)
	if len(suggestions) == 0 {
		return
	}

	fmt.Printf("\nInteractive fix suggestions are enabled via %s.\n", interactiveFixSuggestionsEnv)
	fmt.Println("Suggestions are advisory only; no files are changed automatically.")
	if !a.io.confirm("Show remediation suggestions") {
		return
	}

	fmt.Println()
	for idx, item := range suggestions {
		fmt.Printf("  %d. %s\n", idx+1, item)
	}
	fmt.Println()
}

func buildInteractiveFixSuggestions(report *StructuralReport) []string {
	items := make([]string, 0, 4)

	if len(report.Circular) > 0 {
		items = append(items, firstHintOrFallback(report.Circular[0].Hint,
			"Break dependency cycles by extracting shared contracts into a lower layer and inverting dependencies."))
	}
	if len(report.Layer) > 0 {
		items = append(items, firstHintOrFallback(report.Layer[0].Hint,
			"Move upward imports behind an interface or toward a lower layer allowed by your profile."))
	}
	if len(report.Size) > 0 {
		items = append(items, firstHintOrFallback(report.Size[0].Hint,
			"Split oversized files/functions into smaller focused units and keep one responsibility per unit."))
	}
	if len(report.GodObject) > 0 {
		items = append(items, firstHintOrFallback(report.GodObject[0].Hint,
			"Extract cohesive collaborators and reduce broad structs into dedicated components."))
	}

	return items
}

func firstHintOrFallback(hint, fallback string) string {
	trimmed := strings.TrimSpace(hint)
	if trimmed != "" {
		return trimmed
	}
	return fallback
}

func parseBoolEnv(key string) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return false
	}
	if value, err := strconv.ParseBool(raw); err == nil {
		return value
	}
	return raw == "1" || raw == "yes" || raw == "on" || raw == "enabled"
}
