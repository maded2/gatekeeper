package evaluator_test

import (
	"fmt"
	"testing"
	"time"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
)

// --- Story G-1: Get Fast Feedback on Small Changes ---

// ACCEPTANCE CRITERIA 1:
// "Evaluating a commit with fewer than 10 changed files completes within 300 milliseconds"
func TestPerformance_SmallDiffUnder300ms(t *testing.T) {
	dir := setupWorkspace(t)

	// Create a few source files
	for i := 0; i < 9; i++ {
		mustWrite(t, dir, fmt.Sprintf("file%d.go", i), `package main

func hello() {
	println("hello")
}
`)
	}

	cfg := config.DefaultConfig()
	start := time.Now()
	result := evaluator.CheckWorkspace(cfg, dir)
	elapsed := time.Since(start)

	_ = result // result used

	if elapsed > 300*time.Millisecond {
		t.Errorf("workspace check took %v, expected < 300ms", elapsed)
	}
}

// ACCEPTANCE CRITERIA 3:
// "I see results immediately without perceptible delay"
func TestPerformance_TrivialChangeIsFast(t *testing.T) {
	start := time.Now()
	isTrivial := evaluator.IsTrivialChange("README.md", "")
	elapsed := time.Since(start)

	if !isTrivial {
		t.Error("expected README.md to be trivial")
	}

	if elapsed > 10*time.Millisecond {
		t.Errorf("trivial check took %v, expected < 10ms", elapsed)
	}
}
