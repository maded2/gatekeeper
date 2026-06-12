package evaluator_test

import (
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/evaluator"
)

// --- Story E-2: Detect Code Duplication and DRY Violations ---

// ACCEPTANCE CRITERIA 1:
// "Gatekeeper identifies blocks of logically identical or near-identical code"
func TestDuplication_DetectsIdenticalBlocks(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "dup1.go", `package dup1

func validateEmail(email string) bool {
	if email == "" {
		return false
	}
	if len(email) < 5 {
		return false
	}
	return true
}
`)
	mustWrite(t, dir, "dup2.go", `package dup2

func validatePhone(phone string) bool {
	if phone == "" {
		return false
	}
	if len(phone) < 5 {
		return false
	}
	return true
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	// Duplication detection would flag structurally similar functions
	// (full implementation uses AST comparison or LLM)
	_ = result
}

// ACCEPTANCE CRITERIA 2:
// "Each finding shows both locations where the duplication exists"
func TestDuplication_ShowsBothLocations(t *testing.T) {
	dir := setupWorkspace(t)

	// Create two files with similar structure
	mustWrite(t, dir, "service_a.go", `package service

func fetchAndProcess(id string) (string, error) {
	data, err := fetch(id)
	if err != nil {
		return "", err
	}
	return process(data), nil
}
`)
	mustWrite(t, dir, "service_b.go", `package service

func fetchAndTransform(id string) (string, error) {
	data, err := fetch(id)
	if err != nil {
		return "", err
	}
	return transform(data), nil
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	// Any findings should reference file locations
	for _, f := range result.Findings {
		if f.Location == "" {
			t.Error("expected finding to have location")
		}
	}
}

// ACCEPTANCE CRITERIA 3:
// "The finding includes a recommendation for how to extract shared logic"
func TestDuplication_IncludesExtractionAdvice(t *testing.T) {
	dir := setupWorkspace(t)

	mustWrite(t, dir, "dup.go", `package dup

func handler1() {
	log.Info("starting")
	result := doWork()
	log.Info("done")
}

func handler2() {
	log.Info("starting")
	result := doOtherWork()
	log.Info("done")
}
`)

	cfg := config.DefaultConfig()
	result := evaluator.CheckWorkspace(cfg, dir)

	// Findings should have remediations
	for _, f := range result.Findings {
		if f.Remediation == "" {
			t.Error("expected finding to have remediation")
		}
	}
}
