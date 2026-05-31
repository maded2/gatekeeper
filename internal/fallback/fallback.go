package fallback

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/eddie/gatekeeper/internal/config"
	"github.com/eddie/gatekeeper/internal/diff"
	"github.com/eddie/gatekeeper/internal/evaluator"
)

// FallbackResult holds the result of a fallback evaluation.
type FallbackResult struct {
	IsFallback     bool
	QualityScore   float64
	CoverageScore  float64
	DeployScore    float64
	Issues         []evaluator.Issue
	RetryAttempts  int
	LastError      string
	FallbackReason string
}

// Evaluate runs a deterministic fallback evaluation when LLM API is unavailable.
func Evaluate(d diff.Diff, cfg config.Configuration) FallbackResult {
	var issues []evaluator.Issue

	// Deterministic rules-based checks
	issues = append(issues, fallbackQualityCheck(d)...)
	issues = append(issues, fallbackDeployCheck(d)...)

	// Coverage score is unknown in fallback mode
	coverageScore := 5.0 // neutral

	qualityScore := computeFallbackScore(issues, evaluator.CodeQuality)
	deployScore := computeFallbackScore(issues, evaluator.Deployability)

	return FallbackResult{
		IsFallback:     true,
		QualityScore:   qualityScore,
		CoverageScore:  coverageScore,
		DeployScore:    deployScore,
		Issues:         issues,
		FallbackReason: "LLM API unavailable — using deterministic fallback rules",
	}
}

func fallbackQualityCheck(d diff.Diff) []evaluator.Issue {
	var issues []evaluator.Issue

	for _, f := range d.Files {
		if f.Status == diff.Deleted {
			continue
		}

		// Check for very large files (>1000 lines)
		if f.NewContent != "" {
			lineCount := strings.Count(f.NewContent, "\n")
			if lineCount > 1000 {
				issues = append(issues, evaluator.Issue{
					File:      f.NewPath,
					Severity:  evaluator.Warning,
					Category:  evaluator.CodeQuality,
					Title:     "Large file detected",
					Detail:    fmt.Sprintf("File %s has %d lines (threshold: 1000).", f.NewPath, lineCount),
					Recommendation: "Consider splitting large files into smaller, focused modules.",
				})
			}
		}
	}

	return issues
}

func fallbackDeployCheck(d diff.Diff) []evaluator.Issue {
	var issues []evaluator.Issue

	for _, f := range d.Files {
		if f.Status == diff.Deleted {
			continue
		}
		content := f.NewContent
		if content == "" {
			continue
		}

		// Check for hardcoded credentials (pattern matching)
		sensitivePatterns := []string{"password", "secret", "api_key", "private_key"}
		for i, line := range strings.Split(content, "\n") {
			lower := strings.ToLower(line)
			for _, pattern := range sensitivePatterns {
				if strings.Contains(lower, pattern) &&
					(strings.Contains(lower, "\"") || strings.Contains(lower, "'")) {
					issues = append(issues, evaluator.Issue{
						File:      f.NewPath,
						Line:      i + 1,
						Severity:  evaluator.Critical,
						Category:  evaluator.Deployability,
						Title:     "Possible hardcoded credential",
						Detail:    fmt.Sprintf("Line %d may contain a hardcoded %s.", i+1, pattern),
						Recommendation: "Use environment variables or a secrets manager.",
					})
					break
				}
			}
		}
	}

	return issues
}

func computeFallbackScore(issues []evaluator.Issue, category evaluator.Category) float64 {
	score := 10.0
	for _, issue := range issues {
		if issue.Category != category {
			continue
		}
		switch issue.Severity {
		case evaluator.Critical:
			score -= 3.0
		case evaluator.Warning:
			score -= 1.5
		case evaluator.Info:
			score -= 0.5
		}
	}
	if score < 1.0 {
		score = 1.0
	}
	return score
}

// RetryWithBackoff retries an operation with exponential backoff.
func RetryWithBackoff(cfg config.Configuration, operation func() error) (int, error) {
	maxRetries := cfg.Retry.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	initialDelay := time.Duration(cfg.Retry.InitialDelayMs) * time.Millisecond
	if initialDelay <= 0 {
		initialDelay = time.Second
	}

	maxDelay := time.Duration(cfg.Retry.MaxDelayMs) * time.Millisecond
	if maxDelay <= 0 {
		maxDelay = 10 * time.Second
	}

	multiplier := cfg.Retry.BackoffMultiplier
	if multiplier <= 0 {
		multiplier = 2.0
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		lastErr = operation()
		if lastErr == nil {
			return attempt, nil
		}

		if attempt < maxRetries {
			delay := initialDelay * time.Duration(math.Pow(multiplier, float64(attempt)))
			if delay > maxDelay {
				delay = maxDelay
			}
			time.Sleep(delay)
		}
	}

	return maxRetries, lastErr
}
