package evaluator_test

import (
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
)

// --- Story D-1: Get a Deterministic Pass/Fail Signal for Pipelines ---

// ACCEPTANCE CRITERIA 1:
// "Exit code 0 means the Quality Score meets or exceeds the configured threshold (pass)"
func TestExitCode_PassWhenScoreMeetsThreshold(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.TargetThreshold = 75

	result := evaluator.Evaluate(cfg, 75)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("score=75, threshold=75: expected exit 0 (pass), got %d", result.ExitCode)
	}

	result = evaluator.Evaluate(cfg, 100)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("score=100, threshold=75: expected exit 0 (pass), got %d", result.ExitCode)
	}
}

// ACCEPTANCE CRITERIA 2:
// "Exit code 2 means the Quality Score is below the threshold (quality gate blocked)"
func TestExitCode_FailWhenScoreBelowThreshold(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.TargetThreshold = 75

	result := evaluator.Evaluate(cfg, 74)
	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("score=74, threshold=75: expected exit 2 (fail), got %d", result.ExitCode)
	}

	result = evaluator.Evaluate(cfg, 0)
	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("score=0, threshold=75: expected exit 2 (fail), got %d", result.ExitCode)
	}
}

// ACCEPTANCE CRITERIA 3:
// "Exit code 1 means an unexpected runtime error occurred"
func TestExitCode_ErrorCodeIs1(t *testing.T) {
	if evaluator.ExitError != 1 {
		t.Errorf("expected ExitError to be 1, got %d", evaluator.ExitError)
	}
}

// ACCEPTANCE CRITERIA 4:
// "The exit code behavior is consistent and deterministic across runs"
func TestExitCode_Deterministic(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.TargetThreshold = 80

	for i := 0; i < 10; i++ {
		result := evaluator.Evaluate(cfg, 80)
		if result.ExitCode != evaluator.ExitPass {
			t.Errorf("iteration %d: expected deterministic exit 0, got %d", i, result.ExitCode)
		}
	}
}

// Edge case: score exactly at boundary
func TestExitCode_BoundaryConditions(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.TargetThreshold = 50

	// Exactly at threshold
	result := evaluator.Evaluate(cfg, 50)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("score=50, threshold=50: expected pass, got %d", result.ExitCode)
	}

	// One below
	result = evaluator.Evaluate(cfg, 49)
	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("score=49, threshold=50: expected fail, got %d", result.ExitCode)
	}

	// Zero threshold — everything passes
	cfg.Gatekeeper.TargetThreshold = 0
	result = evaluator.Evaluate(cfg, 0)
	if result.ExitCode != evaluator.ExitPass {
		t.Errorf("score=0, threshold=0: expected pass, got %d", result.ExitCode)
	}

	// 100 threshold — only perfect passes
	cfg.Gatekeeper.TargetThreshold = 100
	result = evaluator.Evaluate(cfg, 99)
	if result.ExitCode != evaluator.ExitFail {
		t.Errorf("score=99, threshold=100: expected fail, got %d", result.ExitCode)
	}
}
