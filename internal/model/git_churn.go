package model

// GitChurnSummary contains deterministic churn metrics extracted from git history.
type GitChurnSummary struct {
	TotalCommits        int            `json:"totalCommits"`
	RecentWindowCommits int            `json:"recentWindowCommits"`
	DistinctAuthors     int            `json:"distinctAuthors"`
	Files               []GitChurnFile `json:"files"`
}

// GitChurnFile captures per-file churn intensity.
type GitChurnFile struct {
	Path          string `json:"path"`
	CommitTouches int    `json:"commitTouches"`
	RecentTouches int    `json:"recentTouches"`
	AddedLines    int    `json:"addedLines"`
	DeletedLines  int    `json:"deletedLines"`
	UniqueAuthors int    `json:"uniqueAuthors"`
}
