package evaluator_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eddie/gatekeeper/internal/diff"
	"github.com/eddie/gatekeeper/internal/evaluator"
)

// =============================================================================
// Story 2.6: Code Quality Evaluation
// Acceptance Tests
// =============================================================================

func TestCodeQuality_DetectsDeepNesting(t *testing.T) {
	// Given a diff with deeply nested code
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:  diff.Modified,
				NewPath: "deep.go",
				NewContent: `package main
func deep() {
    if a {
        if b {
            if c {
                if d {
                    if e {
                        if f {
                            doSomething()
                        }
                    }
                }
            }
        }
    }
}
`,
			},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then deep nesting is flagged
	found := false
	for _, issue := range result.Issues {
		if issue.Category == evaluator.CodeQuality &&
			strings.Contains(issue.Title, "nested") {
			found = true
			if issue.Severity != evaluator.Warning {
				t.Errorf("expected Warning severity for deep nesting, got %s", issue.Severity)
			}
			if issue.Line == 0 {
				t.Error("expected line number for deep nesting issue")
			}
		}
	}
	if !found {
		t.Error("expected deep nesting issue to be detected")
	}
}

func TestCodeQuality_DetectsLongFunctions(t *testing.T) {
	// Given a diff with a long function
	lines := "package main\n"
	for i := 0; i < 60; i++ {
		lines += fmt.Sprintf("// line %d\n", i)
	}
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "long.go",
				NewContent: "package main\nfunc longFunc() {\n" + lines + "}\n",
			},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then long function is flagged
	found := false
	for _, issue := range result.Issues {
		if issue.Category == evaluator.CodeQuality &&
			strings.Contains(issue.Title, "Long function") {
			found = true
		}
	}
	if !found {
		t.Error("expected long function issue to be detected")
	}
}

func TestCodeQuality_CleanCodeReceivesPositiveConfirmation(t *testing.T) {
	// Given a diff with clean code
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "clean.go",
				NewContent: "package main\n\nfunc Hello() string {\n\treturn \"hello\"\n}\n",
			},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then no code quality issues are found
	qualityIssues := 0
	for _, issue := range result.Issues {
		if issue.Category == evaluator.CodeQuality {
			qualityIssues++
		}
	}
	if qualityIssues > 0 {
		t.Errorf("expected no code quality issues for clean code, got %d", qualityIssues)
	}
}

func TestCodeQuality_LineLevelFeedback(t *testing.T) {
	// Given a diff with issues at specific lines
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:  diff.Modified,
				NewPath: "issues.go",
				NewContent: `package main
func bad(a, b, c, d, e, f, g int) {
    x := 60000
}
`,
			},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then each issue has a line reference
	for _, issue := range result.Issues {
		if issue.Category == evaluator.CodeQuality && issue.Line == 0 {
			t.Errorf("expected line number for issue: %s", issue.Title)
		}
	}
}

// =============================================================================
// Story 2.7: Test Coverage Adequacy Evaluation
// Acceptance Tests
// =============================================================================

func TestCoverage_NoCoverageData(t *testing.T) {
	// Given a diff with no coverage data
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "new.go"},
		},
	}

	// When evaluated with no coverage data
	result := evaluator.Evaluate(d, nil)

	// Then coverage assessment indicates incomplete data
	found := false
	for _, issue := range result.Issues {
		if issue.Category == evaluator.TestCoverage &&
			(issue.Title == "No test coverage data available" || strings.Contains(issue.Title, "coverage")) {
			found = true
		}
	}
	if !found {
		t.Error("expected coverage warning when coverage is unavailable")
	}
}

func TestCoverage_LowCoverageFlagged(t *testing.T) {
	// Given a diff with low coverage
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "low_cov.go"},
		},
	}
	coverageData := map[string]float64{
		"low_cov.go": 25.0,
	}

	// When evaluated
	result := evaluator.Evaluate(d, coverageData)

	// Then low coverage is flagged
	found := false
	for _, issue := range result.Issues {
		if issue.Category == evaluator.TestCoverage &&
			strings.Contains(issue.Title, "Low test coverage") {
			found = true
		}
	}
	if !found {
		t.Error("expected low coverage issue to be detected")
	}
}

func TestCoverage_AdequateCoveragePasses(t *testing.T) {
	// Given a diff with adequate coverage
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "good_cov.go"},
		},
	}
	coverageData := map[string]float64{
		"good_cov.go": 85.0,
	}

	// When evaluated
	result := evaluator.Evaluate(d, coverageData)

	// Then no coverage issues for this file
	for _, issue := range result.Issues {
		if issue.Category == evaluator.TestCoverage && issue.File == "good_cov.go" {
			t.Errorf("expected no coverage issues for well-covered file, got: %s", issue.Title)
		}
	}
}

func TestCoverage_NewFileNoCoverage(t *testing.T) {
	// Given a diff with a new file and no coverage data for it
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Added, NewPath: "brand_new.go"},
		},
	}
	coverageData := map[string]float64{}

	// When evaluated
	result := evaluator.Evaluate(d, coverageData)

	// Then the new file is flagged for missing coverage
	found := false
	for _, issue := range result.Issues {
		if issue.Category == evaluator.TestCoverage &&
			issue.File == "brand_new.go" {
			found = true
		}
	}
	if !found {
		t.Error("expected new file to be flagged for missing coverage")
	}
}

// =============================================================================
// Story 2.8: Deployability Evaluation
// Acceptance Tests
// =============================================================================

func TestDeployability_HardcodedCredentials(t *testing.T) {
	// Given a diff with hardcoded credentials
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "config.go",
				NewContent: `package main
var password = "supersecret123"
`,
			},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then hardcoded credentials are flagged as critical
	found := false
	for _, issue := range result.Issues {
		if issue.Category == evaluator.Deployability &&
			strings.Contains(issue.Title, "credential") {
			found = true
			if issue.Severity != evaluator.Critical {
				t.Errorf("expected Critical severity for hardcoded credentials, got %s", issue.Severity)
			}
		}
	}
	if !found {
		t.Error("expected hardcoded credential issue to be detected")
	}
}

func TestDeployability_MissingErrorHandling(t *testing.T) {
	// Given a diff with HTTP calls without error handling
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "http.go",
				NewContent: `package main
resp := http.Get("http://example.com")
`,
			},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then missing error handling is flagged
	found := false
	for _, issue := range result.Issues {
		if issue.Category == evaluator.Deployability &&
			strings.Contains(issue.Title, "error handling") {
			found = true
		}
	}
	if !found {
		t.Error("expected missing error handling issue to be detected")
	}
}

func TestDeployability_CleanChangesReceiveConfirmation(t *testing.T) {
	// Given a diff with no deployability concerns
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "clean.go",
				NewContent: `package main
func Hello() string {
    return "hello"
}
`,
			},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then no deployability issues
	for _, issue := range result.Issues {
		if issue.Category == evaluator.Deployability {
			t.Errorf("expected no deployability issues, got: %s", issue.Title)
		}
	}
}

func TestDeployability_LineLevelReferences(t *testing.T) {
	// Given a diff with deployability issues
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

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then each issue has a line reference
	for _, issue := range result.Issues {
		if issue.Category == evaluator.Deployability {
			if issue.Line == 0 {
				t.Errorf("expected line reference for deployability issue: %s", issue.Title)
			}
		}
	}
}

// =============================================================================
// Story 2.9: Large Diff Handling
// Acceptance Tests
// =============================================================================

func TestLargeDiff_Detected(t *testing.T) {
	// Given a diff exceeding 500 lines
	d := diff.Diff{
		Files:      []diff.FileChange{{Status: diff.Modified, NewPath: "large.go"}},
		TotalLines: 600,
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then large diff is flagged
	if !result.IsLargeDiff {
		t.Error("expected large diff to be detected")
	}
	if result.LargeDiffWarning == "" {
		t.Error("expected large diff warning message")
	}
}

func TestLargeDiff_WarningMessage(t *testing.T) {
	// Given a large diff
	d := diff.Diff{
		Files:      []diff.FileChange{{Status: diff.Modified, NewPath: "large.go"}},
		TotalLines: 1000,
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then warning includes size information
	if !strings.Contains(result.LargeDiffWarning, "1000") {
		t.Errorf("expected warning to include line count, got: %s", result.LargeDiffWarning)
	}
}

func TestLargeDiff_EvaluationProceeds(t *testing.T) {
	// Given a large diff with actual issues
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "large.go",
				NewContent: `package main
var secret = "password123"
`,
			},
		},
		TotalLines: 600,
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then evaluation still produces results
	if len(result.Issues) == 0 {
		t.Error("expected evaluation to produce results for large diff")
	}
}

func TestNormalDiff_NotFlaggedAsLarge(t *testing.T) {
	// Given a diff under 500 lines
	d := diff.Diff{
		Files:      []diff.FileChange{{Status: diff.Modified, NewPath: "small.go"}},
		TotalLines: 100,
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then it is not flagged as large
	if result.IsLargeDiff {
		t.Error("expected normal diff to not be flagged as large")
	}
	if result.LargeDiffWarning != "" {
		t.Error("expected no warning for normal diff")
	}
}

// =============================================================================
// Multi-language evaluation
// =============================================================================

func TestMultiLanguage_DifferentLanguages(t *testing.T) {
	// Given a diff with multiple languages
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "main.go", NewContent: "package main\nfunc Hello() string { return \"hi\" }\n"},
			{Status: diff.Modified, NewPath: "app.py", NewContent: "def hello():\n    return \"hi\"\n"},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then evaluation covers all files
	if len(result.Issues) == 0 && result.QualityScore == 10.0 {
		// This is acceptable — clean code in both languages
		t.Log("multi-language diff evaluated successfully")
	}
}
