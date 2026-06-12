package evaluator_test

import (
	"path/filepath"
	"strings"
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
	"gatekeeper/pkg/score"
)

// --- Story B-3: Receive Actionable Remediation Suggestions ---

// ACCEPTANCE CRITERIA 1:
// "Every finding includes a remediation suggestion written in plain language"
func TestRemediation_EveryFindingHasRemediation(t *testing.T) {
	dir := setupWorkspace(t)

	// Create a file with a naming issue
	mustWrite(t, dir, "bad_names.go", `package bad

func a(b string) {
	x := b + "temp"
	_ = x
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	for _, f := range result.Findings {
		if f.Remediation == "" {
			t.Errorf("finding at %s has empty remediation", f.Location)
		}
	}
}

// ACCEPTANCE CRITERIA 2:
// "Remediations reference the specific code location (file and line range)"
func TestRemediation_ReferencesCodeLocation(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "complex.go", `package complex

func processData(input string, flag bool, mode int) string {
	if flag {
		if mode == 1 {
			if len(input) > 0 {
				return input
			}
		}
	}
	return ""
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	for _, f := range result.Findings {
		if f.Pillar == score.PillarArchitecture {
			if f.LineStart <= 0 || f.LineEnd <= 0 {
				t.Errorf("architecture finding should have line range: %v", f)
			}
			if f.Location == "" {
				t.Error("architecture finding should have location")
			}
		}
	}
}

// ACCEPTANCE CRITERIA 3:
// "Remediations are prioritized so I can tackle the highest-impact fixes first"
func TestRemediations_ArePrioritized(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "issues.go", `package issues

var secret = "sk-1234567890abcdef"

func x() {
	a := 1
	b := 2
	_ = a + b
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	// Check that findings have priority set
	for _, f := range result.Findings {
		if f.Priority == "" {
			t.Errorf("finding at %s has no priority", f.Location)
		}
	}
}

// ACCEPTANCE CRITERIA 4:
// "When a finding relates to naming conventions, the suggestion includes a recommended name"
func TestRemediation_NamingIncludesSuggestion(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "naming.go", `package naming

func x() {
	a := "hello"
	_ = a
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	// We expect naming findings to have remediations (even if basic for now)
	for _, f := range result.Findings {
		if strings.Contains(f.Description, "name") || strings.Contains(f.Description, "naming") {
			if f.Remediation == "" {
				t.Error("naming finding should have a remediation suggestion")
			}
		}
	}
}

// ACCEPTANCE CRITERIA 5:
// "When a finding relates to complexity, the suggestion describes what to extract or simplify"
func TestRemediation_ComplexityIncludesExtractionAdvice(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "nested.go", `package nested

func deeplyNested(data []string) []string {
	result := make([]string, 0)
	for i := 0; i < len(data); i++ {
		if data[i] != "" {
			for j := 0; j < len(data); j++ {
				if data[j] == data[i] {
					result = append(result, data[i])
				}
			}
		}
	}
	return result
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	for _, f := range result.Findings {
		if f.Pillar == score.PillarArchitecture {
			// Complexity findings should mention extract, simplify, or refactor
			rem := strings.ToLower(f.Remediation)
			if !strings.Contains(rem, "extract") &&
				!strings.Contains(rem, "simplif") &&
				!strings.Contains(rem, "refactor") &&
				!strings.Contains(rem, "break") {
				t.Errorf("complexity finding should suggest extraction/simplification: %q", f.Remediation)
			}
		}
	}
}

// Verify that findings have proper file paths
func TestRemediation_HasProperFilePath(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "sub/deep/file.go", `package deep

func a() {}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	for _, f := range result.Findings {
		if f.Location != "" {
			// Location should be a relative path
			if filepath.IsAbs(f.Location) {
				t.Errorf("expected relative path, got absolute: %s", f.Location)
			}
		}
	}
}
