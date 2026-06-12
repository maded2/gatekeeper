package evaluator

import (
	"gatekeeper/internal/config"
)

// Exit codes per spec §5.2.
const (
	ExitPass int = iota // 0: score >= threshold
	ExitError           // 1: runtime error
	ExitFail            // 2: score < threshold
)

// Severity represents the severity level of a finding.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
)

// Finding represents a single quality finding.
type Finding struct {
	Severity Severity
	Pillar   string
}

// Result holds the outcome of a quality evaluation.
type Result struct {
	Score    float64
	ExitCode int
	Findings []Finding
}

// Evaluate computes the exit code based on the score and configured threshold.
func Evaluate(cfg config.GatekeeperConfig, score float64) Result {
	return EvaluateWithFindings(cfg, score, nil)
}

// EvaluateWithFindings computes the exit code considering both score threshold
// and optional hard-fail security findings.
func EvaluateWithFindings(cfg config.GatekeeperConfig, score float64, findings []Finding) Result {
	result := Result{
		Score:    score,
		ExitCode: ExitPass,
		Findings: findings,
	}

	// Check threshold
	if score < cfg.Gatekeeper.TargetThreshold {
		result.ExitCode = ExitFail
		return result
	}

	// Check hard-fail on critical/high security findings
	if cfg.Gatekeeper.FailOnCriticalSecurity != nil && *cfg.Gatekeeper.FailOnCriticalSecurity {
		for _, f := range findings {
			if f.Severity == SeverityCritical || f.Severity == SeverityHigh {
				result.ExitCode = ExitFail
				return result
			}
		}
	}

	return result
}
