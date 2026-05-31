package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eddie/gatekeeper/internal/config"
	"github.com/eddie/gatekeeper/internal/evaluator"
	"github.com/eddie/gatekeeper/internal/farleyscore"
)

// Format indicates the output format.
type Format int

const (
	Text Format = iota
	JSON
)

// TextOutput produces human-readable terminal output.
func TextOutput(score farleyscore.ScoreResult, cfg config.Configuration, mode Mode) string {
	var b strings.Builder

	// Header
	if mode == CI {
		fmt.Fprintf(&b, "\n╔══════════════════════════════════════════════════╗\n")
		fmt.Fprintf(&b, "║           GATEKEEPER QUALITY GATE                ║\n")
		fmt.Fprintf(&b, "╚══════════════════════════════════════════════════╝\n\n")
	} else {
		fmt.Fprintf(&b, "\n🛡️  Gatekeeper Quality Check\n\n")
	}

	// Score
	var emoji string
	if score.FarleyScore >= cfg.Threshold {
		emoji = "✅"
	} else {
		emoji = "❌"
	}
	fmt.Fprintf(&b, "%s Farley Score: %.1f / 10.0 (threshold: %.1f)\n\n", emoji, score.FarleyScore, cfg.Threshold)

	// Pillar breakdown
	fmt.Fprintf(&b, "Pillar Breakdown:\n")
	fmt.Fprintf(&b, "  📊 Code Quality:    %.1f (weight: %.0f%%)\n", score.QualityScore, score.QualityWeight*100)
	fmt.Fprintf(&b, "  🧪 Test Coverage:   %.1f (weight: %.0f%%)\n", score.CoverageScore, score.CoverageWeight*100)
	fmt.Fprintf(&b, "  🚀 Deployability:   %.1f (weight: %.0f%%)\n", score.DeployScore, score.DeployWeight*100)
	fmt.Fprintf(&b, "\n")

	// Issues
	if len(score.Issues) > 0 {
		fmt.Fprintf(&b, "Issues Found:\n")
		for _, issue := range score.Issues {
			var sevEmoji string
			switch issue.Severity {
			case evaluator.Critical:
				sevEmoji = "🔴"
			case evaluator.Warning:
				sevEmoji = "🟡"
			case evaluator.Info:
				sevEmoji = "🔵"
			}
			fmt.Fprintf(&b, "  %s [%s] %s\n", sevEmoji, issue.Severity, issue.Title)
			if issue.Line > 0 {
				fmt.Fprintf(&b, "     → %s:%d\n", issue.File, issue.Line)
			}
			fmt.Fprintf(&b, "     %s\n", issue.Detail)
			if issue.Recommendation != "" {
				fmt.Fprintf(&b, "     💡 %s\n", issue.Recommendation)
			}
			fmt.Fprintf(&b, "\n")
		}
	}

	// Footer
	if score.FarleyScore >= cfg.Threshold {
		if mode == CI {
			fmt.Fprintf(&b, "✅ QUALITY GATE PASSED — Merge allowed\n")
		} else {
			fmt.Fprintf(&b, "✅ Code quality meets threshold — safe to push\n")
		}
	} else {
		if mode == CI {
			fmt.Fprintf(&b, "❌ QUALITY GATE FAILED — Merge blocked\n")
			fmt.Fprintf(&b, "   Score %.1f is below threshold %.1f\n", score.FarleyScore, cfg.Threshold)
		} else {
			fmt.Fprintf(&b, "⚠️  Code quality below threshold — consider improvements before pushing\n")
			fmt.Fprintf(&b, "   Score %.1f is below threshold %.1f\n", score.FarleyScore, cfg.Threshold)
		}
	}

	// Large diff warning
	if score.IsLargeDiff {
		fmt.Fprintf(&b, "\n⚠️  %s\n", score.LargeDiffWarn)
	}

	return b.String()
}

// JSONOutput produces machine-readable JSON output.
func JSONOutput(score farleyscore.ScoreResult, cfg config.Configuration) ([]byte, error) {
	type IssueJSON struct {
		File           string `json:"file"`
		Line           int    `json:"line,omitempty"`
		Severity       string `json:"severity"`
		Category       string `json:"category"`
		Title          string `json:"title"`
		Detail         string `json:"detail"`
		Recommendation string `json:"recommendation,omitempty"`
	}

	type OutputJSON struct {
		FarleyScore    float64       `json:"farley_score"`
		Threshold      float64       `json:"threshold"`
		Passed         bool          `json:"passed"`
		QualityScore   float64       `json:"quality_score"`
		CoverageScore  float64       `json:"coverage_score"`
		DeployScore    float64       `json:"deploy_score"`
		QualityWeight  float64       `json:"quality_weight"`
		CoverageWeight float64       `json:"coverage_weight"`
		DeployWeight   float64       `json:"deploy_weight"`
		IsLargeDiff    bool          `json:"is_large_diff,omitempty"`
		Issues         []IssueJSON   `json:"issues"`
	}

	issues := make([]IssueJSON, len(score.Issues))
	for i, issue := range score.Issues {
		issues[i] = IssueJSON{
			File:           issue.File,
			Line:           issue.Line,
			Severity:       issue.Severity.String(),
			Category:       issue.Category.String(),
			Title:          issue.Title,
			Detail:         issue.Detail,
			Recommendation: issue.Recommendation,
		}
	}

	output := OutputJSON{
		FarleyScore:    score.FarleyScore,
		Threshold:      cfg.Threshold,
		Passed:         score.FarleyScore >= cfg.Threshold,
		QualityScore:   score.QualityScore,
		CoverageScore:  score.CoverageScore,
		DeployScore:    score.DeployScore,
		QualityWeight:  score.QualityWeight,
		CoverageWeight: score.CoverageWeight,
		DeployWeight:   score.DeployWeight,
		IsLargeDiff:    score.IsLargeDiff,
		Issues:         issues,
	}

	return json.MarshalIndent(output, "", "  ")
}

// CIStepSummary produces a concise summary for CI pipeline output.
func CIStepSummary(score farleyscore.ScoreResult, cfg config.Configuration) string {
	var b strings.Builder

	status := "PASS"
	if score.FarleyScore < cfg.Threshold {
		status = "FAIL"
	}

	fmt.Fprintf(&b, "Gatekeeper: %s (score: %.1f, threshold: %.1f)\n", status, score.FarleyScore, cfg.Threshold)
	fmt.Fprintf(&b, "  Quality: %.1f | Coverage: %.1f | Deploy: %.1f\n", score.QualityScore, score.CoverageScore, score.DeployScore)

	if len(score.Issues) > 0 {
		fmt.Fprintf(&b, "  Issues: %d (%d critical, %d warning, %d info)\n",
			len(score.Issues), countSeverity(score.Issues, evaluator.Critical),
			countSeverity(score.Issues, evaluator.Warning), countSeverity(score.Issues, evaluator.Info))
	}

	return b.String()
}

// MarkdownPRComment produces a markdown-formatted PR comment.
func MarkdownPRComment(score farleyscore.ScoreResult, cfg config.Configuration) string {
	var b strings.Builder

	passed := score.FarleyScore >= cfg.Threshold
	if passed {
		fmt.Fprintf(&b, "## ✅ Gatekeeper Quality Gate Passed\n\n")
	} else {
		fmt.Fprintf(&b, "## ❌ Gatekeeper Quality Gate Failed\n\n")
	}

	fmt.Fprintf(&b, "**Farley Score:** %.1f / 10.0 (threshold: %.1f)\n\n", score.FarleyScore, cfg.Threshold)

	fmt.Fprintf(&b, "### Pillar Breakdown\n\n")
	fmt.Fprintf(&b, "| Pillar | Score | Weight |\n")
	fmt.Fprintf(&b, "|--------|-------|--------|\n")
	fmt.Fprintf(&b, "| Code Quality | %.1f | %.0f%% |\n", score.QualityScore, score.QualityWeight*100)
	fmt.Fprintf(&b, "| Test Coverage | %.1f | %.0f%% |\n", score.CoverageScore, score.CoverageWeight*100)
	fmt.Fprintf(&b, "| Deployability | %.1f | %.0f%% |\n", score.DeployScore, score.DeployWeight*100)
	fmt.Fprintf(&b, "\n")

	if len(score.Issues) > 0 {
		fmt.Fprintf(&b, "### Issues\n\n")
		for _, issue := range score.Issues {
			var sevIcon string
			switch issue.Severity {
			case evaluator.Critical:
				sevIcon = "🔴"
			case evaluator.Warning:
				sevIcon = "🟡"
			case evaluator.Info:
				sevIcon = "🔵"
			}
			loc := ""
			if issue.Line > 0 {
				loc = fmt.Sprintf(" (`%s:%d`)", issue.File, issue.Line)
			}
			fmt.Fprintf(&b, "- %s **%s**%s — %s\n", sevIcon, issue.Title, loc, issue.Detail)
			if issue.Recommendation != "" {
				fmt.Fprintf(&b, "  - 💡 %s\n", issue.Recommendation)
			}
		}
		fmt.Fprintf(&b, "\n")
	}

	if !passed {
		fmt.Fprintf(&b, "> ⚠️ Score %.1f is below threshold %.1f. Please address the issues above.\n", score.FarleyScore, cfg.Threshold)
	}

	return b.String()
}

func countSeverity(issues []evaluator.Issue, sev evaluator.Severity) int {
	count := 0
	for _, issue := range issues {
		if issue.Severity == sev {
			count++
		}
	}
	return count
}
