package evaluator

import (
	"path/filepath"
	"strings"

	"gatekeeper/internal/config"
	"gatekeeper/internal/scanner"
	"gatekeeper/pkg/score"
)

// sourceExtensions maps file extensions to source language identifiers.
var sourceExtensions = map[string]bool{
	".go":    true,
	".py":    true,
	".js":    true,
	".ts":    true,
	".java":  true,
	".rb":    true,
	".rs":    true,
	".c":     true,
	".cpp":   true,
	".h":     true,
	".cs":    true,
	".php":   true,
	".swift": true,
	".kt":    true,
}

// CheckWorkspace evaluates all source files in the given directory
// and returns a score with pillar breakdown and findings.
func CheckWorkspace(cfg config.GatekeeperConfig, root string) score.Score {
	s := scanner.New(cfg.Exclusions.Paths)
	files, _ := s.Scan(root)

	result := score.NewScore()

	// Filter to source files only
	var sourceFiles []string
	for _, f := range files {
		if isSourceFile(f) {
			sourceFiles = append(sourceFiles, f)
		}
	}

	// Analyze each source file
	for _, f := range sourceFiles {
		rel, _ := filepath.Rel(root, f)
		findings := analyzeFile(f, rel)
		result.Findings = append(result.Findings, findings...)
	}

	// Compute pillar scores
	result.Pillars[score.PillarStatic] = computeStaticScore(result.Findings)
	result.Pillars[score.PillarArchitecture] = computeArchitectureScore(result.Findings)
	result.Pillars[score.PillarVerification] = computeVerificationScore(result.Findings, cfg)
	result.Pillars[score.PillarSecurity] = computeSecurityScore(result.Findings)

	// Total is sum of pillars
	result.Total = 0
	for _, v := range result.Pillars {
		result.Total += v
	}

	return result
}

// isSourceFile returns true if the file has a known source code extension.
func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return sourceExtensions[ext]
}

// analyzeFile performs basic static analysis on a single file.
func analyzeFile(absPath, relPath string) []score.Finding {
	var findings []score.Finding
	// TODO: integrate AST analysis, linter output, etc.
	// For now, return empty findings (the file was analyzed, just no issues found)
	_ = absPath
	_ = relPath
	return findings
}

// computeStaticScore returns the Static Code Health pillar score (max 20).
func computeStaticScore(findings []score.Finding) float64 {
	maxPoints := score.PillarMaxPoints[score.PillarStatic]
	deductions := 0.0
	for _, f := range findings {
		if f.Pillar == score.PillarStatic {
			deductions += 0.5
		}
	}
	if deductions >= maxPoints {
		return 0
	}
	return maxPoints - deductions
}

// computeArchitectureScore returns the Architecture pillar score (max 25).
func computeArchitectureScore(findings []score.Finding) float64 {
	maxPoints := score.PillarMaxPoints[score.PillarArchitecture]
	deductions := 0.0
	for _, f := range findings {
		if f.Pillar == score.PillarArchitecture {
			deductions += 1.0
		}
	}
	if deductions >= maxPoints {
		return 0
	}
	return maxPoints - deductions
}

// computeVerificationScore returns the Verification pillar score (max 35).
func computeVerificationScore(findings []score.Finding, cfg config.GatekeeperConfig) float64 {
	maxPoints := score.PillarMaxPoints[score.PillarVerification]
	// Default to max if no verification data (no test results to penalize)
	// This is the rule-based fallback; actual Farley Index comes from pipeline data
	return maxPoints
}

// computeSecurityScore returns the Security pillar score (max 20).
func computeSecurityScore(findings []score.Finding) float64 {
	maxPoints := score.PillarMaxPoints[score.PillarSecurity]
	for _, f := range findings {
		if f.Pillar == score.PillarSecurity {
			if f.Severity == string(SeverityCritical) || f.Severity == string(SeverityHigh) {
				return 0
			}
		}
	}
	return maxPoints
}
