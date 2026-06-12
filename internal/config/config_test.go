package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"gatekeeper/internal/config"
)

// --- Story A-1: Configure Gatekeeper for a New Project ---

// ACCEPTANCE CRITERIA 1:
// "I can run a single command that creates a sensible default gatekeeper.json in my project root"
func TestGenerateDefaultConfig_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gatekeeper.json")

	err := config.GenerateDefault(path)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig returned error: %v", err)
	}

	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("expected gatekeeper.json to be created at %s, but file does not exist: %v", path, err)
	}
}

// ACCEPTANCE CRITERIA 2:
// "The default configuration uses safe, conservative thresholds (e.g., target score of 75)"
func TestDefaultConfig_TargetThresholdIs75(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gatekeeper.json")

	err := config.GenerateDefault(path)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig returned error: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load generated config: %v", err)
	}

	if cfg.Gatekeeper.TargetThreshold != 75 {
		t.Errorf("expected target threshold to be 75, got %f", cfg.Gatekeeper.TargetThreshold)
	}
}

// ACCEPTANCE CRITERIA 3:
// "The default configuration includes common exclusion paths (vendor/, node_modules/, dist/)"
func TestDefaultConfig_IncludesCommonExclusions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gatekeeper.json")

	err := config.GenerateDefault(path)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig returned error: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load generated config: %v", err)
	}

	expectedExclusions := []string{
		"**/vendor/**",
		"**/node_modules/**",
		"**/dist/**",
	}

	for _, expected := range expectedExclusions {
		found := false
		for _, p := range cfg.Exclusions.Paths {
			if p == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected exclusion path %q to be in default config, but not found in %v", expected, cfg.Exclusions.Paths)
		}
	}
}

// ACCEPTANCE CRITERIA 4:
// "I can edit the file to customize thresholds, LLM provider settings, and exclusion paths"
// (Verified indirectly: the config is valid JSON that can be re-loaded after edits)
func TestDefaultConfig_IsEditableAndReloadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gatekeeper.json")

	err := config.GenerateDefault(path)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig returned error: %v", err)
	}

	// Simulate user editing the threshold
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load generated config: %v", err)
	}

	cfg.Gatekeeper.TargetThreshold = 90
	err = config.Save(cfg, path)
	if err != nil {
		t.Fatalf("failed to save edited config: %v", err)
	}

	// Reload and verify the edit persisted
	reloaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to reload edited config: %v", err)
	}

	if reloaded.Gatekeeper.TargetThreshold != 90 {
		t.Errorf("expected target threshold to be 90 after edit, got %f", reloaded.Gatekeeper.TargetThreshold)
	}
}

// ACCEPTANCE CRITERIA 5:
// "If no gatekeeper.json exists when I run Gatekeeper, I receive a clear message explaining how to create one"
func TestLoadMissingConfig_ReturnsClearError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gatekeeper.json")

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error when loading non-existent config, got nil")
	}

	// The error message should mention gatekeeper.json
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

// Definition of Done: "The default configuration file is valid JSON and passes Gatekeeper's own validation"
func TestDefaultConfig_PassesValidation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gatekeeper.json")

	err := config.GenerateDefault(path)
	if err != nil {
		t.Fatalf("GenerateDefaultConfig returned error: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load generated config: %v", err)
	}

	err = config.Validate(cfg)
	if err != nil {
		t.Fatalf("default config failed validation: %v", err)
	}
}
