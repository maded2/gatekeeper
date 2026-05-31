package output_test

import (
	"strings"
	"testing"

	"github.com/eddie/gatekeeper/internal/config"
	"github.com/eddie/gatekeeper/internal/evaluator"
	"github.com/eddie/gatekeeper/internal/farleyscore"
	"github.com/eddie/gatekeeper/internal/output"
)

// =============================================================================
// Story 6.1: Actionable Rejection Feedback
// Acceptance Tests
// =============================================================================

func TestFeedback_RejectionWithSpecificRecommendations(t *testing.T) {
	// Given a rejected evaluation with specific issues
	score := farleyscore.ScoreResult{
		FarleyScore:   4.5,
		QualityScore:  5.0,
		CoverageScore: 3.0,
		DeployScore:   5.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
		Issues: []evaluator.Issue{
			{
				Severity:       evaluator.Critical,
				Category:       evaluator.Deployability,
				Title:          "hardcoded credential",
				File:           "config.go",
				Line:           15,
				Detail:         "API key found in source code",
				Recommendation: "Use environment variables instead",
			},
			{
				Severity:       evaluator.Warning,
				Category:       evaluator.CodeQuality,
				Title:          "deep nesting",
				File:           "main.go",
				Line:           45,
				Detail:         "Nesting level 6 exceeds threshold",
				Recommendation: "Extract nested logic into separate functions",
			},
		},
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When feedback is produced
	feedback := output.FeedbackOutput(score, cfg)

	// Then recommendations are specific and line-level
	if !strings.Contains(feedback, "config.go") {
		t.Error("expected file reference in feedback")
	}
	if !strings.Contains(feedback, "15") {
		t.Error("expected line reference in feedback")
	}
	if !strings.Contains(feedback, "environment variables") {
		t.Error("expected specific fix recommendation")
	}
}

func TestFeedback_PrioritizedBySeverity(t *testing.T) {
	// Given issues with different severities
	score := farleyscore.ScoreResult{
		FarleyScore:   4.0,
		QualityScore:  5.0,
		CoverageScore: 3.0,
		DeployScore:   4.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Info, Category: evaluator.CodeQuality, Title: "info issue"},
			{Severity: evaluator.Critical, Category: evaluator.Deployability, Title: "critical issue"},
			{Severity: evaluator.Warning, Category: evaluator.TestCoverage, Title: "warning issue"},
		},
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When feedback is produced
	feedback := output.FeedbackOutput(score, cfg)

	// Then critical issues appear first
	criticalIdx := strings.Index(feedback, "critical issue")
	warningIdx := strings.Index(feedback, "warning issue")
	infoIdx := strings.Index(feedback, "info issue")

	if criticalIdx >= warningIdx || criticalIdx >= infoIdx {
		t.Error("expected critical issues to appear first")
	}
	if warningIdx >= infoIdx {
		t.Error("expected warning issues before info issues")
	}
}

func TestFeedback_NoVagueLanguage(t *testing.T) {
	// Given a rejection with specific issues
	score := farleyscore.ScoreResult{
		FarleyScore:   4.0,
		QualityScore:  5.0,
		CoverageScore: 3.0,
		DeployScore:   4.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
		Issues: []evaluator.Issue{
			{
				Severity:       evaluator.Warning,
				Category:       evaluator.CodeQuality,
				Title:          "deep nesting",
				File:           "main.go",
				Line:           45,
				Detail:         "Nesting level 6 exceeds threshold of 4",
				Recommendation: "Extract the nested loop on lines 45-52 into a separate function",
			},
		},
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When feedback is produced
	feedback := output.FeedbackOutput(score, cfg)

	// Then it avoids vague language
	if strings.Contains(feedback, "improve code quality") {
		t.Error("expected no vague 'improve code quality' language")
	}
	if strings.Contains(feedback, "lines 45-52") {
		// Specific line references are good
	}
}

// =============================================================================
// Story 6.2: Positive Feedback for Clean Code
// Acceptance Tests
// =============================================================================

func TestFeedback_PositiveForHighScore(t *testing.T) {
	// Given a high-scoring evaluation
	score := farleyscore.ScoreResult{
		FarleyScore:   8.5,
		QualityScore:  9.0,
		CoverageScore: 8.0,
		DeployScore:   8.5,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When feedback is produced
	feedback := output.FeedbackOutput(score, cfg)

	// Then positive feedback is displayed
	if !strings.Contains(feedback, "passed") && !strings.Contains(feedback, "Passed") {
		t.Error("expected positive message")
	}
	if !strings.Contains(feedback, "8.5") {
		t.Error("expected Farley Score in feedback")
	}
}

func TestFeedback_ExcellentCodeReceivesRecognition(t *testing.T) {
	// Given an excellent score (>= 8.0)
	score := farleyscore.ScoreResult{
		FarleyScore:   9.0,
		QualityScore:  10.0,
		CoverageScore: 9.0,
		DeployScore:   8.5,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When feedback is produced
	feedback := output.FeedbackOutput(score, cfg)

	// Then it includes encouraging language
	if !strings.Contains(feedback, "Excellent") && !strings.Contains(feedback, "excellent") {
		t.Error("expected recognition for excellent code")
	}
}

func TestFeedback_NoFalsePositives(t *testing.T) {
	// Given a passing score with some minor issues
	score := farleyscore.ScoreResult{
		FarleyScore:   6.5,
		QualityScore:  7.0,
		CoverageScore: 6.0,
		DeployScore:   6.5,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Info, Category: evaluator.CodeQuality, Title: "minor style issue"},
		},
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When feedback is produced
	feedback := output.FeedbackOutput(score, cfg)

	// Then it passes but doesn't claim perfection
	if strings.Contains(feedback, "Excellent") {
		t.Error("expected no 'excellent' for borderline passing score")
	}
}

// =============================================================================
// Story 6.3: Evaluation Evidence
// Acceptance Tests
// =============================================================================

func TestEvidence_ContainsRawData(t *testing.T) {
	// Given an evaluation result
	score := farleyscore.ScoreResult{
		FarleyScore:    6.5,
		QualityScore:   7.0,
		CoverageScore:  6.0,
		DeployScore:    6.5,
		QualityWeight:  0.4,
		CoverageWeight: 0.35,
		DeployWeight:   0.25,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Warning, Category: evaluator.CodeQuality, Title: "test issue", Line: 10},
		},
	}
	cfg := config.Default()

	// When evidence is produced
	evidence := output.EvidenceOutput(score, cfg, "3 files, 120 lines")

	// Then it includes raw evaluation data
	if !strings.Contains(evidence, "Threshold") {
		t.Error("expected threshold in evidence")
	}
	if !strings.Contains(evidence, "6.5") {
		t.Error("expected score in evidence")
	}
	if !strings.Contains(evidence, "test issue") {
		t.Error("expected issues in evidence")
	}
}

func TestEvidence_ContainsDiffDescription(t *testing.T) {
	// Given an evaluation with diff metadata
	score := farleyscore.ScoreResult{FarleyScore: 7.0}
	cfg := config.Default()
	diffDesc := "Working directory changes: 3 file(s), 120 line(s)"

	// When evidence is produced
	evidence := output.EvidenceOutput(score, cfg, diffDesc)

	// Then diff description is included
	if !strings.Contains(evidence, "3 file(s)") {
		t.Error("expected diff description in evidence")
	}
}

func TestEvidence_NoSensitiveData(t *testing.T) {
	// Given an evaluation result
	score := farleyscore.ScoreResult{FarleyScore: 7.0}
	cfg := config.Default()

	// When evidence is produced
	evidence := output.EvidenceOutput(score, cfg, "normal diff")

	// Then no sensitive data is exposed
	if strings.Contains(evidence, "password") || strings.Contains(evidence, "secret") ||
		strings.Contains(evidence, "api_key") {
		t.Error("expected no sensitive data in evidence")
	}
}

// =============================================================================
// Story 11.3: Borderline Score Handling
// Acceptance Tests
// =============================================================================

func TestBorderline_WithinThreshold(t *testing.T) {
	// Given a score within 0.5 of threshold
	score := farleyscore.ScoreResult{
		FarleyScore:   5.6,
		QualityScore:  6.0,
		CoverageScore: 5.0,
		DeployScore:   5.5,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Warning, Category: evaluator.CodeQuality, Title: "minor issue"},
		},
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When feedback is produced
	feedback := output.FeedbackOutput(score, cfg)

	// Then borderline status is indicated
	if !strings.Contains(feedback, "Borderline") && !strings.Contains(feedback, "borderline") {
		t.Error("expected borderline indication")
	}
}

func TestBorderline_NotWithinThreshold(t *testing.T) {
	// Given a score far below threshold
	score := farleyscore.ScoreResult{
		FarleyScore:   3.0,
		QualityScore:  4.0,
		CoverageScore: 2.0,
		DeployScore:   3.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When feedback is produced
	feedback := output.FeedbackOutput(score, cfg)

	// Then borderline is NOT indicated
	if strings.Contains(feedback, "Borderline") {
		t.Error("expected no borderline indication for clearly failing score")
	}
}
