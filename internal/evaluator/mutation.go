package evaluator

import "fmt"

// mutationPenaltyFactor is the 0.8x multiplier applied when mutation
// coverage falls below the configured floor.
const mutationPenaltyFactor = 0.8

// ApplyMutationPenalty applies the 0.8x penalty to the verification pillar
// score when mutation coverage falls below the configured floor.
func ApplyMutationPenalty(baseScore, mutationCoverage, floor float64) float64 {
	if mutationCoverage >= floor {
		return baseScore
	}
	return baseScore * mutationPenaltyFactor
}

// MutationPenaltyMessage returns a human-readable explanation of the penalty.
func MutationPenaltyMessage(coverage, floor float64) string {
	return fmt.Sprintf("Mutation coverage %.0f%% is below the %.0f%% floor; "+
		"Verification pillar penalized by 20%%", coverage, floor)
}
