package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gatekeeper/pkg/score"
)

// Reporter formats quality scores for output.
type Reporter interface {
	Print(s score.Score) error
}

// PrettyReporter produces human-readable terminal table output.
type PrettyReporter struct {
	w io.Writer
}

// NewPretty creates a PrettyReporter writing to the given writer.
func NewPretty(w io.Writer) *PrettyReporter {
	return &PrettyReporter{w: w}
}

// JSONReporter produces machine-readable JSON output.
type JSONReporter struct {
	w io.Writer
}

// NewJSON creates a JSONReporter writing to the given writer.
func NewJSON(w io.Writer) *JSONReporter {
	return &JSONReporter{w: w}
}

// Print formats the score as a pretty table.
func (r *PrettyReporter) Print(s score.Score) error {
	// Score summary
	fmt.Fprintf(r.w, "\n")
	fmt.Fprintf(r.w, "  Quality Score: %.1f / 100\n", s.Total)
	fmt.Fprintf(r.w, "\n")

	// Pillar breakdown
	fmt.Fprintf(r.w, "  Pillar Breakdown:\n")
	fmt.Fprintf(r.w, "  %-25s %s\n", "Pillar", "Score")
	fmt.Fprintf(r.w, "  %-25s %s\n", strings.Repeat("-", 25), strings.Repeat("-", 8))

	pillars := []struct {
		name  string
		key   string
		max   float64
	}{
		{"Static Code Health", score.PillarStatic, 20},
		{"Engineering Architecture", score.PillarArchitecture, 25},
		{"Dynamic Verification", score.PillarVerification, 35},
		{"Security & Supply Chain", score.PillarSecurity, 20},
	}

	for _, p := range pillars {
		val := s.Pillars[p.key]
		fmt.Fprintf(r.w, "  %-25s %.1f / %.0f\n", p.name, val, p.max)
	}

	fmt.Fprintf(r.w, "\n")

	// Findings table
	if len(s.Findings) == 0 {
		fmt.Fprintf(r.w, "  No findings.\n\n")
		return nil
	}

	// Sort findings by priority
	sorted := make([]score.Finding, len(s.Findings))
	copy(sorted, s.Findings)
	sort.Slice(sorted, func(i, j int) bool {
		return priorityRank(sorted[i].Priority) < priorityRank(sorted[j].Priority)
	})

	// Table header
	fmt.Fprintf(r.w, "  Findings:\n")
	fmt.Fprintf(r.w, "  %-10s %-25s %-6s %s\n", "PRIORITY", "LOCATION", "LINES", "DESCRIPTION")
	fmt.Fprintf(r.w, "  %-10s %-25s %-6s %s\n",
		strings.Repeat("-", 10), strings.Repeat("-", 25), strings.Repeat("-", 6), strings.Repeat("-", 50))

	for _, f := range sorted {
		lines := fmt.Sprintf("%d-%d", f.LineStart, f.LineEnd)
		loc := truncate(f.Location, 25)
		fmt.Fprintf(r.w, "  %-10s %-25s %-6s %s\n",
			priorityLabel(f.Priority), loc, lines, f.Description)
		fmt.Fprintf(r.w, "  %-10s %-25s         → %s\n",
			"", "", f.Remediation)
	}

	fmt.Fprintf(r.w, "\n")
	return nil
}

// Print formats the score as JSON.
func (r *JSONReporter) Print(s score.Score) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	_, err = r.w.Write(data)
	if err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	// Add trailing newline
	fmt.Fprintln(r.w)
	return nil
}

// WriteJSON writes the score as JSON to a file.
func WriteJSON(s score.Score, path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// priorityRank returns a numeric rank for sorting (lower = higher priority).
func priorityRank(p string) int {
	switch strings.ToUpper(p) {
	case "CRITICAL":
		return 0
	case "HIGH":
		return 1
	case "MEDIUM":
		return 2
	case "LOW":
		return 3
	default:
		return 4
	}
}

// priorityLabel returns a display label for the priority.
func priorityLabel(p string) string {
	switch strings.ToUpper(p) {
	case "CRITICAL":
		return "CRITICAL"
	case "HIGH":
		return "HIGH"
	case "MEDIUM":
		return "MEDIUM"
	case "LOW":
		return "LOW"
	default:
		return p
	}
}

// truncate shortens a string to max length, adding "..." if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
