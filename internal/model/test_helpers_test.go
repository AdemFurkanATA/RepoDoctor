package model

import (
	"sort"
	"strings"
)

func sortedNodeIDs(nodes []*Node) []string {
	ids := make([]string, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.ID)
	}
	sort.Strings(ids)
	return ids
}

func cycleSignatures(cycles [][]string) []string {
	if len(cycles) == 0 {
		return nil
	}
	signatures := make([]string, 0, len(cycles))
	for _, cycle := range cycles {
		copied := append([]string{}, cycle...)
		sort.Strings(copied)
		signatures = append(signatures, strings.Join(copied, "->"))
	}
	sort.Strings(signatures)
	return signatures
}
