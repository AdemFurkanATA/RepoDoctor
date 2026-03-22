package domain

import (
	"fmt"
	"strings"
)

type ArchitectureProfile string

const (
	ArchitectureProfileClean           ArchitectureProfile = "clean"
	ArchitectureProfileLayered         ArchitectureProfile = "layered"
	ArchitectureProfileModularMonolith ArchitectureProfile = "modular-monolith"
)

func ParseArchitectureProfile(raw string) (ArchitectureProfile, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return ArchitectureProfileLayered, nil
	}
	profile := ArchitectureProfile(normalized)
	if !profile.Valid() {
		return "", fmt.Errorf("architecture.profile must be one of: clean, layered, modular-monolith")
	}
	return profile, nil
}

func (p ArchitectureProfile) Valid() bool {
	switch p {
	case ArchitectureProfileClean, ArchitectureProfileLayered, ArchitectureProfileModularMonolith:
		return true
	default:
		return false
	}
}
