package evaluator

import (
	"path/filepath"
	"strings"

	"gatekeeper/internal/config"
	"gatekeeper/pkg/score"
)

// CheckDiff evaluates a list of changed files and returns a quality score.
func CheckDiff(cfg config.GatekeeperConfig, changedFiles []string) score.Score {
	result := score.NewScore()

	for _, f := range changedFiles {
		absPath := f
		if !filepath.IsAbs(f) {
			absPath = filepath.Join(".", f)
		}

		// Skip non-code files
		if !isSourceFile(absPath) {
			continue
		}

		findings := analyzeFile(absPath, f)
		result.Findings = append(result.Findings, findings...)
	}

	// Compute pillar scores
	result.Pillars[score.PillarStatic] = computeStaticScore(result.Findings)
	result.Pillars[score.PillarArchitecture] = computeArchitectureScore(result.Findings)
	result.Pillars[score.PillarVerification] = computeVerificationScore(result.Findings, cfg)
	result.Pillars[score.PillarSecurity] = computeSecurityScore(result.Findings)

	// Apply mutation penalty
	if cfg.Pillars.Verification.MutationCoverageFloor > 0 {
		verificationScore := result.Pillars[score.PillarVerification]
		// Default: no mutation data, no penalty
		result.Pillars[score.PillarVerification] = verificationScore
	}

	// Total is sum of pillars
	result.Total = 0
	for _, v := range result.Pillars {
		result.Total += v
	}

	return result
}

// GetChangedFilesFromGit returns changed files between two git refs.
func GetChangedFilesFromGit(repoPath, base, target string) ([]string, error) {
	// This is a placeholder; actual git integration is in internal/git
	_ = repoPath
	_ = base
	_ = target
	return nil, nil
}

// --- Helper functions for diff analysis ---

// analyzeDiffContent analyzes a diff string for quality issues.
func analyzeDiffContent(content string) []score.Finding {
	var findings []score.Finding
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// Check for secrets in added lines
		if strings.HasPrefix(line, "+") {
			lineContent := strings.TrimPrefix(line, "+")
			for _, re := range secretPatterns {
				if re.MatchString(lineContent) {
					findings = append(findings, score.Finding{
						Priority:    "HIGH",
						Pillar:      score.PillarSecurity,
						Description: "Potential hardcoded secret in diff",
						Remediation: "Move this secret to an environment variable or secrets manager",
						Severity:    string(SeverityHigh),
					})
					break
				}
			}
		}
	}

	return findings
}
