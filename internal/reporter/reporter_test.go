package reporter_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"gatekeeper/internal/reporter"
	"gatekeeper/pkg/score"
)

// --- Story B-2: See Readable Findings in My Terminal ---

func sampleScore() score.Score {
	return score.Score{
		Total:     72.5,
		Timestamp: time.Now(),
		Pillars: map[string]float64{
			score.PillarStatic:       15,
			score.PillarArchitecture: 20,
			score.PillarVerification: 25,
			score.PillarSecurity:     12.5,
		},
		Findings: []score.Finding{
			{
				Priority:    "HIGH",
				Pillar:      score.PillarStatic,
				Location:    "auth/auth.go",
				LineStart:   5,
				LineEnd:     12,
				Description: "Hardcoded credentials detected",
				Remediation: "Move credentials to environment variables",
			},
			{
				Priority:    "MEDIUM",
				Pillar:      score.PillarArchitecture,
				Location:    "main.go",
				LineStart:   10,
				LineEnd:     25,
				Description: "Function has cyclomatic complexity of 15",
				Remediation: "Extract conditional logic into helper functions",
			},
			{
				Priority:    "LOW",
				Pillar:      score.PillarStatic,
				Location:    "utils.go",
				LineStart:   3,
				LineEnd:     3,
				Description: "Unused import 'fmt'",
				Remediation: "Remove the unused import",
			},
		},
	}
}

// ACCEPTANCE CRITERIA 1:
// "Findings are displayed in a tabular format with columns for priority, file location, and description"
func TestPrettyPrint_HasTabularFormat(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewPretty(&buf)
	err := r.Print(s)
	if err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	output := buf.String()

	// Check for table-like structure with column headers
	if !strings.Contains(output, "PRIORITY") {
		t.Error("expected PRIORITY column header in output")
	}
	if !strings.Contains(output, "LOCATION") {
		t.Error("expected LOCATION column header in output")
	}
	if !strings.Contains(output, "DESCRIPTION") {
		t.Error("expected DESCRIPTION column header in output")
	}
}

// ACCEPTANCE CRITERIA 2:
// "High-priority findings are visually distinct from low-priority findings"
func TestPrettyPrint_HighPriorityDistinct(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewPretty(&buf)
	err := r.Print(s)
	if err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	output := buf.String()

	// High priority findings should appear
	if !strings.Contains(output, "HIGH") {
		t.Error("expected HIGH priority finding in output")
	}
}

// ACCEPTANCE CRITERIA 3:
// "Each finding includes the file path, line range, and a plain-language description"
func TestPrettyPrint_IncludesLocationAndDescription(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewPretty(&buf)
	err := r.Print(s)
	if err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "auth/auth.go") {
		t.Error("expected file path in output")
	}
	if !strings.Contains(output, "Hardcoded credentials") {
		t.Error("expected finding description in output")
	}
}

// ACCEPTANCE CRITERIA 4:
// "Each finding includes an actionable recommendation for how to fix the issue"
func TestPrettyPrint_IncludesRemediation(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewPretty(&buf)
	err := r.Print(s)
	if err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "environment variables") {
		t.Error("expected remediation suggestion in output")
	}
}

// ACCEPTANCE CRITERIA 5:
// "The default output format is the human-readable table; I can opt into JSON output when needed"
func TestJSONPrint_OutputIsValidJSON(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewJSON(&buf)
	err := r.Print(s)
	if err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	output := buf.String()

	// Basic JSON validation: starts with { and ends with }
	trimmed := strings.TrimSpace(output)
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		t.Errorf("expected JSON output, got: %s", trimmed[:min(100, len(trimmed))])
	}

	// Check for score data
	if !strings.Contains(output, `"total"`) {
		t.Error("expected total score in JSON output")
	}
	if !strings.Contains(output, `"findings"`) {
		t.Error("expected findings in JSON output")
	}
}

// Verify findings are ordered by priority (HIGH first)
func TestPrettyPrint_FindingsOrderedByPriority(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewPretty(&buf)
	err := r.Print(s)
	if err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	output := buf.String()

	highIdx := strings.Index(output, "HIGH")
	mediumIdx := strings.Index(output, "MEDIUM")
	lowIdx := strings.Index(output, "LOW")

	if highIdx == -1 || mediumIdx == -1 || lowIdx == -1 {
		t.Fatal("expected all priority levels in output")
	}

	if highIdx > mediumIdx {
		t.Error("expected HIGH priority before MEDIUM")
	}
	if mediumIdx > lowIdx {
		t.Error("expected MEDIUM priority before LOW")
	}
}

// Verify the score summary is shown
func TestPrettyPrint_ShowsScoreSummary(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewPretty(&buf)
	err := r.Print(s)
	if err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "72.5") {
		t.Error("expected total score in output")
	}
}

// Verify pillar breakdown is shown
func TestPrettyPrint_ShowsPillarBreakdown(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewPretty(&buf)
	err := r.Print(s)
	if err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Static") && !strings.Contains(output, "static") {
		t.Error("expected Static pillar in output")
	}
	if !strings.Contains(output, "Architecture") && !strings.Contains(output, "architecture") {
		t.Error("expected Architecture pillar in output")
	}
}

// Verify JSON output can be written to file
func TestJSONPrint_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/report.json"

	s := sampleScore()
	err := reporter.WriteJSON(s, path)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if !strings.Contains(string(data), `"total"`) {
		t.Error("expected JSON content in file")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
