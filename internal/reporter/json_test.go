package reporter_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gatekeeper/internal/reporter"
	"gatekeeper/pkg/score"
)

// --- Story D-2: Output Machine-Readable Results for Pipeline Logging ---

// ACCEPTANCE CRITERIA 1:
// "I can request JSON output via a command-line flag"
// (Verified by cmd/check.go --format=json flag)

// ACCEPTANCE CRITERIA 2:
// "The JSON includes the overall score, per-pillar scores, and all findings with locations"
func TestJSON_OutputStructure(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewJSON(&buf)
	if err := r.Print(s); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	// Parse as JSON to verify structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Check required fields
	if _, ok := parsed["total"]; !ok {
		t.Error("expected 'total' field in JSON output")
	}
	if _, ok := parsed["pillars"]; !ok {
		t.Error("expected 'pillars' field in JSON output")
	}
	if _, ok := parsed["findings"]; !ok {
		t.Error("expected 'findings' field in JSON output")
	}
}

// ACCEPTANCE CRITERIA 5:
// "The JSON output includes a timestamp for correlation with pipeline runs"
func TestJSON_IncludesTimestamp(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewJSON(&buf)
	if err := r.Print(s); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "timestamp") {
		t.Error("expected 'timestamp' field in JSON output")
	}
}

// Verify JSON findings have all required fields
func TestJSON_FindingsStructure(t *testing.T) {
	var buf bytes.Buffer
	s := sampleScore()

	r := reporter.NewJSON(&buf)
	if err := r.Print(s); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	findings := parsed["findings"].([]interface{})
	if len(findings) == 0 {
		t.Fatal("expected findings in output")
	}

	first := findings[0].(map[string]interface{})
	requiredFields := []string{"priority", "pillar", "location", "line_start", "line_end", "description", "remediation"}
	for _, field := range requiredFields {
		if _, ok := first[field]; !ok {
			t.Errorf("expected finding field %q", field)
		}
	}
}

// Verify JSON output is stable (same input → same output)
func TestJSON_StableOutput(t *testing.T) {
	s := score.Score{
		Total:     85.5,
		Pillars: map[string]float64{
			score.PillarStatic:       18,
			score.PillarArchitecture: 22,
			score.PillarVerification: 30,
			score.PillarSecurity:     15.5,
		},
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Findings: []score.Finding{
			{
				Priority:    "HIGH",
				Pillar:      score.PillarStatic,
				Location:    "main.go",
				LineStart:   1,
				LineEnd:     10,
				Description: "test finding",
				Remediation: "test fix",
			},
		},
	}

	// Generate output twice
	var buf1, buf2 bytes.Buffer
	reporter.NewJSON(&buf1).Print(s)
	reporter.NewJSON(&buf2).Print(s)

	if buf1.String() != buf2.String() {
		t.Error("expected stable JSON output for same input")
	}
}
