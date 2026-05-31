package evaluator

import (
	"fmt"
	"strings"

	"github.com/eddie/gatekeeper/internal/diff"
)

// Issue represents a specific quality finding.
type Issue struct {
	File      string
	Line      int
	Severity  Severity
	Category  Category
	Title     string
	Detail    string
	Recommendation string
}

// Severity indicates the issue severity.
type Severity int

const (
	Critical Severity = iota
	Warning
	Info
)

func (s Severity) String() string {
	switch s {
	case Critical:
		return "critical"
	case Warning:
		return "warning"
	case Info:
		return "info"
	default:
		return "unknown"
	}
}

// Category groups issues by type.
type Category int

const (
	CodeQuality Category = iota
	TestCoverage
	Deployability
)

func (c Category) String() string {
	switch c {
	case CodeQuality:
		return "code_quality"
	case TestCoverage:
		return "test_coverage"
	case Deployability:
		return "deployability"
	default:
		return "unknown"
	}
}

// EvaluationResult holds the results of a diff evaluation.
type EvaluationResult struct {
	Issues       []Issue
	QualityScore float64
	CoverageScore float64
	DeployScore  float64
	IsLargeDiff  bool
	LargeDiffWarning string
}

// Evaluate runs all evaluation pillars on a diff.
func Evaluate(d diff.Diff, coverageData map[string]float64) EvaluationResult {
	var issues []Issue

	// Check for large diff
	isLarge := d.TotalChangedLines() > 500
	var largeWarning string
	if isLarge {
		largeWarning = fmt.Sprintf("Large diff detected: %d lines changed (threshold: 500). Evaluation proceeds but may be less detailed.", d.TotalChangedLines())
	}

	// Run each pillar
	issues = append(issues, evaluateCodeQuality(d)...)
	issues = append(issues, evaluateTestCoverage(d, coverageData)...)
	issues = append(issues, evaluateDeployability(d)...)

	// Compute scores
	qualityScore := computeScore(issues, CodeQuality)
	coverageScore := computeScore(issues, TestCoverage)
	deployScore := computeScore(issues, Deployability)

	return EvaluationResult{
		Issues:          issues,
		QualityScore:    qualityScore,
		CoverageScore:   coverageScore,
		DeployScore:     deployScore,
		IsLargeDiff:     isLarge,
		LargeDiffWarning: largeWarning,
	}
}

// evaluateCodeQuality checks for structural problems, anti-patterns, and maintainability.
func evaluateCodeQuality(d diff.Diff) []Issue {
	var issues []Issue

	for _, f := range d.Files {
		if f.Status == diff.Deleted {
			continue
		}
		content := f.NewContent
		if content == "" {
			content = f.NewPath // use path for pattern matching
		}

		issues = append(issues, checkDeepNesting(f)...)
		issues = append(issues, checkLongFunctions(f)...)
		issues = append(issues, checkMagicNumbers(f)...)
		issues = append(issues, checkLongParameterLists(f)...)
	}

	return issues
}

// evaluateTestCoverage checks whether changed code has test coverage.
func evaluateTestCoverage(d diff.Diff, coverageData map[string]float64) []Issue {
	var issues []Issue

	hasCoverageData := len(coverageData) > 0

	for _, f := range d.Files {
		if f.Status == diff.Deleted {
			continue
		}
		path := f.NewPath
		if path == "" {
			path = f.OldPath
		}

		if hasCoverageData {
			if coverage, ok := coverageData[path]; ok {
				if coverage < 50 {
					issues = append(issues, Issue{
						File:      path,
						Severity:  Warning,
						Category:  TestCoverage,
						Title:     fmt.Sprintf("Low test coverage: %.0f%%", coverage),
						Detail:    fmt.Sprintf("File %s has %.0f%% test coverage, below the 50%% threshold.", path, coverage),
						Recommendation: fmt.Sprintf("Add tests to increase coverage of %s above 50%%.", path),
					})
				}
			} else {
				// New or modified file with no coverage data
				if f.Status == diff.Added || f.Status == diff.Modified {
					issues = append(issues, Issue{
						File:      path,
						Severity:  Warning,
						Category:  TestCoverage,
						Title:     "No test coverage data for changed file",
						Detail:    fmt.Sprintf("File %s was changed but has no coverage data.", path),
						Recommendation: fmt.Sprintf("Add tests covering the changes in %s.", path),
					})
				}
			}
		} else {
			// No coverage data at all
			if f.Status == diff.Added || f.Status == diff.Modified {
				issues = append(issues, Issue{
					File:      path,
					Severity:  Warning,
					Category:  TestCoverage,
					Title:     "No test coverage data available",
					Detail:    fmt.Sprintf("File %s was changed but no coverage data was provided. Coverage assessment is incomplete.", path),
					Recommendation: "Run your test suite with coverage enabled and provide coverage data to Gatekeeper.",
				})
			}
		}
	}

	return issues
}

// evaluateDeployability checks for deployment blockers.
func evaluateDeployability(d diff.Diff) []Issue {
	var issues []Issue

	for _, f := range d.Files {
		if f.Status == diff.Deleted {
			continue
		}
		content := f.NewContent
		if content == "" {
			continue
		}

		issues = append(issues, checkHardcodedCredentials(f)...)
		issues = append(issues, checkMissingErrorHandling(f)...)
	}

	return issues
}

func checkDeepNesting(f diff.FileChange) []Issue {
	content := f.NewContent
	if content == "" {
		return nil
	}
	var issues []Issue
	lines := strings.Split(content, "\n")
	indentLevel := 0
	maxIndent := 0
	maxIndentLine := 0

	for i, line := range lines {
		indent := strings.IndexFunc(line, func(r rune) bool {
			return r != ' ' && r != '\t'
		})
		if indent < 0 {
			indent = len(line)
		}
		// Estimate nesting from indentation (4 spaces or 1 tab per level)
		level := indent / 4
		if level > maxIndent {
			maxIndent = level
			maxIndentLine = i + 1
		}
	}

	if maxIndent >= 5 {
		issues = append(issues, Issue{
			File:   f.NewPath,
			Line:   maxIndentLine,
			Severity: Warning,
			Category: CodeQuality,
			Title:    "Deeply nested code detected",
			Detail:   fmt.Sprintf("Code at line %d has nesting level %d (threshold: 4).", maxIndentLine, maxIndent),
			Recommendation: "Extract nested logic into separate functions to reduce nesting depth.",
		})
	}

	_ = indentLevel
	return issues
}

func checkLongFunctions(f diff.FileChange) []Issue {
	content := f.NewContent
	if content == "" {
		return nil
	}
	lines := strings.Split(content, "\n")
	var issues []Issue
	funcStart := -1
	funcName := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "function ") {
			if funcStart >= 0 && i-funcStart > 50 {
				issues = append(issues, Issue{
					File:   f.NewPath,
					Line:   funcStart + 1,
					Severity: Warning,
					Category: CodeQuality,
					Title:    fmt.Sprintf("Long function: %s", funcName),
					Detail:   fmt.Sprintf("Function %s is %d lines long (threshold: 50).", funcName, i-funcStart),
					Recommendation: fmt.Sprintf("Consider breaking %s into smaller, focused functions.", funcName),
				})
			}
			funcStart = i
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				funcName = strings.TrimSuffix(parts[1], "(")
			}
		}
	}

	// Check last function against end of file
	if funcStart >= 0 && len(lines)-funcStart > 50 {
		issues = append(issues, Issue{
			File:   f.NewPath,
			Line:   funcStart + 1,
			Severity: Warning,
			Category: CodeQuality,
			Title:    fmt.Sprintf("Long function: %s", funcName),
			Detail:   fmt.Sprintf("Function %s is %d lines long (threshold: 50).", funcName, len(lines)-funcStart),
			Recommendation: fmt.Sprintf("Consider breaking %s into smaller, focused functions.", funcName),
		})
	}

	return issues
}

func checkMagicNumbers(f diff.FileChange) []Issue {
	content := f.NewContent
	if content == "" {
		return nil
	}
	var issues []Issue
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip comments and constant declarations
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "const ") || strings.HasPrefix(trimmed, "final ") {
			continue
		}
		// Simple magic number detection (numbers not 0, 1, 2, 100)
		// This is a simplified check
		for _, num := range []string{"3.14", "255", "1024", "60000", "86400"} {
			if strings.Contains(trimmed, num) && !strings.Contains(trimmed, "const") {
				issues = append(issues, Issue{
					File:   f.NewPath,
					Line:   i + 1,
					Severity: Info,
					Category: CodeQuality,
					Title:    "Magic number detected",
					Detail:   fmt.Sprintf("Magic number %s found on line %d.", num, i+1),
					Recommendation: "Extract magic numbers into named constants for clarity.",
				})
				break // one per line
			}
		}
	}

	return issues
}

func checkLongParameterLists(f diff.FileChange) []Issue {
	content := f.NewContent
	if content == "" {
		return nil
	}
	var issues []Issue
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "def ") {
			// Count parameters
			if idx := strings.Index(trimmed, "("); idx >= 0 {
				params := strings.Split(trimmed[idx+1:], ",")
				if len(params) > 5 {
					issues = append(issues, Issue{
						File:   f.NewPath,
						Line:   i + 1,
						Severity: Warning,
						Category: CodeQuality,
						Title:    "Long parameter list",
						Detail:   fmt.Sprintf("Function on line %d has %d parameters (threshold: 5).", i+1, len(params)),
						Recommendation: "Consider grouping parameters into a struct or config object.",
					})
				}
			}
		}
	}

	return issues
}

func checkHardcodedCredentials(f diff.FileChange) []Issue {
	content := f.NewContent
	if content == "" {
		return nil
	}
	var issues []Issue
	lines := strings.Split(content, "\n")

	sensitivePatterns := []string{
		"password", "passwd", "secret", "api_key", "apikey", "api_secret",
		"access_token", "private_key", "aws_secret", "db_password",
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		// Skip comments
		if strings.HasPrefix(lower, "//") || strings.HasPrefix(lower, "#") {
			continue
		}
		for _, pattern := range sensitivePatterns {
			if strings.Contains(lower, pattern) &&
				(strings.Contains(lower, "=") || strings.Contains(lower, ":")) {
				// Check if it looks like a hardcoded value (not a variable reference)
				if strings.Contains(lower, "\"") || strings.Contains(lower, "'") {
					issues = append(issues, Issue{
						File:      f.NewPath,
						Line:      i + 1,
						Severity:  Critical,
						Category:  Deployability,
						Title:     "Possible hardcoded credential",
						Detail:    fmt.Sprintf("Line %d appears to contain a hardcoded %s.", i+1, pattern),
						Recommendation: "Use environment variables or a secrets manager instead of hardcoding credentials.",
					})
					break
				}
			}
		}
	}

	return issues
}

func checkMissingErrorHandling(f diff.FileChange) []Issue {
	content := f.NewContent
	if content == "" {
		return nil
	}
	var issues []Issue
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check for HTTP calls without error handling
		if (strings.Contains(trimmed, "http.Get") || strings.Contains(trimmed, "http.Post") ||
			strings.Contains(trimmed, ".Fetch(") || strings.Contains(trimmed, "requests.get")) &&
			!strings.Contains(trimmed, "if") && !strings.Contains(trimmed, "try") {
			issues = append(issues, Issue{
				File:      f.NewPath,
				Line:      i + 1,
				Severity:  Warning,
				Category:  Deployability,
				Title:     "Missing error handling for HTTP call",
				Detail:    fmt.Sprintf("HTTP call on line %d may not handle errors.", i+1),
				Recommendation: "Add error handling for the HTTP call to handle production failure modes.",
			})
		}
	}

	return issues
}

func computeScore(issues []Issue, category Category) float64 {
	score := 10.0
	for _, issue := range issues {
		if issue.Category != category {
			continue
		}
		switch issue.Severity {
		case Critical:
			score -= 3.0
		case Warning:
			score -= 1.5
		case Info:
			score -= 0.5
		}
	}
	if score < 1.0 {
		score = 1.0
	}
	return score
}
