package farleyscore_test

import (
	"testing"

	"github.com/eddie/gatekeeper/internal/config"
	"github.com/eddie/gatekeeper/internal/evaluator"
	"github.com/eddie/gatekeeper/internal/farleyscore"
)

// =============================================================================
// Story 3.1: Farley Score Computation
// Acceptance Tests
// =============================================================================

func TestFarleyScore_ComputesNumericScore(t *testing.T) {
	// Given clean evaluation results
	result := evaluator.EvaluationResult{
		QualityScore:  10.0,
		CoverageScore: 10.0,
		DeployScore:   10.0,
	}
	cfg := config.Default()

	// When the Farley Score is computed
	score := farleyscore.Compute(result, cfg)

	// Then a numeric score on 1-10 scale is returned
	if score.FarleyScore < 1.0 || score.FarleyScore > 10.0 {
		t.Errorf("expected score between 1-10, got %f", score.FarleyScore)
	}
}

func TestFarleyScore_PureQualityScore(t *testing.T) {
	// Given perfect quality but zero coverage and deploy
	result := evaluator.EvaluationResult{
		QualityScore:  10.0,
		CoverageScore: 1.0,
		DeployScore:   1.0,
	}
	cfg := config.Default()

	// When the Farley Score is computed
	score := farleyscore.Compute(result, cfg)

	// Then the score reflects the weighted average
	if score.FarleyScore > 5.0 {
		t.Errorf("expected low score for poor coverage/deploy, got %f", score.FarleyScore)
	}
	if score.QualityScore != 10.0 {
		t.Errorf("expected quality score 10.0, got %f", score.QualityScore)
	}
}

func TestFarleyScore_AggregatesAllPillars(t *testing.T) {
	// Given evaluation results from all three pillars
	result := evaluator.EvaluationResult{
		QualityScore:  8.0,
		CoverageScore: 6.0,
		DeployScore:   9.0,
	}
	cfg := config.Default()

	// When the Farley Score is computed
	score := farleyscore.Compute(result, cfg)

	// Then all pillar contributions are visible
	if score.QualityScore != 8.0 {
		t.Errorf("expected quality 8.0, got %f", score.QualityScore)
	}
	if score.CoverageScore != 6.0 {
		t.Errorf("expected coverage 6.0, got %f", score.CoverageScore)
	}
	if score.DeployScore != 9.0 {
		t.Errorf("expected deploy 9.0, got %f", score.DeployScore)
	}
}

func TestFarleyScore_ConsistentGivenSameInput(t *testing.T) {
	// Given the same evaluation result and config
	result := evaluator.EvaluationResult{
		QualityScore:  7.5,
		CoverageScore: 5.0,
		DeployScore:   8.0,
	}
	cfg := config.Default()

	// When the score is computed twice
	score1 := farleyscore.Compute(result, cfg)
	score2 := farleyscore.Compute(result, cfg)

	// Then the scores are identical
	if score1.FarleyScore != score2.FarleyScore {
		t.Errorf("expected consistent score, got %f and %f", score1.FarleyScore, score2.FarleyScore)
	}
}

func TestFarleyScore_RespectsConfigurableWeights(t *testing.T) {
	// Given a config with heavy quality weighting
	result := evaluator.EvaluationResult{
		QualityScore:  10.0,
		CoverageScore: 1.0,
		DeployScore:   1.0,
	}
	cfg := config.Configuration{
		Threshold: 6.0,
		PillarWeights: config.PillarWeights{
			CodeQuality:   0.8,
			TestCoverage:  0.1,
			Deployability: 0.1,
		},
	}

	// When the Farley Score is computed
	score := farleyscore.Compute(result, cfg)

	// Then quality has the most influence
	if score.FarleyScore < 7.0 {
		t.Errorf("expected high score with heavy quality weight, got %f", score.FarleyScore)
	}
}

func TestFarleyScore_PerfectScoreIsTen(t *testing.T) {
	// Given perfect scores across all pillars
	result := evaluator.EvaluationResult{
		QualityScore:  10.0,
		CoverageScore: 10.0,
		DeployScore:   10.0,
	}
	cfg := config.Default()

	// When the Farley Score is computed
	score := farleyscore.Compute(result, cfg)

	// Then the score is 10
	if score.FarleyScore != 10.0 {
		t.Errorf("expected perfect score 10.0, got %f", score.FarleyScore)
	}
}

func TestFarleyScore_WorstScoreIsOne(t *testing.T) {
	// Given worst scores across all pillars
	result := evaluator.EvaluationResult{
		QualityScore:  1.0,
		CoverageScore: 1.0,
		DeployScore:   1.0,
	}
	cfg := config.Default()

	// When the Farley Score is computed
	score := farleyscore.Compute(result, cfg)

	// Then the score is 1
	if score.FarleyScore != 1.0 {
		t.Errorf("expected worst score 1.0, got %f", score.FarleyScore)
	}
}

func TestFarleyScore_IssuesIncluded(t *testing.T) {
	// Given evaluation results with issues
	result := evaluator.EvaluationResult{
		QualityScore: 7.0,
		CoverageScore: 6.0,
		DeployScore: 8.0,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Warning, Category: evaluator.CodeQuality, Title: "test issue"},
		},
	}
	cfg := config.Default()

	// When the Farley Score is computed
	score := farleyscore.Compute(result, cfg)

	// Then issues are included in the result
	if len(score.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(score.Issues))
	}
}

func TestFarleyScore_WeightsNormalized(t *testing.T) {
	// Given weights that don't sum to 1.0
	cfg := config.Configuration{
		Threshold: 6.0,
		PillarWeights: config.PillarWeights{
			CodeQuality:   4,
			TestCoverage:  3,
			Deployability: 3,
		},
	}
	result := evaluator.EvaluationResult{
		QualityScore:  10.0,
		CoverageScore: 10.0,
		DeployScore:   10.0,
	}

	// When the Farley Score is computed
	score := farleyscore.Compute(result, cfg)

	// Then weights are normalized and score is still 10 for perfect input
	if score.FarleyScore != 10.0 {
		t.Errorf("expected 10.0 for perfect scores, got %f", score.FarleyScore)
	}
}

// =============================================================================
// Story 3.2: Score Consistency
// Acceptance Tests
// =============================================================================

func TestScoreConsistency_SameInputSameScore(t *testing.T) {
	// Given the same input
	result := evaluator.EvaluationResult{
		QualityScore:  7.2,
		CoverageScore: 5.8,
		DeployScore:   9.1,
	}
	cfg := config.Default()

	// When computed 10 times
	var scores []float64
	for i := 0; i < 10; i++ {
		s := farleyscore.Compute(result, cfg)
		scores = append(scores, s.FarleyScore)
	}

	// Then all scores are identical
	for i := 1; i < len(scores); i++ {
		if scores[i] != scores[0] {
			t.Errorf("score %d (%f) differs from score 0 (%f)", i, scores[i], scores[0])
		}
	}
}

func TestScoreConsistency_DeterministicScoring(t *testing.T) {
	// Given a fixed evaluation result
	result := evaluator.EvaluationResult{
		QualityScore:  6.0,
		CoverageScore: 8.0,
		DeployScore:   7.0,
	}
	cfg := config.Default()

	// When the score is computed
	score := farleyscore.Compute(result, cfg)

	// Then the score is deterministic (same value every time)
	for i := 0; i < 5; i++ {
		s := farleyscore.Compute(result, cfg)
		if s.FarleyScore != score.FarleyScore {
			t.Errorf("non-deterministic score on iteration %d", i)
		}
	}
}

// =============================================================================
// Story 3.3: Score Breakdown Transparency
// Acceptance Tests
// =============================================================================

func TestScoreBreakdown_ShowsFinalScore(t *testing.T) {
	// Given evaluation results
	result := evaluator.EvaluationResult{
		QualityScore:  8.0,
		CoverageScore: 6.0,
		DeployScore:   7.0,
	}
	cfg := config.Default()

	// When the score is computed
	score := farleyscore.Compute(result, cfg)

	// Then the final score is present
	if score.FarleyScore == 0 {
		t.Error("expected non-zero final score")
	}
}

func TestScoreBreakdown_ShowsPillarScores(t *testing.T) {
	// Given evaluation results with different pillar scores
	result := evaluator.EvaluationResult{
		QualityScore:  9.0,
		CoverageScore: 4.0,
		DeployScore:   7.0,
	}
	cfg := config.Default()

	// When the score is computed
	score := farleyscore.Compute(result, cfg)

	// Then each pillar score is visible
	if score.QualityScore != 9.0 {
		t.Errorf("expected quality 9.0, got %f", score.QualityScore)
	}
	if score.CoverageScore != 4.0 {
		t.Errorf("expected coverage 4.0, got %f", score.CoverageScore)
	}
	if score.DeployScore != 7.0 {
		t.Errorf("expected deploy 7.0, got %f", score.DeployScore)
	}
}

func TestScoreBreakdown_ShowsWeights(t *testing.T) {
	// Given a configuration with specific weights
	cfg := config.Configuration{
		Threshold: 6.0,
		PillarWeights: config.PillarWeights{
			CodeQuality:   0.5,
			TestCoverage:  0.3,
			Deployability: 0.2,
		},
	}
	result := evaluator.EvaluationResult{
		QualityScore:  8.0,
		CoverageScore: 6.0,
		DeployScore:   7.0,
	}

	// When the score is computed
	score := farleyscore.Compute(result, cfg)

	// Then weights are visible in the result
	if score.QualityWeight == 0 {
		t.Error("expected quality weight to be set")
	}
	if score.CoverageWeight == 0 {
		t.Error("expected coverage weight to be set")
	}
	if score.DeployWeight == 0 {
		t.Error("expected deploy weight to be set")
	}
}

func TestScoreBreakdown_IssuesWithPillarAttribution(t *testing.T) {
	// Given evaluation results with issues from different pillars
	result := evaluator.EvaluationResult{
		QualityScore: 7.0,
		CoverageScore: 5.0,
		DeployScore: 8.0,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Warning, Category: evaluator.CodeQuality, Title: "deep nesting"},
			{Severity: evaluator.Warning, Category: evaluator.TestCoverage, Title: "low coverage"},
			{Severity: evaluator.Critical, Category: evaluator.Deployability, Title: "hardcoded secret"},
		},
	}
	cfg := config.Default()

	// When the score is computed
	score := farleyscore.Compute(result, cfg)

	// Then each issue retains its category
	categories := make(map[evaluator.Category]bool)
	for _, issue := range score.Issues {
		categories[issue.Category] = true
	}
	if !categories[evaluator.CodeQuality] {
		t.Error("expected CodeQuality issue in breakdown")
	}
	if !categories[evaluator.TestCoverage] {
		t.Error("expected TestCoverage issue in breakdown")
	}
	if !categories[evaluator.Deployability] {
		t.Error("expected Deployability issue in breakdown")
	}
}
