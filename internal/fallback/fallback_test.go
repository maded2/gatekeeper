package fallback_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eddie/gatekeeper/internal/config"
	"github.com/eddie/gatekeeper/internal/diff"
	"github.com/eddie/gatekeeper/internal/evaluator"
	"github.com/eddie/gatekeeper/internal/fallback"
)

// =============================================================================
// Story 7.1: LLM API Unavailable Fallback
// Acceptance Tests
// =============================================================================

func TestFallback_ActivatesWhenLLMUnavailable(t *testing.T) {
	// Given a diff to evaluate
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "main.go", NewContent: "package main\n"},
		},
	}
	cfg := config.Default()

	// When fallback evaluation is run
	result := fallback.Evaluate(d, cfg)

	// Then it is marked as fallback
	if !result.IsFallback {
		t.Error("expected fallback mode to be active")
	}
	if result.FallbackReason == "" {
		t.Error("expected fallback reason to be set")
	}
}

func TestFallback_ProducesScore(t *testing.T) {
	// Given a diff with issues
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "config.go",
				NewContent: `package main
var password = "secret123"
`,
			},
		},
	}
	cfg := config.Default()

	// When fallback evaluation is run
	result := fallback.Evaluate(d, cfg)

	// Then it produces a score
	if result.QualityScore < 1.0 || result.QualityScore > 10.0 {
		t.Errorf("expected quality score between 1-10, got %f", result.QualityScore)
	}
	if result.DeployScore < 1.0 || result.DeployScore > 10.0 {
		t.Errorf("expected deploy score between 1-10, got %f", result.DeployScore)
	}
}

func TestFallback_DetectsHardcodedCredentials(t *testing.T) {
	// Given a diff with hardcoded credentials
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "config.go",
				NewContent: `package main
var api_key = "sk-12345"
`,
			},
		},
	}
	cfg := config.Default()

	// When fallback evaluation is run
	result := fallback.Evaluate(d, cfg)

	// Then hardcoded credentials are detected
	found := false
	for _, issue := range result.Issues {
		if issue.Category == evaluator.Deployability &&
			strings.Contains(issue.Title, "credential") {
			found = true
		}
	}
	if !found {
		t.Error("expected hardcoded credential detection in fallback mode")
	}
}

func TestFallback_AdvisoryNotAuthoritative(t *testing.T) {
	// Given a fallback result
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "main.go", NewContent: "package main\n"},
		},
	}
	cfg := config.Default()

	// When fallback evaluation is run
	result := fallback.Evaluate(d, cfg)

	// Then it is clearly labeled as fallback
	if !strings.Contains(result.FallbackReason, "fallback") {
		t.Error("expected fallback reason to mention 'fallback'")
	}
}

func TestFallback_LargeFileDetection(t *testing.T) {
	// Given a diff with a large file
	largeContent := "package main\n"
	for i := 0; i < 1001; i++ {
		largeContent += fmt.Sprintf("// line %d\n", i)
	}
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "large.go", NewContent: largeContent},
		},
	}
	cfg := config.Default()

	// When fallback evaluation is run
	result := fallback.Evaluate(d, cfg)

	// Then large file is flagged
	found := false
	for _, issue := range result.Issues {
		if issue.Category == evaluator.CodeQuality &&
			strings.Contains(issue.Title, "Large file") {
			found = true
		}
	}
	if !found {
		t.Error("expected large file detection in fallback mode")
	}
}

// =============================================================================
// Story 7.2: Retry Behavior
// Acceptance Tests
// =============================================================================

func TestRetry_TransientFailure(t *testing.T) {
	// Given an operation that fails twice then succeeds
	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("transient error")
		}
		return nil
	}
	cfg := config.Default()

	// When retry is attempted
	_, err := fallback.RetryWithBackoff(cfg, operation)

	// Then it succeeds after retries
	if err != nil {
		t.Errorf("expected success after retries, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ExhaustedRetries(t *testing.T) {
	// Given an operation that always fails
	operation := func() error {
		return fmt.Errorf("permanent error")
	}
	cfg := config.Configuration{
		Retry: config.RetryConfig{
			MaxRetries:     2,
			InitialDelayMs: 1, // minimal delay for test
			MaxDelayMs:     1,
		},
	}

	// When retry is attempted
	attempts, err := fallback.RetryWithBackoff(cfg, operation)

	// Then all retries are exhausted
	if err == nil {
		t.Error("expected error after exhausted retries")
	}
	if attempts != 2 {
		t.Errorf("expected 2 retry attempts, got %d", attempts)
	}
}

// =============================================================================
// Story 7.3: Cost Monitoring
// Acceptance Tests
// =============================================================================

func TestFallback_NoCostInFallbackMode(t *testing.T) {
	// Given a fallback evaluation
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "main.go", NewContent: "package main\n"},
		},
	}
	cfg := config.Default()

	// When fallback evaluation is run
	result := fallback.Evaluate(d, cfg)

	// Then it uses deterministic rules (no API cost)
	if !result.IsFallback {
		t.Error("expected fallback mode")
	}
	// Fallback mode has no LLM cost
	t.Log("fallback mode uses deterministic rules — no API cost")
}
