package evaluator_test

import (
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
)

// --- Story A-3: Set a Custom Quality Score Threshold ---

// ACCEPTANCE CRITERIA 1:
// "I can set a numeric threshold (0–100) in my configuration"
func TestThreshold_AcceptsValidRange(t *testing.T) {
	cfg := config.DefaultConfig()

	// Should accept 0
	cfg.Gatekeeper.TargetThreshold = 0
	result := evaluator.Evaluate(cfg, 0)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("threshold=0, score=0: expected pass, got exit code %d", result.ExitCode)
	}

	// Should accept 100
	cfg.Gatekeeper.TargetThreshold = 100
	result = evaluator.Evaluate(cfg, 100)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("threshold=100, score=100: expected pass, got exit code %d", result.ExitCode)
	}

	// Should accept 50
	cfg.Gatekeeper.TargetThreshold = 50
	result = evaluator.Evaluate(cfg, 75)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("threshold=50, score=75: expected pass, got exit code %d", result.ExitCode)
	}
}

// ACCEPTANCE CRITERIA 2:
// "Scores at or above the threshold result in a pass signal"
func TestThreshold_AtThresholdIsPass(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.TargetThreshold = 80

	result := evaluator.Evaluate(cfg, 80)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("score=80 at threshold=80: expected pass (exit 0), got exit code %d", result.ExitCode)
	}

	result = evaluator.Evaluate(cfg, 81)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("score=81 above threshold=80: expected pass (exit 0), got exit code %d", result.ExitCode)
	}
}

// ACCEPTANCE CRITERIA 3:
// "Scores below the threshold result in a fail signal"
func TestThreshold_BelowThresholdIsFail(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.TargetThreshold = 75

	result := evaluator.Evaluate(cfg, 74)
	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("score=74 below threshold=75: expected fail (exit 2), got exit code %d", result.ExitCode)
	}

	result = evaluator.Evaluate(cfg, 0)
	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("score=0 below threshold=75: expected fail (exit 2), got exit code %d", result.ExitCode)
	}
}

// ACCEPTANCE CRITERIA 4:
// "The default threshold is 75 if I do not configure one"
func TestThreshold_DefaultIs75(t *testing.T) {
	cfg := config.DefaultConfig()

	// Score exactly at default threshold should pass
	result := evaluator.Evaluate(cfg, 75)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("score=75 with default config: expected pass, got exit code %d", result.ExitCode)
	}

	// Score just below default threshold should fail
	result = evaluator.Evaluate(cfg, 74)
	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("score=74 with default config: expected fail, got exit code %d", result.ExitCode)
	}
}

// ACCEPTANCE CRITERIA 5:
// "I can enable a hard fail for any critical or high severity security finding, regardless of overall score"
func TestThreshold_HardFailOnCriticalSecurity(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.TargetThreshold = 50 // low threshold so score alone would pass
	trueVal := true
	cfg.Gatekeeper.FailOnCriticalSecurity = &trueVal

	// High score but with critical security finding should still fail
	result := evaluator.EvaluateWithFindings(cfg, 95, []evaluator.Finding{
		{
			Severity: evaluator.SeverityCritical,
			Pillar:   "security",
		},
	})
	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("score=95 with critical finding: expected fail (exit 2), got exit code %d", result.ExitCode)
	}

	// High severity also triggers hard fail
	result = evaluator.EvaluateWithFindings(cfg, 95, []evaluator.Finding{
		{
			Severity: evaluator.SeverityHigh,
			Pillar:   "security",
		},
	})
	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("score=95 with high finding: expected fail (exit 2), got exit code %d", result.ExitCode)
	}
}

// Hard fail only applies when the toggle is enabled
func TestThreshold_HardFailDisabledByDefault(t *testing.T) {
	cfg := config.DefaultConfig()
	// FailOnCriticalSecurity is nil (not set), so hard fail should not apply

	result := evaluator.EvaluateWithFindings(cfg, 95, []evaluator.Finding{
		{
			Severity: evaluator.SeverityCritical,
			Pillar:   "security",
		},
	})
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("score=95 with critical finding but hard-fail disabled: expected pass, got exit code %d", result.ExitCode)
	}
}

// Medium severity should NOT trigger hard fail even when enabled
func TestThreshold_HardFailIgnoresMediumSeverity(t *testing.T) {
	cfg := config.DefaultConfig()
	trueVal := true
	cfg.Gatekeeper.FailOnCriticalSecurity = &trueVal

	result := evaluator.EvaluateWithFindings(cfg, 95, []evaluator.Finding{
		{
			Severity: evaluator.SeverityMedium,
			Pillar:   "security",
		},
	})
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("score=95 with medium finding: expected pass, got exit code %d", result.ExitCode)
	}
}
