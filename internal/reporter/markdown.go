package reporter

import (
	"fmt"
	"io"
	"sort"

	"gatekeeper/pkg/score"
)

// MarkdownReporter produces GitHub/Markdown formatted output.
type MarkdownReporter struct {
	w io.Writer
}

// NewMarkdown creates a MarkdownReporter writing to the given writer.
func NewMarkdown(w io.Writer) *MarkdownReporter {
	return &MarkdownReporter{w: w}
}

// Print formats the score as a markdown summary.
func (r *MarkdownReporter) Print(s score.Score) error {
	fmt.Fprintf(r.w, "## Gatekeeper Quality Report\n\n")
	fmt.Fprintf(r.w, "**Quality Score: %.1f / 100**\n\n", s.Total)

	// Pillar breakdown
	fmt.Fprintf(r.w, "### Pillar Breakdown\n\n")
	fmt.Fprintf(r.w, "| Pillar | Score |\n")
	fmt.Fprintf(r.w, "| --- | --- |\n")

	pillars := []struct {
		name string
		key  string
		max  float64
	}{
		{"Static Code Health", score.PillarStatic, 20},
		{"Engineering Architecture", score.PillarArchitecture, 25},
		{"Dynamic Verification", score.PillarVerification, 35},
		{"Security & Supply Chain", score.PillarSecurity, 20},
	}

	for _, p := range pillars {
		val := s.Pillars[p.key]
		fmt.Fprintf(r.w, "| %s | %.1f / %.0f |\n", p.name, val, p.max)
	}

	// Findings
	if len(s.Findings) == 0 {
		fmt.Fprintf(r.w, "\nNo findings.\n")
		return nil
	}

	fmt.Fprintf(r.w, "\n### Findings\n\n")
	fmt.Fprintf(r.w, "| Priority | Location | Lines | Description |\n")
	fmt.Fprintf(r.w, "| --- | --- | --- | --- |\n")

	// Sort by priority
	sorted := make([]score.Finding, len(s.Findings))
	copy(sorted, s.Findings)
	sort.Slice(sorted, func(i, j int) bool {
		return priorityRank(sorted[i].Priority) < priorityRank(sorted[j].Priority)
	})

	for _, f := range sorted {
		lines := fmt.Sprintf("%d-%d", f.LineStart, f.LineEnd)
		fmt.Fprintf(r.w, "| %s | `%s` | %s | %s |\n",
			f.Priority, f.Location, lines, f.Description)
	}

	// Remediations
	fmt.Fprintf(r.w, "\n### Remediations\n\n")
	for _, f := range sorted {
		if f.Remediation != "" {
			fmt.Fprintf(r.w, "- **%s** (%s): %s\n", f.Priority, f.Location, f.Remediation)
		}
	}

	fmt.Fprintf(r.w, "\n")
	return nil
}
