package evaluator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gatekeeper/internal/config"
	"gatekeeper/internal/scanner"
	"gatekeeper/pkg/score"
)

// secretPatterns matches common hardcoded secret patterns.
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[=:]\s*["'][^"']{8,}["']`),
	regexp.MustCompile(`(?i)(secret|password|passwd|pwd)\s*[=:]\s*["'][^"']{4,}["']`),
	regexp.MustCompile(`(?i)sk-[a-zA-Z0-9]{10,}`),
}

// poorNames matches single-letter or meaningless variable names.
var poorNames = regexp.MustCompile(`\b[a]\b`)

// maxNestingDepth is the threshold for cognitive complexity warnings.
const maxNestingDepth = 3

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

	content, err := os.ReadFile(absPath)
	if err != nil {
		return findings
	}

	lines := strings.Split(string(content), "\n")

	// Check for hardcoded secrets
	findings = append(findings, checkSecrets(relPath, lines)...)

	// Check for cognitive complexity (nesting depth)
	findings = append(findings, checkNestingDepth(relPath, lines)...)

	return findings
}

// checkSecrets scans file lines for hardcoded secret patterns.
func checkSecrets(relPath string, lines []string) []score.Finding {
	var findings []score.Finding
	for i, line := range lines {
		for _, re := range secretPatterns {
			if re.MatchString(line) {
				findings = append(findings, score.Finding{
					Priority:    "HIGH",
					Pillar:      score.PillarSecurity,
					Location:    relPath,
					LineStart:   i + 1,
					LineEnd:     i + 1,
					Description: "Potential hardcoded secret detected",
					Remediation: "Move this secret to an environment variable or a secure secrets manager",
					Severity:    string(SeverityHigh),
				})
				break
			}
		}
	}
	return findings
}

// checkNestingDepth detects functions with nesting deeper than maxNestingDepth.
func checkNestingDepth(relPath string, lines []string) []score.Finding {
	var findings []score.Finding
	var funcStart, maxDepth, depth int
	inFunc := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect function start (simple heuristic)
		if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "def ") {
			if inFunc && maxDepth > maxNestingDepth {
				findings = append(findings, score.Finding{
					Priority:    "MEDIUM",
					Pillar:      score.PillarArchitecture,
					Location:    relPath,
					LineStart:   funcStart + 1,
					LineEnd:     i,
					Description: fmt.Sprintf("Function has nesting depth of %d (max recommended: %d)", maxDepth, maxNestingDepth),
					Remediation: "Extract nested logic into separate helper functions to reduce complexity",
				})
			}
			inFunc = true
			funcStart = i
			maxDepth = 0
			depth = 0
			continue
		}

		if !inFunc {
			continue
		}

		// Track nesting via braces/indentation
		for _, ch := range trimmed {
			if ch == '{' {
				depth++
				if depth > maxDepth {
					maxDepth = depth
				}
			} else if ch == '}' {
				if depth > 0 {
					depth--
				}
			}
		}
	}

	// Handle last function
	if inFunc && maxDepth > maxNestingDepth {
		findings = append(findings, score.Finding{
			Priority:    "MEDIUM",
			Pillar:      score.PillarArchitecture,
			Location:    relPath,
			LineStart:   funcStart + 1,
			LineEnd:     len(lines),
			Description: fmt.Sprintf("Function has nesting depth of %d (max recommended: %d)", maxDepth, maxNestingDepth),
			Remediation: "Extract nested logic into separate helper functions to reduce complexity",
		})
	}

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
