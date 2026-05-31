package output_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/eddie/gatekeeper/internal/config"
	"github.com/eddie/gatekeeper/internal/evaluator"
	"github.com/eddie/gatekeeper/internal/farleyscore"
	"github.com/eddie/gatekeeper/internal/output"
)

// =============================================================================
// Story 4.1: Local Advisory Output
// Acceptance Tests
// =============================================================================

func TestLocalOutput_AboveThreshold(t *testing.T) {
	// Given a score above threshold
	score := farleyscore.ScoreResult{
		FarleyScore:   8.0,
		QualityScore:  9.0,
		CoverageScore: 7.0,
		DeployScore:   8.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Default()

	// When the local output is produced
	outputText := output.TextOutput(score, cfg, output.Local)

	// Then a positive message is displayed
	if !strings.Contains(outputText, "8.0") {
		t.Error("expected score in output")
	}
	if !strings.Contains(outputText, "safe to push") {
		t.Error("expected positive message for above-threshold score")
	}
}

func TestLocalOutput_BelowThreshold(t *testing.T) {
	// Given a score below threshold with issues
	score := farleyscore.ScoreResult{
		FarleyScore:   4.5,
		QualityScore:  5.0,
		CoverageScore: 3.0,
		DeployScore:   5.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Warning, Category: evaluator.CodeQuality, Title: "deep nesting", File: "main.go", Line: 10},
		},
	}
	cfg := config.Default()

	// When the local output is produced
	outputText := output.TextOutput(score, cfg, output.Local)

	// Then a warning with recommendations is displayed
	if !strings.Contains(outputText, "4.5") {
		t.Error("expected score in output")
	}
	if !strings.Contains(outputText, "below threshold") {
		t.Error("expected warning for below-threshold score")
	}
	if !strings.Contains(outputText, "deep nesting") {
		t.Error("expected issue details in output")
	}
}

func TestLocalOutput_NeverBlocks(t *testing.T) {
	// Given a very low score
	score := farleyscore.ScoreResult{
		FarleyScore:   1.0,
		QualityScore:  1.0,
		CoverageScore: 1.0,
		DeployScore:   1.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When the local output is produced
	outputText := output.TextOutput(score, cfg, output.Local)

	// Then it warns but does not block
	if strings.Contains(outputText, "blocked") || strings.Contains(outputText, "BLOCKED") {
		t.Error("expected local mode to never block")
	}
	if !strings.Contains(outputText, "consider improvements") {
		t.Error("expected advisory message")
	}
}

func TestLocalOutput_HumanReadable(t *testing.T) {
	// Given any evaluation result
	score := farleyscore.ScoreResult{FarleyScore: 7.0}
	cfg := config.Default()

	// When the local output is produced
	outputText := output.TextOutput(score, cfg, output.Local)

	// Then it is human-readable (not JSON)
	if strings.HasPrefix(strings.TrimSpace(outputText), "{") {
		t.Error("expected human-readable output, not JSON")
	}
}

// =============================================================================
// Story 4.2: Local Structured Output (JSON)
// Acceptance Tests
// =============================================================================

func TestJSONOutput_ValidJSON(t *testing.T) {
	// Given an evaluation result with issues
	score := farleyscore.ScoreResult{
		FarleyScore:   6.5,
		QualityScore:  8.0,
		CoverageScore: 5.0,
		DeployScore:   7.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Warning, Category: evaluator.CodeQuality, Title: "test issue", File: "main.go", Line: 42},
		},
	}
	cfg := config.Default()

	// When JSON output is produced
	data, err := output.JSONOutput(score, cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Then the output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
}

func TestJSONOutput_ContainsFarleyScore(t *testing.T) {
	// Given an evaluation result
	score := farleyscore.ScoreResult{
		FarleyScore:  7.5,
		QualityScore: 8.0,
		CoverageScore: 7.0,
		DeployScore:  7.0,
	}
	cfg := config.Default()

	// When JSON output is produced
	data, _ := output.JSONOutput(score, cfg)

	// Then it includes the Farley Score and pillar scores
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	if parsed["farley_score"].(float64) != 7.5 {
		t.Errorf("expected farley_score 7.5, got %v", parsed["farley_score"])
	}
	if parsed["quality_score"].(float64) != 8.0 {
		t.Errorf("expected quality_score 8.0, got %v", parsed["quality_score"])
	}
}

func TestJSONOutput_ContainsIssueDetails(t *testing.T) {
	// Given an evaluation result with issues
	score := farleyscore.ScoreResult{
		FarleyScore: 6.0,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Warning, Category: evaluator.CodeQuality, Title: "deep nesting", File: "main.go", Line: 10, Detail: "too deep"},
		},
	}
	cfg := config.Default()

	// When JSON output is produced
	data, _ := output.JSONOutput(score, cfg)

	// Then each issue includes file, line, severity, description
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	issues := parsed["issues"].([]interface{})
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	issue := issues[0].(map[string]interface{})
	if issue["file"] != "main.go" {
		t.Errorf("expected file 'main.go', got %v", issue["file"])
	}
	if int(issue["line"].(float64)) != 10 {
		t.Errorf("expected line 10, got %v", issue["line"])
	}
	if issue["severity"] != "warning" {
		t.Errorf("expected severity 'warning', got %v", issue["severity"])
	}
}

func TestJSONOutput_OptIn(t *testing.T) {
	// Given the default text output
	score := farleyscore.ScoreResult{FarleyScore: 7.0}
	cfg := config.Default()

	// When text output is produced (default)
	textOutput := output.TextOutput(score, cfg, output.Local)

	// Then it is not JSON
	if strings.HasPrefix(strings.TrimSpace(textOutput), "{") {
		t.Error("expected default output to be text, not JSON")
	}
}

// =============================================================================
// Story 4.3: Local Latency Expectation
// (Latency is verified via benchmarks, not unit tests)
// Acceptance Tests
// =============================================================================

func TestTextOutput_CompletesQuickly(t *testing.T) {
	// Given a typical evaluation result
	score := farleyscore.ScoreResult{
		FarleyScore:   7.0,
		QualityScore:  8.0,
		CoverageScore: 6.0,
		DeployScore:   7.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Default()

	// When the output is produced
	_ = output.TextOutput(score, cfg, output.Local)

	// Then it completes (latency verified by benchmarks)
	// This test ensures the function doesn't hang or panic
}

// =============================================================================
// Story 5.1: CI Hard Rejection
// Acceptance Tests
// =============================================================================

func TestCIOutput_HardRejection(t *testing.T) {
	// Given a score below threshold
	score := farleyscore.ScoreResult{
		FarleyScore:   4.0,
		QualityScore:  5.0,
		CoverageScore: 3.0,
		DeployScore:   4.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When CI output is produced
	outputText := output.TextOutput(score, cfg, output.CI)

	// Then rejection is signaled
	if !strings.Contains(outputText, "FAILED") && !strings.Contains(outputText, "Failed") {
		t.Error("expected CI rejection message")
	}
	if !strings.Contains(outputText, "Merge blocked") {
		t.Error("expected merge blocked message")
	}
}

func TestCIOutput_AllowMerge(t *testing.T) {
	// Given a score at or above threshold
	score := farleyscore.ScoreResult{
		FarleyScore:   7.0,
		QualityScore:  8.0,
		CoverageScore: 6.0,
		DeployScore:   7.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When CI output is produced
	outputText := output.TextOutput(score, cfg, output.CI)

	// Then merge is allowed
	if !strings.Contains(outputText, "PASSED") && !strings.Contains(outputText, "Passed") {
		t.Error("expected CI pass message")
	}
	if !strings.Contains(outputText, "Merge allowed") {
		t.Error("expected merge allowed message")
	}
}

// =============================================================================
// Story 5.2: CI Step Summary
// Acceptance Tests
// =============================================================================

func TestCIStepSummary_ContainsScoreAndThreshold(t *testing.T) {
	// Given an evaluation result
	score := farleyscore.ScoreResult{
		FarleyScore:   7.5,
		QualityScore:  8.0,
		CoverageScore: 7.0,
		DeployScore:   7.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When the CI step summary is produced
	summary := output.CIStepSummary(score, cfg)

	// Then it includes score, threshold, status
	if !strings.Contains(summary, "7.5") {
		t.Error("expected score in summary")
	}
	if !strings.Contains(summary, "6.0") {
		t.Error("expected threshold in summary")
	}
	if !strings.Contains(summary, "PASS") {
		t.Error("expected PASS status")
	}
}

func TestCIStepSummary_FailStatus(t *testing.T) {
	// Given a failing score
	score := farleyscore.ScoreResult{FarleyScore: 4.0}
	cfg := config.Configuration{Threshold: 6.0}

	// When the CI step summary is produced
	summary := output.CIStepSummary(score, cfg)

	// Then it shows FAIL status
	if !strings.Contains(summary, "FAIL") {
		t.Error("expected FAIL status")
	}
}

func TestCIStepSummary_PlainText(t *testing.T) {
	// Given any evaluation result
	score := farleyscore.ScoreResult{FarleyScore: 7.0}
	cfg := config.Default()

	// When the CI step summary is produced
	summary := output.CIStepSummary(score, cfg)

	// Then it is plain text (no ANSI escapes)
	if strings.Contains(summary, "\033[") {
		t.Error("expected no ANSI escape codes in summary")
	}
}

// =============================================================================
// Story 5.3: CI Pull Request Comments
// Acceptance Tests
// =============================================================================

func TestPRComment_RejectedPR(t *testing.T) {
	// Given a rejected PR
	score := farleyscore.ScoreResult{
		FarleyScore:   4.0,
		QualityScore:  5.0,
		CoverageScore: 3.0,
		DeployScore:   4.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
		Issues: []evaluator.Issue{
			{Severity: evaluator.Critical, Category: evaluator.Deployability, Title: "hardcoded secret", File: "config.go", Line: 15},
		},
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When the PR comment is produced
	comment := output.MarkdownPRComment(score, cfg)

	// Then it includes score, breakdown, and issues
	if !strings.Contains(comment, "Failed") {
		t.Error("expected failed status in PR comment")
	}
	if !strings.Contains(comment, "4.0") {
		t.Error("expected score in PR comment")
	}
	if !strings.Contains(comment, "hardcoded secret") {
		t.Error("expected issue details in PR comment")
	}
}

func TestPRComment_PassedPR(t *testing.T) {
	// Given a passed PR
	score := farleyscore.ScoreResult{
		FarleyScore:   8.0,
		QualityScore:  9.0,
		CoverageScore: 7.0,
		DeployScore:   8.0,
		QualityWeight: 0.4,
		CoverageWeight: 0.35,
		DeployWeight:  0.25,
	}
	cfg := config.Configuration{Threshold: 6.0}

	// When the PR comment is produced
	comment := output.MarkdownPRComment(score, cfg)

	// Then it confirms the score and approval
	if !strings.Contains(comment, "Passed") {
		t.Error("expected passed status in PR comment")
	}
	if !strings.Contains(comment, "8.0") {
		t.Error("expected score in PR comment")
	}
}

func TestPRComment_MarkdownFormatted(t *testing.T) {
	// Given any evaluation result
	score := farleyscore.ScoreResult{FarleyScore: 7.0}
	cfg := config.Default()

	// When the PR comment is produced
	comment := output.MarkdownPRComment(score, cfg)

	// Then it is markdown-formatted
	if !strings.Contains(comment, "##") {
		t.Error("expected markdown headers")
	}
	if !strings.Contains(comment, "|") {
		t.Error("expected markdown table")
	}
}

// =============================================================================
// Story 5.4: Branch Protection Integration
// Acceptance Tests
// =============================================================================

func TestMode_String(t *testing.T) {
	if output.Local.String() != "local" {
		t.Errorf("expected 'local', got %q", output.Local.String())
	}
	if output.CI.String() != "ci" {
		t.Errorf("expected 'ci', got %q", output.CI.String())
	}
}
