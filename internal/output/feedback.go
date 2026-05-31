package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/eddie/gatekeeper/internal/config"
	"github.com/eddie/gatekeeper/internal/evaluator"
	"github.com/eddie/gatekeeper/internal/farleyscore"
)

// FeedbackOutput produces specific, actionable improvement recommendations.
func FeedbackOutput(score farleyscore.ScoreResult, cfg config.Configuration) string {
	if score.FarleyScore >= cfg.Threshold {
		return positiveFeedback(score)
	}
	return rejectionFeedback(score, cfg)
}

func positiveFeedback(score farleyscore.ScoreResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n✅ Quality gate passed with score %.1f/10.0\n", score.FarleyScore)

	if score.FarleyScore >= 8.0 {
		fmt.Fprintf(&b, "🌟 Excellent code quality — your changes meet high standards\n")
	} else {
		fmt.Fprintf(&b, "👍 Code quality meets the required threshold\n")
	}

	return b.String()
}

func rejectionFeedback(score farleyscore.ScoreResult, cfg config.Configuration) string {
	var b strings.Builder

	// Sort issues by severity (critical first)
	issues := make([]evaluator.Issue, len(score.Issues))
	copy(issues, score.Issues)
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Severity < issues[j].Severity // Critical=0 < Warning=1 < Info=2
	})

	fmt.Fprintf(&b, "\n❌ Quality gate failed — score %.1f below threshold %.1f\n\n", score.FarleyScore, cfg.Threshold)
	fmt.Fprintf(&b, "Priority improvements needed:\n\n")

	for i, issue := range issues {
		var sevEmoji string
		switch issue.Severity {
		case evaluator.Critical:
			sevEmoji = "🔴"
		case evaluator.Warning:
			sevEmoji = "🟡"
		case evaluator.Info:
			sevEmoji = "🔵"
		}

		loc := ""
		if issue.Line > 0 {
			loc = fmt.Sprintf(" at %s:%d", issue.File, issue.Line)
		}

		fmt.Fprintf(&b, "%d. %s %s%s\n", i+1, sevEmoji, issue.Title, loc)
		fmt.Fprintf(&b, "   Problem: %s\n", issue.Detail)
		if issue.Recommendation != "" {
			fmt.Fprintf(&b, "   Fix: %s\n", issue.Recommendation)
		}
		fmt.Fprintf(&b, "\n")
	}

	// Borderline warning
	if score.FarleyScore >= cfg.Threshold-0.5 {
		fmt.Fprintf(&b, "⚠️  Borderline score — only %.1f points below threshold.\n", cfg.Threshold-score.FarleyScore)
		fmt.Fprintf(&b, "    Addressing the highest-priority issues above should bring you over the threshold.\n")
	}

	return b.String()
}

// EvidenceOutput produces raw evaluation data for auditability.
func EvidenceOutput(score farleyscore.ScoreResult, cfg config.Configuration, diffDescription string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Gatekeeper Evaluation Evidence ===\n\n")
	fmt.Fprintf(&b, "Configuration:\n")
	fmt.Fprintf(&b, "  Threshold: %.1f\n", cfg.Threshold)
	fmt.Fprintf(&b, "  Quality Weight: %.0f%%\n", score.QualityWeight*100)
	fmt.Fprintf(&b, "  Coverage Weight: %.0f%%\n", score.CoverageWeight*100)
	fmt.Fprintf(&b, "  Deploy Weight: %.0f%%\n", score.DeployWeight*100)
	fmt.Fprintf(&b, "\nScore Breakdown:\n")
	fmt.Fprintf(&b, "  Farley Score: %.1f\n", score.FarleyScore)
	fmt.Fprintf(&b, "  Quality Pillar: %.1f\n", score.QualityScore)
	fmt.Fprintf(&b, "  Coverage Pillar: %.1f\n", score.CoverageScore)
	fmt.Fprintf(&b, "  Deploy Pillar: %.1f\n", score.DeployScore)
	fmt.Fprintf(&b, "\nDiff Analyzed:\n")
	fmt.Fprintf(&b, "  %s\n", diffDescription)
	fmt.Fprintf(&b, "\nIssues (%d total):\n", len(score.Issues))

	for _, issue := range score.Issues {
		fmt.Fprintf(&b, "  - [%s/%s] %s (line %d)\n", issue.Severity, issue.Category, issue.Title, issue.Line)
	}

	return b.String()
}
