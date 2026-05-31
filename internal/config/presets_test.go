package config_test

import (
	"strings"
	"testing"

	"github.com/eddie/gatekeeper/internal/config"
)

// =============================================================================
// Story 1.3: Configuration Presets
// Acceptance Tests
// =============================================================================

func TestPresets_AtLeastThreeAvailable(t *testing.T) {
	// When listing available presets
	presets := config.AvailablePresets

	// Then at least three presets exist
	if len(presets) < 3 {
		t.Errorf("expected at least 3 presets, got %d", len(presets))
	}
}

func TestPresets_IncludeStrictBalancedPermissive(t *testing.T) {
	// Given the available presets
	presets := config.AvailablePresets

	// Then they include strict, balanced, and permissive
	found := map[string]bool{}
	for _, p := range presets {
		found[p.Name] = true
	}

	for _, expected := range []string{"strict", "balanced", "permissive"} {
		if !found[expected] {
			t.Errorf("expected preset %q to be available", expected)
		}
	}
}

func TestPresets_EachHasDescription(t *testing.T) {
	// Given each preset
	for _, p := range config.AvailablePresets {
		t.Run(p.Name, func(t *testing.T) {
			// Then it has a non-empty description
			if p.Description == "" {
				t.Errorf("preset %q has no description", p.Name)
			}
		})
	}
}

func TestPresets_StrictHasHighThreshold(t *testing.T) {
	// Given the strict preset
	preset, err := config.GetPreset("strict")
	if err != nil {
		t.Fatal(err)
	}

	// Then it has a high threshold
	if preset.Config.Threshold < 7.0 {
		t.Errorf("expected strict preset threshold >= 7.0, got %f", preset.Config.Threshold)
	}
}

func TestPresets_BalancedHasModerateThreshold(t *testing.T) {
	// Given the balanced preset
	preset, err := config.GetPreset("balanced")
	if err != nil {
		t.Fatal(err)
	}

	// Then it has a moderate threshold
	if preset.Config.Threshold < 5.0 || preset.Config.Threshold > 7.0 {
		t.Errorf("expected balanced preset threshold between 5-7, got %f", preset.Config.Threshold)
	}
}

func TestPresets_PermissiveHasLowThreshold(t *testing.T) {
	// Given the permissive preset
	preset, err := config.GetPreset("permissive")
	if err != nil {
		t.Fatal(err)
	}

	// Then it has a low threshold
	if preset.Config.Threshold > 5.0 {
		t.Errorf("expected permissive preset threshold <= 5.0, got %f", preset.Config.Threshold)
	}
}

func TestPresets_GetPresetByName(t *testing.T) {
	// When getting a preset by name
	preset, err := config.GetPreset("strict")

	// Then it returns the correct preset
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if preset.Name != "strict" {
		t.Errorf("expected preset name 'strict', got %q", preset.Name)
	}
}

func TestPresets_GetPresetUnknownName(t *testing.T) {
	// When getting an unknown preset
	_, err := config.GetPreset("nonexistent")

	// Then an error is returned with available preset names
	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "strict") || !strings.Contains(errMsg, "balanced") {
		t.Errorf("expected error to list available presets, got: %s", errMsg)
	}
}

func TestPresets_OverrideThreshold(t *testing.T) {
	// Given a preset with user override for threshold
	userCfg := config.Configuration{
		Threshold: 9.0,
	}

	// When applying the preset with override
	result, err := config.ApplyPreset("balanced", userCfg)
	if err != nil {
		t.Fatal(err)
	}

	// Then the threshold is overridden but other settings come from preset
	if result.Threshold != 9.0 {
		t.Errorf("expected overridden threshold 9.0, got %f", result.Threshold)
	}
}

func TestPresets_OverridePillarWeights(t *testing.T) {
	// Given a preset with user override for pillar weights
	userCfg := config.Configuration{
		PillarWeights: config.PillarWeights{
			CodeQuality: 0.6,
		},
	}

	// When applying the preset with override
	result, err := config.ApplyPreset("balanced", userCfg)
	if err != nil {
		t.Fatal(err)
	}

	// Then the code quality weight is overridden
	if result.PillarWeights.CodeQuality != 0.6 {
		t.Errorf("expected overridden code_quality 0.6, got %f", result.PillarWeights.CodeQuality)
	}
	// And other weights come from preset
	if result.PillarWeights.TestCoverage == 0 {
		t.Error("expected test_coverage from preset, got 0")
	}
}

func TestPresets_DefaultIsBalanced(t *testing.T) {
	// When getting the default configuration
	cfg := config.Default()

	// Then it matches the balanced preset
	balanced, _ := config.GetPreset("balanced")
	if cfg.Threshold != balanced.Config.Threshold {
		t.Errorf("expected default threshold to match balanced preset (%f), got %f",
			balanced.Config.Threshold, cfg.Threshold)
	}
}

func TestPresets_ListPresetsOutput(t *testing.T) {
	// When listing presets
	output := config.ListPresets()

	// Then all preset names appear in the output
	for _, p := range config.AvailablePresets {
		if !strings.Contains(output, p.Name) {
			t.Errorf("expected preset %q in list output", p.Name)
		}
		if !strings.Contains(output, p.Description) {
			t.Errorf("expected description for %q in list output", p.Name)
		}
	}
}

func TestPresets_ApplyPresetUnknownName(t *testing.T) {
	// When applying an unknown preset
	_, err := config.ApplyPreset("unknown", config.Configuration{})

	// Then an error is returned
	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
}
