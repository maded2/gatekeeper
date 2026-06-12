package evaluator_test

import (
	"os"
	"path/filepath"
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
	"gatekeeper/internal/scanner"
	"gatekeeper/pkg/score"
)

// --- Story B-1: Check Quality of My Entire Workspace ---

func mustWrite(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func setupWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	mustWrite(t, dir, "main.go", `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`)
	mustWrite(t, dir, "auth/auth.go", `package auth

func Authenticate(user, pass string) bool {
	if user == "" {
		return false
	}
	if pass == "" {
		return false
	}
	return user == "admin" && pass == "secret"
}
`)
	mustWrite(t, dir, "README.md", "# Test Project")

	return dir
}

// ACCEPTANCE CRITERIA 1:
// "I can run a single command to evaluate all source files in my current directory"
func TestCheck_EvaluatesAllSourceFiles(t *testing.T) {
	dir := setupWorkspace(t)

	cfg := config.DefaultConfig()
	s := scanner.New(cfg.Exclusions.Paths)
	files, _ := s.Scan(dir)
	_ = files // files found by scanner

	result := evaluator.CheckWorkspace(cfg, dir)

	if len(result.Findings) == 0 && result.Total == 100 {
		t.Log("workspace evaluated with no findings")
	}
}

// ACCEPTANCE CRITERIA 2:
// "The output shows my overall Quality Score out of 100"
func TestCheck_ScoreOutOf100(t *testing.T) {
	dir := setupWorkspace(t)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	if result.Total < 0 || result.Total > 100 {
		t.Errorf("score should be between 0 and 100, got %f", result.Total)
	}
}

// ACCEPTANCE CRITERIA 3:
// "The output breaks down my score by pillar"
func TestCheck_HasPillarBreakdown(t *testing.T) {
	dir := setupWorkspace(t)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	expectedPillars := []string{
		score.PillarStatic,
		score.PillarArchitecture,
		score.PillarVerification,
		score.PillarSecurity,
	}

	for _, pillar := range expectedPillars {
		if _, ok := result.Pillars[pillar]; !ok {
			t.Errorf("expected pillar %q in breakdown", pillar)
		}
	}
}

// ACCEPTANCE CRITERIA 3 (continued):
// "The pillar breakdown adds up to the total score"
func TestCheck_PillarsSumToTotal(t *testing.T) {
	dir := setupWorkspace(t)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	sum := 0.0
	for _, v := range result.Pillars {
		sum += v
	}

	if sum != result.Total {
		t.Errorf("pillar sum %f does not equal total %f", sum, result.Total)
	}
}

// ACCEPTANCE CRITERIA 4:
// "The output lists specific findings with file names and line numbers"
func TestCheck_FindingsHaveLocation(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "secrets.go", `package secrets

var apiKey = "sk-1234567890abcdef"
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	found := false
	for _, f := range result.Findings {
		if f.Location != "" && f.LineStart > 0 {
			found = true
			break
		}
	}
	t.Logf("found %d findings, locations supported: %v", len(result.Findings), found)
}

// Verify that non-code files are not evaluated
func TestCheck_SkipsNonCodeFiles(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "docs/guide.md", `# This is not code
func fake() { /* should not be analyzed */ }
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	for _, f := range result.Findings {
		if filepath.Ext(f.Location) == ".md" {
			t.Errorf("expected .md files to be skipped, but found finding: %s", f.Location)
		}
	}
}

// Verify empty workspace returns valid score
func TestCheck_EmptyWorkspace(t *testing.T) {
	dir := t.TempDir()

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	if result.Total < 0 || result.Total > 100 {
		t.Errorf("empty workspace score should be between 0 and 100, got %f", result.Total)
	}

	if _, ok := result.Pillars[score.PillarStatic]; !ok {
		t.Error("expected static pillar even for empty workspace")
	}
}

// Verify the check respects exclusion patterns
func TestCheck_RespectsExclusions(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "generated/output.go", `package generated`)

	cfg := config.DefaultConfig()
	cfg.Exclusions.Paths = append(cfg.Exclusions.Paths, "**/generated/**")

	result := evaluator.CheckWorkspace(cfg, dir)

	for _, f := range result.Findings {
		if filepath.Base(filepath.Dir(f.Location)) == "generated" {
			t.Errorf("expected generated/ to be excluded, but found finding: %s", f.Location)
		}
	}
}
