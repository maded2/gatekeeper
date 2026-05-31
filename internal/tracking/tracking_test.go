package tracking_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eddie/gatekeeper/internal/tracking"
)

// =============================================================================
// Story 8.1: Quality Trend Reporting
// Acceptance Tests
// =============================================================================

func TestTracking_RecordsEvaluation(t *testing.T) {
	// Given a tracker
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, true)

	// When an evaluation is recorded
	record := tracking.EvaluationRecord{
		Score:        7.5,
		Threshold:    6.0,
		Passed:       true,
		Branch:       "main",
		Author:       "dev1",
		Repository:   "test-repo",
		Mode:         "local",
		QualityScore: 8.0,
		CoverageScore: 7.0,
		DeployScore:  7.5,
	}
	err := tracker.Record(record)
	if err != nil {
		t.Fatal(err)
	}

	// Then the record is saved
	records, err := tracker.Records()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
}

func TestTracking_TrendReport(t *testing.T) {
	// Given a tracker with multiple evaluations
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, true)

	// Record evaluations with improving scores
	scores := []float64{5.0, 5.5, 6.0, 6.5, 7.0, 7.5}
	for _, score := range scores {
		err := tracker.Record(tracking.EvaluationRecord{
			Score:        score,
			Threshold:    6.0,
			Passed:       score >= 6.0,
			Repository:   "test-repo",
			Mode:         "local",
			QualityScore: score,
			CoverageScore: score,
			DeployScore:  score,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// When a trend report is generated
	report, err := tracker.GenerateTrendReport("last_30_days")
	if err != nil {
		t.Fatal(err)
	}

	// Then the report shows trends
	if report.TotalEvaluations != 6 {
		t.Errorf("expected 6 evaluations, got %d", report.TotalEvaluations)
	}
	if !report.Improving {
		t.Error("expected improving trend")
	}
	if report.PassRate < 50.0 {
		t.Errorf("expected pass rate >= 50%%, got %.1f%%", report.PassRate)
	}
}

func TestTracking_FilterByBranch(t *testing.T) {
	// Given evaluations on different branches
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, true)

	err := tracker.Record(tracking.EvaluationRecord{Score: 7.0, Branch: "main", Repository: "test"})
	if err != nil {
		t.Fatal(err)
	}
	err = tracker.Record(tracking.EvaluationRecord{Score: 5.0, Branch: "feature", Repository: "test"})
	if err != nil {
		t.Fatal(err)
	}

	// When records are retrieved
	records, err := tracker.Records()
	if err != nil {
		t.Fatal(err)
	}

	// Then they can be filtered by branch
	mainCount := 0
	featureCount := 0
	for _, r := range records {
		if r.Branch == "main" {
			mainCount++
		}
		if r.Branch == "feature" {
			featureCount++
		}
	}
	if mainCount != 1 {
		t.Errorf("expected 1 main record, got %d", mainCount)
	}
	if featureCount != 1 {
		t.Errorf("expected 1 feature record, got %d", featureCount)
	}
}

func TestTracking_MachineReadableFormat(t *testing.T) {
	// Given a tracker with records
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, true)

	err := tracker.Record(tracking.EvaluationRecord{Score: 7.0, Repository: "test"})
	if err != nil {
		t.Fatal(err)
	}

	// When records are retrieved
	records, err := tracker.Records()
	if err != nil {
		t.Fatal(err)
	}

	// Then they are machine-readable (JSON)
	data, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("expected records to be JSON-serializable, got error: %v", err)
	}
	if !strings.Contains(string(data), "\"score\"") {
		t.Error("expected JSON format")
	}
}

// =============================================================================
// Story 8.2: Rejection Category Analytics
// Acceptance Tests
// =============================================================================

func TestRejectionReport_CategoryBreakdown(t *testing.T) {
	// Given evaluations with different failure categories
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, true)

	// Record rejections with different failing pillars
	err := tracker.Record(tracking.EvaluationRecord{
		Score: 4.0, Passed: false, Repository: "test",
		QualityScore: 3.0, CoverageScore: 7.0, DeployScore: 7.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = tracker.Record(tracking.EvaluationRecord{
		Score: 4.0, Passed: false, Repository: "test",
		QualityScore: 7.0, CoverageScore: 3.0, DeployScore: 7.0,
	})
	if err != nil {
		t.Fatal(err)
	}

	// When a rejection report is generated
	report, err := tracker.GenerateRejectionReport()
	if err != nil {
		t.Fatal(err)
	}

	// Then categories are tracked
	if report.TotalRejections != 2 {
		t.Errorf("expected 2 rejections, got %d", report.TotalRejections)
	}
	if report.CategoryBreakdown["code_quality"] != 1 {
		t.Error("expected 1 code_quality rejection")
	}
	if report.CategoryBreakdown["test_coverage"] != 1 {
		t.Error("expected 1 test_coverage rejection")
	}
}

func TestRejectionReport_TopCategories(t *testing.T) {
	// Given evaluations with a dominant rejection category
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, true)

	// Record 3 code quality rejections, 1 coverage rejection
	for i := 0; i < 3; i++ {
		err := tracker.Record(tracking.EvaluationRecord{
			Score: 4.0, Passed: false, Repository: "test",
			QualityScore: 3.0, CoverageScore: 7.0, DeployScore: 7.0,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	err := tracker.Record(tracking.EvaluationRecord{
		Score: 4.0, Passed: false, Repository: "test",
		QualityScore: 7.0, CoverageScore: 3.0, DeployScore: 7.0,
	})
	if err != nil {
		t.Fatal(err)
	}

	// When a rejection report is generated
	report, err := tracker.GenerateRejectionReport()
	if err != nil {
		t.Fatal(err)
	}

	// Then code_quality is the top category
	topCategory := ""
	topCount := 0
	for _, cat := range report.TopCategories {
		if cat.Count > topCount {
			topCount = cat.Count
			topCategory = cat.Category
		}
	}
	if topCategory != "code_quality" {
		t.Errorf("expected code_quality as top category, got %s", topCategory)
	}
}

// =============================================================================
// Story 8.3: Local Adoption Tracking
// Acceptance Tests
// =============================================================================

func TestTracking_LocalCheckAdoption(t *testing.T) {
	// Given a tracker with local evaluations
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, true)

	// Record local checks from different developers
	for _, author := range []string{"dev1", "dev2", "dev1", "dev1"} {
		err := tracker.Record(tracking.EvaluationRecord{
			Score:      7.0,
			Passed:     true,
			Author:     author,
			Repository: "test-repo",
			Mode:       "local",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// When records are retrieved
	records, err := tracker.Records()
	if err != nil {
		t.Fatal(err)
	}

	// Then adoption can be tracked per developer
	devCounts := make(map[string]int)
	for _, r := range records {
		devCounts[r.Author]++
	}
	if devCounts["dev1"] != 3 {
		t.Errorf("expected dev1 to have 3 checks, got %d", devCounts["dev1"])
	}
	if devCounts["dev2"] != 1 {
		t.Errorf("expected dev2 to have 1 check, got %d", devCounts["dev2"])
	}
}

func TestTracking_NoCodeContentTracked(t *testing.T) {
	// Given a tracker
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, true)

	// When a record is saved
	err := tracker.Record(tracking.EvaluationRecord{
		Score:      7.0,
		Repository: "test-repo",
		Mode:       "local",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Then only metadata is tracked (no code content)
	filePath := filepath.Join(tmpDir, "evaluations.jsonl")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "func ") || strings.Contains(content, "package ") {
		t.Error("expected no code content in tracking data")
	}
}

func TestTracking_OptIn(t *testing.T) {
	// Given a disabled tracker
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, false)

	// When a record is saved
	err := tracker.Record(tracking.EvaluationRecord{Score: 7.0})
	if err != nil {
		t.Fatal(err)
	}

	// Then no data is written
	records, err := tracker.Records()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Errorf("expected no records when disabled, got %d", len(records))
	}
}

func TestTracking_TimestampRecorded(t *testing.T) {
	// Given a tracker
	tmpDir := t.TempDir()
	tracker := tracking.NewTracker(tmpDir, true)

	// When a record is saved
	before := time.Now()
	err := tracker.Record(tracking.EvaluationRecord{Score: 7.0})
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now()

	// Then a timestamp is recorded
	records, _ := tracker.Records()
	if len(records) != 1 {
		t.Fatal("expected 1 record")
	}
	ts := records[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("expected timestamp between before and after, got %v", ts)
	}
}
