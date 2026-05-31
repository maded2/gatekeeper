package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eddie/gatekeeper/internal/config"
)

// =============================================================================
// Story 1.2: Configuration Validation
// Acceptance Tests
// =============================================================================

func TestValidation_ValidConfigurationPasses(t *testing.T) {
	// Given a valid configuration
	cfg := config.Default()

	// When the configuration is validated
	errors := config.Validate(cfg)

	// Then no errors are returned
	if len(errors) > 0 {
		t.Errorf("expected no errors for valid config, got: %v", errors)
	}
}

func TestValidation_ThresholdOutOfRange(t *testing.T) {
	// Given a configuration with out-of-range threshold
	for _, tc := range []struct {
		name      string
		threshold float64
	}{
		{"threshold below minimum", 0.0},
		{"threshold above maximum", 15.0},
		{"threshold negative", -1.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.Threshold = tc.threshold

			// When the configuration is validated
			errors := config.Validate(cfg)

			// Then an error is returned with the field name and valid range
			if len(errors) == 0 {
				t.Error("expected validation error for out-of-range threshold")
			}
			errMsg := errors[0].Error()
			if !strings.Contains(errMsg, "threshold") {
				t.Errorf("expected error to mention 'threshold', got: %s", errMsg)
			}
			if !strings.Contains(errMsg, "1") || !strings.Contains(errMsg, "10") {
				t.Errorf("expected error to mention valid range, got: %s", errMsg)
			}
		})
	}
}

func TestValidation_ThresholdAtBoundaries(t *testing.T) {
	// Given configurations at boundary values
	for _, tc := range []struct {
		name      string
		threshold float64
		valid     bool
	}{
		{"threshold at minimum (1.0)", 1.0, true},
		{"threshold at maximum (10.0)", 10.0, true},
		{"threshold just below minimum", 0.99, false},
		{"threshold just above maximum", 10.01, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.Threshold = tc.threshold

			errors := config.Validate(cfg)

			if tc.valid && len(errors) > 0 {
				t.Errorf("expected valid config, got errors: %v", errors)
			}
			if !tc.valid && len(errors) == 0 {
				t.Error("expected validation error")
			}
		})
	}
}

func TestValidation_PillarWeightsOutOfRange(t *testing.T) {
	// Given a configuration with out-of-range pillar weights
	cfg := config.Default()
	cfg.PillarWeights.CodeQuality = 1.5

	// When the configuration is validated
	errors := config.Validate(cfg)

	// Then an error is returned for the specific field
	if len(errors) == 0 {
		t.Error("expected validation error for out-of-range weight")
	}
	found := false
	for _, err := range errors {
		if strings.Contains(err.Error(), "code_quality") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error to mention 'code_quality', got: %v", errors)
	}
}

func TestValidation_PillarWeightsDoNotSumToOne(t *testing.T) {
	// Given a configuration where weights don't sum to 1.0
	cfg := config.Default()
	cfg.PillarWeights.CodeQuality = 0.5
	cfg.PillarWeights.TestCoverage = 0.5
	cfg.PillarWeights.Deployability = 0.5

	// When the configuration is validated
	errors := config.Validate(cfg)

	// Then an error is returned about the sum
	if len(errors) == 0 {
		t.Error("expected validation error for weights not summing to 1.0")
	}
	found := false
	for _, err := range errors {
		if strings.Contains(err.Error(), "sum") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error to mention sum, got: %v", errors)
	}
}

func TestValidation_PillarWeightsSumWithTolerance(t *testing.T) {
	// Given a configuration where weights sum close to 1.0 (floating point tolerance)
	cfg := config.Default()
	cfg.PillarWeights.CodeQuality = 0.33
	cfg.PillarWeights.TestCoverage = 0.33
	cfg.PillarWeights.Deployability = 0.34 // sums to 1.0

	// When the configuration is validated
	errors := config.Validate(cfg)

	// Then no sum error is returned
	sumErrors := 0
	for _, err := range errors {
		if strings.Contains(err.Error(), "sum") {
			sumErrors++
		}
	}
	if sumErrors > 0 {
		t.Errorf("expected no sum error for weights summing to 1.0, got: %v", errors)
	}
}

func TestValidation_InvalidConfigFromFile(t *testing.T) {
	// Given a config file with invalid threshold
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")
	err := os.WriteFile(configPath, []byte("threshold: 15\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the configuration is loaded
	cfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("expected no load error, got: %v", err)
	}

	// Then validation catches the error
	errors := config.Validate(cfg)
	if len(errors) == 0 {
		t.Error("expected validation error for threshold 15")
	}
}

func TestValidation_ClearErrorMessage(t *testing.T) {
	// Given a configuration with multiple issues
	cfg := config.Configuration{
		Threshold: 0,
		PillarWeights: config.PillarWeights{
			CodeQuality:   -0.5,
			TestCoverage:  2.0,
			Deployability: 0.0,
		},
	}

	// When the configuration is validated
	errors := config.Validate(cfg)

	// Then each error identifies its specific field
	if len(errors) < 2 {
		t.Errorf("expected at least 2 errors, got %d", len(errors))
	}

	// Each error should mention its field
	for _, err := range errors {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "threshold") &&
			!strings.Contains(errMsg, "code_quality") &&
			!strings.Contains(errMsg, "test_coverage") &&
			!strings.Contains(errMsg, "sum") {
			t.Errorf("unexpected error format: %s", errMsg)
		}
	}
}

func TestValidation_ValidConfigFromYAML(t *testing.T) {
	// Given a valid YAML config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")
	content := `threshold: 7.0
pillar_weights:
  code_quality: 0.4
  test_coverage: 0.35
  deployability: 0.25
`
	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the configuration is loaded and validated
	cfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("expected no load error, got: %v", err)
	}
	errors := config.Validate(cfg)

	// Then no validation errors
	if len(errors) > 0 {
		t.Errorf("expected no validation errors, got: %v", errors)
	}
}

func TestValidation_MissingRequiredFields(t *testing.T) {
	// Given an empty configuration file (all zeros)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gatekeeper.yml")
	err := os.WriteFile(configPath, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the configuration is loaded
	cfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("expected no load error, got: %v", err)
	}

	// Then the config uses defaults (not zeros)
	if cfg.Threshold == 0 {
		t.Error("expected threshold to default to 6.0, not 0")
	}
	if cfg.PillarWeights.CodeQuality == 0 {
		t.Error("expected code_quality to have a default, not 0")
	}
}

func TestValidation_ErrorFormat(t *testing.T) {
	// Given a validation error
	err := config.ValidationError{
		Field:   "threshold",
		Message: "must be between 1 and 10",
	}

	// When the error is formatted
	msg := err.Error()

	// Then it includes both field and message
	expected := fmt.Sprintf("invalid config field %q: %s", "threshold", "must be between 1 and 10")
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}
