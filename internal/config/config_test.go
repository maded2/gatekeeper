package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eddie/gatekeeper/internal/config"
)

// =============================================================================
// Story 1.1: Repository Configuration File
// Acceptance Tests
// =============================================================================

func TestConfigurationFile_DefinesQualityGateSettings(t *testing.T) {
	// Given a repository with a .gatekeeper.yml config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")
	err := os.WriteFile(configPath, []byte("threshold: 7.0\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the tool loads the configuration
	cfg, err := config.LoadFromPath(configPath)

	// Then the configuration is loaded successfully
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Threshold != 7.0 {
		t.Errorf("expected threshold 7.0, got %f", cfg.Threshold)
	}
}

func TestConfigurationFile_SupportsQualityThreshold(t *testing.T) {
	// Given a config file with a threshold value
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")

	for _, tc := range []struct {
		name      string
		content   string
		expected  float64
	}{
		{"minimum threshold", "threshold: 1\n", 1.0},
		{"maximum threshold", "threshold: 10\n", 10.0},
		{"mid-range threshold", "threshold: 6.5\n", 6.5},
		{"default threshold", "threshold: 6\n", 6.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := os.WriteFile(configPath, []byte(tc.content), 0644)
			if err != nil {
				t.Fatal(err)
			}

			cfg, err := config.LoadFromPath(configPath)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if cfg.Threshold != tc.expected {
				t.Errorf("expected threshold %f, got %f", tc.expected, cfg.Threshold)
			}
		})
	}
}

func TestConfigurationFile_SupportsExcludePatterns(t *testing.T) {
	// Given a config file with exclude patterns
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")
	err := os.WriteFile(configPath, []byte(
		"exclude:\n  - \"*.lock\"\n  - \"*.min.js\"\n  - \"vendor/*\"\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the configuration is loaded
	cfg, err := config.LoadFromPath(configPath)

	// Then the exclude patterns are present
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(cfg.ExcludePatterns) == 0 {
		t.Error("expected exclude patterns to be loaded")
	}
}

func TestConfigurationFile_SupportsPillarWeights(t *testing.T) {
	// Given a config file with pillar weights
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")
	err := os.WriteFile(configPath, []byte(
		"pillar_weights:\n  code_quality: 0.5\n  test_coverage: 0.3\n  deployability: 0.2\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the configuration is loaded
	cfg, err := config.LoadFromPath(configPath)

	// Then the pillar weights are set correctly
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.PillarWeights.CodeQuality != 0.5 {
		t.Errorf("expected code_quality 0.5, got %f", cfg.PillarWeights.CodeQuality)
	}
	if cfg.PillarWeights.TestCoverage != 0.3 {
		t.Errorf("expected test_coverage 0.3, got %f", cfg.PillarWeights.TestCoverage)
	}
	if cfg.PillarWeights.Deployability != 0.2 {
		t.Errorf("expected deployability 0.2, got %f", cfg.PillarWeights.Deployability)
	}
}

func TestConfigurationFile_UsesHumanReadableFormat(t *testing.T) {
	// Given a YAML configuration file (human-readable format)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")
	yamlContent := `# Gatekeeper quality gate configuration
threshold: 6.0
exclude:
  - "*.lock"
  - "vendor/*"
pillar_weights:
  code_quality: 0.4
  test_coverage: 0.35
  deployability: 0.25
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the configuration is loaded
	cfg, err := config.LoadFromPath(configPath)

	// Then all settings are parsed correctly
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Threshold != 6.0 {
		t.Errorf("expected threshold 6.0, got %f", cfg.Threshold)
	}
}

func TestConfigurationFile_HasDefaultValues(t *testing.T) {
	// Given no configuration file exists
	_ = t.TempDir() // ensure temp dir cleanup

	// When the tool requests the default configuration
	cfg := config.Default()

	// Then default values are returned
	if cfg.Threshold != 6.0 {
		t.Errorf("expected default threshold 6.0, got %f", cfg.Threshold)
	}
	if cfg.PillarWeights.CodeQuality == 0 {
		t.Error("expected default code_quality weight to be set")
	}
	if cfg.PillarWeights.TestCoverage == 0 {
		t.Error("expected default test_coverage weight to be set")
	}
	if cfg.PillarWeights.Deployability == 0 {
		t.Error("expected default deployability weight to be set")
	}
}

func TestConfigurationFile_LoadsDefaultsWhenNoFilePresent(t *testing.T) {
	// Given a directory with no configuration file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")

	// When the tool tries to load the configuration
	cfg, err := config.LoadFromPath(configPath)

	// Then default configuration is returned without error
	if err != nil {
		t.Fatalf("expected no error for missing config, got: %v", err)
	}
	if cfg.Threshold != 6.0 {
		t.Errorf("expected default threshold 6.0, got %f", cfg.Threshold)
	}
}

func TestConfigurationFile_ReadsFromRepositoryRoot(t *testing.T) {
	// Given a config file at the repository root
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")
	err := os.WriteFile(configPath, []byte("threshold: 8.0\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the tool loads configuration from the repo root
	cfg, err := config.LoadFromRoot(tmpDir)

	// Then the configuration is found and loaded
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Threshold != 8.0 {
		t.Errorf("expected threshold 8.0, got %f", cfg.Threshold)
	}
}

func TestConfigurationFile_LoadsFromSubdirectory(t *testing.T) {
	// Given a config file at the repository root
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")
	err := os.WriteFile(configPath, []byte("threshold: 7.5\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	// And a subdirectory exists
	subDir := filepath.Join(tmpDir, "src", "internal")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// When the tool loads configuration from a subdirectory
	cfg, err := config.LoadFromRoot(subDir)

	// Then it walks up to find the config at the root
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Threshold != 7.5 {
		t.Errorf("expected threshold 7.5, got %f", cfg.Threshold)
	}
}
