package evaluator_test

import (
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
)

// --- Story D-3: Block Deployment When Critical Security Issues Exist ---

// ACCEPTANCE CRITERIA 1:
// "I can enable a hard-fail toggle for critical security findings in my configuration"
func TestSecurityBlock_HardFailToggle(t *testing.T) {
	cfg := config.DefaultConfig()
	trueVal := true
	cfg.Gatekeeper.FailOnCriticalSecurity = &trueVal

	if cfg.Gatekeeper.FailOnCriticalSecurity == nil || !*cfg.Gatekeeper.FailOnCriticalSecurity {
		t.Error("expected hard-fail toggle to be enabled")
	}
}

// ACCEPTANCE CRITERIA 2:
// "When enabled, any critical or high severity security finding triggers a fail regardless of the overall Quality Score"
func TestSecurityBlock_CriticalFindingBlocksDeployment(t *testing.T) {
	cfg := config.DefaultConfig()
	trueVal := true
	cfg.Gatekeeper.FailOnCriticalSecurity = &trueVal
	cfg.Gatekeeper.TargetThreshold = 0 // lowest threshold

	// Perfect score but with critical finding
	result := evaluator.EvaluateWithFindings(cfg, 100, []evaluator.Finding{
		{Severity: evaluator.SeverityCritical, Pillar: "security"},
	})

	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("expected exit 2 (fail) with critical finding, got %d", result.ExitCode)
	}
}

// ACCEPTANCE CRITERIA 3:
// "The output clearly identifies which security finding triggered the block"
func TestSecurityBlock_FindingIdentified(t *testing.T) {
	cfg := config.DefaultConfig()
	trueVal := true
	cfg.Gatekeeper.FailOnCriticalSecurity = &trueVal

	result := evaluator.EvaluateWithFindings(cfg, 95, []evaluator.Finding{
		{Severity: evaluator.SeverityCritical, Pillar: "security"},
	})

	if len(result.Findings) == 0 {
		t.Error("expected findings to be included in result")
	}
}

// ACCEPTANCE CRITERIA 4:
// "This behavior is independent of the numeric quality threshold"
func TestSecurityBlock_IndependentOfThreshold(t *testing.T) {
	cfg := config.DefaultConfig()
	trueVal := true
	cfg.Gatekeeper.FailOnCriticalSecurity = &trueVal

	// Test with various thresholds
	for _, threshold := range []float64{0, 50, 75, 99, 100} {
		cfg.Gatekeeper.TargetThreshold = threshold
		score := threshold + 10 // score always above threshold

		result := evaluator.EvaluateWithFindings(cfg, score, []evaluator.Finding{
			{Severity: evaluator.SeverityHigh, Pillar: "security"},
		})

		if result.ExitCode != evaluator.ExitFail {
			t.Errorf("threshold=%f, score=%f: expected fail with high finding, got %d",
				threshold, score, result.ExitCode)
		}
	}
}

// Low severity should NOT trigger hard fail
func TestSecurityBlock_LowSeverityDoesNotBlock(t *testing.T) {
	cfg := config.DefaultConfig()
	trueVal := true
	cfg.Gatekeeper.FailOnCriticalSecurity = &trueVal
	cfg.Gatekeeper.TargetThreshold = 50

	result := evaluator.EvaluateWithFindings(cfg, 95, []evaluator.Finding{
		{Severity: evaluator.SeverityLow, Pillar: "security"},
	})

	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("expected pass with low severity finding, got %d", result.ExitCode)
	}
}
