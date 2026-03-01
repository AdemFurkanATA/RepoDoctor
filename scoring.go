package main

import "fmt"

// StructuralScore represents the overall structural health score
type StructuralScore struct {
	TotalScore        float64
	CircularPenalty   float64
	LayerPenalty      float64
	GodObjectPenalty  float64
	ViolationCount    int
	CircularCount     int
	LayerCount        int
	GodObjectCount    int
	MaxScore          float64
}

// ScoringWeights defines penalty weights for different violation types
type ScoringWeights struct {
	CircularDependencyPenalty float64
	LayerViolationPenalty     float64
	GodObjectPenalty          float64
}

// DefaultScoringWeights returns the default scoring weights
func DefaultScoringWeights() *ScoringWeights {
	return &ScoringWeights{
		CircularDependencyPenalty: 10.0, // High penalty for circular dependencies
		LayerViolationPenalty:     5.0,  // Medium penalty for layer violations
		GodObjectPenalty:          5.0,  // Medium penalty for god objects
	}
}

// StructuralScorer calculates structural health scores
type StructuralScorer struct {
	weights        *ScoringWeights
	circularRule   *CircularDependencyRule
	layerRule      *LayerValidationRule
	godObjectRule  *GodObjectRule
	score          *StructuralScore
}

// NewStructuralScorer creates a new structural scorer
func NewStructuralScorer(graph Graph, weights *ScoringWeights, dirPath string) *StructuralScorer {
	if weights == nil {
		weights = DefaultScoringWeights()
	}

	scorer := &StructuralScorer{
		weights:       weights,
		circularRule:  NewCircularDependencyRule(graph),
		layerRule:     NewLayerValidationRule(graph),
		godObjectRule: NewGodObjectRule(),
		score: &StructuralScore{
			MaxScore: 100.0,
		},
	}

	// Run god object rule check if directory path provided
	if dirPath != "" {
		scorer.godObjectRule.Check(dirPath)
	}

	return scorer
}

// CalculateScore computes the structural health score
func (s *StructuralScorer) CalculateScore() *StructuralScore {
	s.score = &StructuralScore{
		MaxScore: 100.0,
	}

	// Check circular dependencies
	s.circularRule.Check()
	circularViolations := s.circularRule.Violations()
	s.score.CircularCount = len(circularViolations)
	s.score.CircularPenalty = float64(len(circularViolations)) * s.weights.CircularDependencyPenalty

	// Check layer violations
	s.layerRule.Check()
	layerViolations := s.layerRule.Violations()
	s.score.LayerCount = len(layerViolations)
	s.score.LayerPenalty = float64(len(layerViolations)) * s.weights.LayerViolationPenalty

	// Check god object violations
	godObjectViolations := s.godObjectRule.Violations()
	s.score.GodObjectCount = len(godObjectViolations)
	s.score.GodObjectPenalty = float64(len(godObjectViolations)) * s.weights.GodObjectPenalty

	// Calculate total violations and penalty
	s.score.ViolationCount = s.score.CircularCount + s.score.LayerCount + s.score.GodObjectCount
	totalPenalty := s.score.CircularPenalty + s.score.LayerPenalty + s.score.GodObjectPenalty

	// Calculate final score (deterministic, no duplicate penalty)
	s.score.TotalScore = s.score.MaxScore - totalPenalty
	if s.score.TotalScore < 0 {
		s.score.TotalScore = 0
	}

	return s.score
}

// GetCircularRule returns the circular dependency rule checker
func (s *StructuralScorer) GetCircularRule() *CircularDependencyRule {
	return s.circularRule
}

// GetLayerRule returns the layer validation rule checker
func (s *StructuralScorer) GetLayerRule() *LayerValidationRule {
	return s.layerRule
}

// HasCriticalViolations returns true if there are critical violations
func (s *StructuralScorer) HasCriticalViolations() bool {
	return s.circularRule.Check()
}

// GetScoreExplanation returns a detailed explanation of the score calculation
func (s *StructuralScorer) GetScoreExplanation() string {
	explanation := "Structural Score Breakdown:\n"
	explanation += "=========================\n"
	explanation += fmt.Sprintf("Base Score: %.1f\n", s.score.MaxScore)
	explanation += fmt.Sprintf("Circular Dependencies: %d violation(s) x %.1f penalty = %.1f\n",
		s.score.CircularCount, s.weights.CircularDependencyPenalty, s.score.CircularPenalty)
	explanation += fmt.Sprintf("Layer Violations: %d violation(s) x %.1f penalty = %.1f\n",
		s.score.LayerCount, s.weights.LayerViolationPenalty, s.score.LayerPenalty)
	explanation += fmt.Sprintf("God Objects: %d violation(s) x %.1f penalty = %.1f\n",
		s.score.GodObjectCount, s.weights.GodObjectPenalty, s.score.GodObjectPenalty)
	explanation += fmt.Sprintf("Total Penalty: %.1f\n", s.score.CircularPenalty+s.score.LayerPenalty+s.score.GodObjectPenalty)
	explanation += fmt.Sprintf("Final Score: %.1f / %.1f\n", s.score.TotalScore, s.score.MaxScore)

	return explanation
}

// GetAllViolations returns all violations from all rules
func (s *StructuralScorer) GetAllViolations() struct {
	Circular  []CycleViolation
	Layer     []LayerViolation
	GodObject []GodObjectViolation
} {
	return struct {
		Circular  []CycleViolation
		Layer     []LayerViolation
		GodObject []GodObjectViolation
	}{
		Circular:  s.circularRule.Violations(),
		Layer:     s.layerRule.Violations(),
		GodObject: s.godObjectRule.Violations(),
	}
}
