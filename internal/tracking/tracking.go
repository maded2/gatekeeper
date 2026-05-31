package tracking

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EvaluationRecord stores metadata about a single evaluation.
type EvaluationRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	Score        float64   `json:"score"`
	Threshold    float64   `json:"threshold"`
	Passed       bool      `json:"passed"`
	Branch       string    `json:"branch,omitempty"`
	Author       string    `json:"author,omitempty"`
	Repository   string    `json:"repository"`
	Mode         string    `json:"mode"` // "local" or "ci"
	IssueCount   int       `json:"issue_count"`
	CriticalCount int      `json:"critical_count"`
	QualityScore float64   `json:"quality_score"`
	CoverageScore float64  `json:"coverage_score"`
	DeployScore  float64   `json:"deploy_score"`
}

// Tracker records and queries evaluation results.
type Tracker struct {
	dataDir string
	enabled bool
}

// NewTracker creates a new evaluation tracker.
func NewTracker(dataDir string, enabled bool) *Tracker {
	return &Tracker{
		dataDir: dataDir,
		enabled: enabled,
	}
}

// Record saves an evaluation result.
func (t *Tracker) Record(record EvaluationRecord) error {
	if !t.enabled {
		return nil
	}

	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	// Ensure data directory exists
	if err := os.MkdirAll(t.dataDir, 0755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	// Append to evaluations.jsonl
	filePath := filepath.Join(t.dataDir, "evaluations.jsonl")
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening evaluations file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(record); err != nil {
		return fmt.Errorf("encoding evaluation record: %w", err)
	}

	return nil
}

// Records returns all recorded evaluations.
func (t *Tracker) Records() ([]EvaluationRecord, error) {
	var records []EvaluationRecord

	filePath := filepath.Join(t.dataDir, "evaluations.jsonl")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return records, nil
		}
		return nil, fmt.Errorf("reading evaluations file: %w", err)
	}

	lines := splitLines(string(data))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var record EvaluationRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue // skip malformed lines
		}
		records = append(records, record)
	}

	return records, nil
}

// TrendReport produces a summary of quality trends.
type TrendReport struct {
	AverageScore    float64 `json:"average_score"`
	TotalEvaluations int    `json:"total_evaluations"`
	PassRate        float64 `json:"pass_rate"`
	Improving       bool    `json:"improving"`
	Period          string  `json:"period"`
}

// GenerateTrendReport produces a trend report for the given time period.
func (t *Tracker) GenerateTrendReport(period string) (TrendReport, error) {
	records, err := t.Records()
	if err != nil {
		return TrendReport{}, err
	}

	if len(records) == 0 {
		return TrendReport{Period: period}, nil
	}

	var totalScore float64
	var passedCount int

	for _, r := range records {
		totalScore += r.Score
		if r.Passed {
			passedCount++
		}
	}

	avgScore := totalScore / float64(len(records))
	passRate := float64(passedCount) / float64(len(records)) * 100

	// Determine if improving (compare first half vs second half)
	improving := false
	if len(records) >= 4 {
		mid := len(records) / 2
		var firstHalf, secondHalf float64
		for i := 0; i < mid; i++ {
			firstHalf += records[i].Score
		}
		for i := mid; i < len(records); i++ {
			secondHalf += records[i].Score
		}
		firstAvg := firstHalf / float64(mid)
		secondAvg := secondHalf / float64(len(records)-mid)
		improving = secondAvg > firstAvg
	}

	return TrendReport{
		AverageScore:     roundTo1(avgScore),
		TotalEvaluations: len(records),
		PassRate:         roundTo1(passRate),
		Improving:        improving,
		Period:           period,
	}, nil
}

// RejectionReport produces a breakdown of rejection reasons.
type RejectionReport struct {
	TotalRejections   int             `json:"total_rejections"`
	CategoryBreakdown map[string]int  `json:"category_breakdown"`
	TopCategories     []CategoryStat  `json:"top_categories"`
}

// CategoryStat holds statistics for a single rejection category.
type CategoryStat struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Percentage float64 `json:"percentage"`
}

// GenerateRejectionReport produces a rejection category breakdown.
func (t *Tracker) GenerateRejectionReport() (RejectionReport, error) {
	records, err := t.Records()
	if err != nil {
		return RejectionReport{}, err
	}

	var totalRejections int
	categoryCounts := make(map[string]int)

	for _, r := range records {
		if !r.Passed {
			totalRejections++
			// Count issues by category (simplified - in real implementation, this would come from issue data)
			if r.QualityScore < 6.0 {
				categoryCounts["code_quality"]++
			}
			if r.CoverageScore < 6.0 {
				categoryCounts["test_coverage"]++
			}
			if r.DeployScore < 6.0 {
				categoryCounts["deployability"]++
			}
		}
	}

	// Sort categories by count
	var topCategories []CategoryStat
	for cat, count := range categoryCounts {
		pct := float64(count) / float64(totalRejections) * 100
		topCategories = append(topCategories, CategoryStat{
			Category:   cat,
			Count:      count,
			Percentage: roundTo1(pct),
		})
	}

	return RejectionReport{
		TotalRejections:   totalRejections,
		CategoryBreakdown: categoryCounts,
		TopCategories:     topCategories,
	}, nil
}

func splitLines(s string) []string {
	var lines []string
	var current string
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func roundTo1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
