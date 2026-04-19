package main

import (
	"fmt"
	"sort"

	"RepoDoctor/internal/model"
)

type HotspotEntry struct {
	File          string `json:"file"`
	RiskScore     int    `json:"riskScore"`
	Complexity    int    `json:"complexity"`
	CommitTouches int    `json:"commitTouches"`
	RecentTouches int    `json:"recentTouches"`
	Explanation   string `json:"explanation"`
}

// collectHotspotSummary computes deterministic hotspot rankings.
// Risk score is intentionally explainable: complexity × commitTouches.
func collectHotspotSummary(root string, churn model.GitChurnSummary) []HotspotEntry {
	complexities := collectGoFileComplexitySnapshot(root)
	if len(complexities) == 0 || len(churn.Files) == 0 {
		return nil
	}

	churnByPath := make(map[string]model.GitChurnFile, len(churn.Files))
	for _, file := range churn.Files {
		churnByPath[file.Path] = file
	}

	hotspots := make([]HotspotEntry, 0)
	for _, complexity := range complexities {
		churnFile, ok := churnByPath[complexity.Path]
		if !ok {
			continue
		}
		riskScore := complexity.TotalComplexity * churnFile.CommitTouches
		if riskScore <= 0 {
			continue
		}
		hotspots = append(hotspots, HotspotEntry{
			File:          complexity.Path,
			RiskScore:     riskScore,
			Complexity:    complexity.TotalComplexity,
			CommitTouches: churnFile.CommitTouches,
			RecentTouches: churnFile.RecentTouches,
			Explanation:   fmt.Sprintf("risk=%d (complexity:%d × churnTouches:%d)", riskScore, complexity.TotalComplexity, churnFile.CommitTouches),
		})
	}

	sort.SliceStable(hotspots, func(i, j int) bool {
		if hotspots[i].RiskScore != hotspots[j].RiskScore {
			return hotspots[i].RiskScore > hotspots[j].RiskScore
		}
		return hotspots[i].File < hotspots[j].File
	})

	return hotspots
}
