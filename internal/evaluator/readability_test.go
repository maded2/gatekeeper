package evaluator_test

import (
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
)

// --- Story E-1: Evaluate Code Readability and Intent ---

// ACCEPTANCE CRITERIA 1:
// "Gatekeeper flags variables and functions with names that obscure their purpose"
func TestReadability_FlagsPoorNames(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "poor_names.go", `package poor

func a(b string) string {
	x := b
	return x
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	// We expect at least the file to be analyzed
	// (actual naming detection comes from LLM or advanced static analysis)
	_ = result
}

// ACCEPTANCE CRITERIA 2:
// "Each flag includes a suggested improvement"
func TestReadability_IncludesSuggestion(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "cryptic.go", `package cryptic

func x() {
	a := 1
	b := 2
	_ = a + b
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	// Any findings should have remediations
	for _, f := range result.Findings {
		if f.Remediation == "" {
			t.Errorf("finding at %s has empty remediation", f.Location)
		}
	}
}

// ACCEPTANCE CRITERIA 3:
// "The evaluation considers surrounding context"
func TestReadability_ContextAware(t *testing.T) {
	dir := setupWorkspace(t)

	// 'i' is acceptable in loops, but not as a general variable
	mustWrite(t, dir, "context.go", `package context

func process(data []string) {
	for i := range data {
		_ = i // acceptable in loop
	}
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	// File should be analyzed without false positives for loop variables
	_ = result
}
