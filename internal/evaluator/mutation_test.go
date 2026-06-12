package evaluator_test

import (
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
)

// --- Story D-4: Enforce Mutation Testing Coverage Requirements ---

// ACCEPTANCE CRITERIA 1:
// "I can configure a minimum mutation coverage threshold (default 80%)"
func TestMutationCoverage_DefaultThreshold(t *testing.T) {
	cfg := config.DefaultConfig()
	// Default mutation coverage floor should be 80%
	if cfg.Pillars.Verification.MutationCoverageFloor != 80 {
		t.Errorf("expected default mutation coverage floor of 80, got %f",
			cfg.Pillars.Verification.MutationCoverageFloor)
	}
}

// ACCEPTANCE CRITERIA 2:
// "When mutation coverage falls below the threshold, the Verification pillar score is reduced by 20%"
func TestMutationCoverage_BelowThresholdPenalty(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Pillars.Verification.MutationCoverageFloor = 80

	// With 70% mutation coverage (below 80% floor), verification should be penalized
	score := evaluator.ApplyMutationPenalty(35, 70, cfg.Pillars.Verification.MutationCoverageFloor)
	expected := 35 * 0.8 // 28

	if score != expected {
		t.Errorf("expected penalized score of %f, got %f", expected, score)
	}
}

// ACCEPTANCE CRITERIA 3:
// "The output explains that the penalty was applied due to insufficient mutation coverage"
func TestMutationCoverage_PenaltyMessage(t *testing.T) {
	msg := evaluator.MutationPenaltyMessage(70, 80)
	if msg == "" {
		t.Error("expected non-empty penalty message")
	}
}

// ACCEPTANCE CRITERIA 4:
// "When mutation coverage meets or exceeds the threshold, no penalty is applied"
func TestMutationCoverage_AboveThresholdNoPenalty(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Pillars.Verification.MutationCoverageFloor = 80

	// With 90% mutation coverage (above 80% floor), no penalty
	score := evaluator.ApplyMutationPenalty(35, 90, cfg.Pillars.Verification.MutationCoverageFloor)
	if score != 35 {
		t.Errorf("expected no penalty, score should be 35, got %f", score)
	}

	// Exactly at threshold
	score = evaluator.ApplyMutationPenalty(35, 80, cfg.Pillars.Verification.MutationCoverageFloor)
	if score != 35 {
		t.Errorf("expected no penalty at exact threshold, got %f", score)
	}
}

// Edge case: zero mutation coverage
func TestMutationCoverage_ZeroCoverage(t *testing.T) {
	score := evaluator.ApplyMutationPenalty(35, 0, 80)
	expected := 35 * 0.8
	if score != expected {
		t.Errorf("expected penalized score of %f for 0%% coverage, got %f", expected, score)
	}
}

// Edge case: 100% coverage
func TestMutationCoverage_PerfectCoverage(t *testing.T) {
	score := evaluator.ApplyMutationPenalty(35, 100, 80)
	if score != 35 {
		t.Errorf("expected no penalty for 100%% coverage, got %f", score)
	}
}

// Verify penalty is exactly 20% reduction
func TestMutationCoverage_PenaltyIs20Percent(t *testing.T) {
	base := 35.0
	penalized := evaluator.ApplyMutationPenalty(base, 50, 80)
	expectedReduction := base * 0.2
	actualReduction := base - penalized

	if actualReduction != expectedReduction {
		t.Errorf("expected 20%% reduction (%f), got reduction of %f",
			expectedReduction, actualReduction)
	}
}
