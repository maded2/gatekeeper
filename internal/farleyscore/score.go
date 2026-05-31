package farleyscore

import (
	"github.com/eddie/gatekeeper/internal/config"
	"github.com/eddie/gatekeeper/internal/evaluator"
)

// ScoreResult holds the computed Farley Score and its breakdown.
type ScoreResult struct {
	FarleyScore    float64
	QualityScore   float64
	CoverageScore  float64
	DeployScore    float64
	QualityWeight  float64
	CoverageWeight float64
	DeployWeight   float64
	Issues         []evaluator.Issue
	IsLargeDiff    bool
	LargeDiffWarn  string
}

// Compute calculates the Farley Score from evaluation results and configuration.
func Compute(result evaluator.EvaluationResult, cfg config.Configuration) ScoreResult {
	qw := cfg.PillarWeights.CodeQuality
	cw := cfg.PillarWeights.TestCoverage
	dw := cfg.PillarWeights.Deployability

	// Normalize weights to sum to 1.0
	total := qw + cw + dw
	if total > 0 {
		qw /= total
		cw /= total
		dw /= total
	}

	farley := qw*result.QualityScore + cw*result.CoverageScore + dw*result.DeployScore

	// Clamp to 1-10 range
	if farley < 1.0 {
		farley = 1.0
	}
	if farley > 10.0 {
		farley = 10.0
	}

	// Round to 1 decimal
	farley = roundTo1(farley)

	return ScoreResult{
		FarleyScore:    farley,
		QualityScore:   roundTo1(result.QualityScore),
		CoverageScore:  roundTo1(result.CoverageScore),
		DeployScore:    roundTo1(result.DeployScore),
		QualityWeight:  qw,
		CoverageWeight: cw,
		DeployWeight:   dw,
		Issues:         result.Issues,
		IsLargeDiff:    result.IsLargeDiff,
		LargeDiffWarn:  result.LargeDiffWarning,
	}
}

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
