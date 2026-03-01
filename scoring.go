package main

import "fmt"

// StructuralScore represents the overall structural health score
type StructuralScore struct {
	TotalScore        float64
	CircularPenalty   float64
	LayerPenalty      float64
	SizePenalty       float64
	GodObjectPenalty  float64
	ViolationCount    int
	CircularCount     int
	LayerCount        int
	SizeCount         int
	GodObjectCount    int
	MaxScore          float64
}

// ScoringWeights defines penalty weights for different violation types
type ScoringWeights struct {
	CircularDependencyPenalty float64
	LayerViolationPenalty     float64
	SizeViolationPenalty      float64
	GodObjectPenalty          float64
}

// DefaultScoringWeights returns the default scoring weights
func DefaultScoringWeights() *ScoringWeights {
	return &ScoringWeights{
		CircularDependencyPenalty: 10.0, // High penalty for circular dependencies
		LayerViolationPenalty:     5.0,  // Medium penalty for layer violations
		SizeViolationPenalty:      3.0,  // Low penalty for size violations
		GodObjectPenalty:          5.0,  // Medium penalty for god objects
	}
}

// StructuralScorer calculates structural health scores
type StructuralScorer struct {
	weights        *ScoringWeights
	circularRule   *CircularDependencyRule
	layerRule      *LayerValidationRule
	sizeRule       *SizeRule
	godObjectRule  *GodObjectRule
	score          *StructuralScore
}

// NewStructuralScorer creates a new structural scorer with configuration
func NewStructuralScorer(graph Graph, config *Config, dirPath string) *StructuralScorer {
	if config == nil {
		config = (&ConfigLoader{}).getDefaultConfig()
	}

	// Create rules with config thresholds
	sizeRule := NewSizeRule()
	godObjectRule := NewGodObjectRule()
	
	// Apply config thresholds
	if config.Size != nil {
		sizeRule.MaxFileLines = config.Size.MaxFileLines
		sizeRule.MaxFunctionLines = config.Size.MaxFunctionLines
	}
	
	if config.GodObject != nil {
		godObjectRule.MaxFields = config.GodObject.MaxFields
		godObjectRule.MaxMethods = config.GodObject.MaxMethods
	}

	scorer := &StructuralScorer{
		weights:       DefaultScoringWeights(),
		circularRule:  NewCircularDependencyRule(graph),
		layerRule:     NewLayerValidationRule(graph),
		sizeRule:      sizeRule,
		godObjectRule: godObjectRule,
		score: &StructuralScore{
			MaxScore: 100.0,
		},
	}

	// Run rule checks if directory path provided
	if dirPath != "" {
		sizeRule.Check(dirPath)
		godObjectRule.Check(dirPath)
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
	s.score.CircularPenalty = float64(len(circularViolations)) * 10.0

	// Check layer violations
	s.layerRule.Check()
	layerViolations := s.layerRule.Violations()
	s.score.LayerCount = len(layerViolations)
	s.score.LayerPenalty = float64(len(layerViolations)) * 5.0

	// Check size violations
	sizeViolations := s.sizeRule.Violations()
	s.score.SizeCount = len(sizeViolations)
	s.score.SizePenalty = float64(len(sizeViolations)) * s.weights.SizeViolationPenalty

	// Check god object violations
	godObjectViolations := s.godObjectRule.Violations()
	s.score.GodObjectCount = len(godObjectViolations)
	s.score.GodObjectPenalty = float64(len(godObjectViolations)) * s.weights.GodObjectPenalty

	// Calculate total violations and penalty
	s.score.ViolationCount = s.score.CircularCount + s.score.LayerCount + s.score.SizeCount + s.score.GodObjectCount
	totalPenalty := s.score.CircularPenalty + s.score.LayerPenalty + s.score.SizePenalty + s.score.GodObjectPenalty

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
	explanation += fmt.Sprintf("Size Violations: %d violation(s) x %.1f penalty = %.1f\n",
		s.score.SizeCount, s.weights.SizeViolationPenalty, s.score.SizePenalty)
	explanation += fmt.Sprintf("God Objects: %d violation(s) x %.1f penalty = %.1f\n",
		s.score.GodObjectCount, s.weights.GodObjectPenalty, s.score.GodObjectPenalty)
	explanation += fmt.Sprintf("Total Penalty: %.1f\n", s.score.CircularPenalty+s.score.LayerPenalty+s.score.SizePenalty+s.score.GodObjectPenalty)
	explanation += fmt.Sprintf("Final Score: %.1f / %.1f\n", s.score.TotalScore, s.score.MaxScore)

	return explanation
}

// GetAllViolations returns all violations from all rules
func (s *StructuralScorer) GetAllViolations() struct {
	Circular  []CycleViolation
	Layer     []LayerViolation
	Size      []SizeViolation
	GodObject []GodObjectViolation
} {
	return struct {
		Circular  []CycleViolation
		Layer     []LayerViolation
		Size      []SizeViolation
		GodObject []GodObjectViolation
	}{
		Circular:  s.circularRule.Violations(),
		Layer:     s.layerRule.Violations(),
		Size:      s.sizeRule.Violations(),
		GodObject: s.godObjectRule.Violations(),
	}
}
