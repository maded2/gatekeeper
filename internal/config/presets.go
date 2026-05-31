package config

import (
	"fmt"
	"strings"
)

// Preset represents a pre-built configuration profile.
type Preset struct {
	Name        string
	Description string
	Config      Configuration
}

// AvailablePresets returns all built-in presets.
var AvailablePresets = []Preset{
	{
		Name:        "strict",
		Description: "Maximum quality enforcement — for teams with mature CI/CD and experienced developers. High threshold (8/10) with heavy code quality weighting.",
		Config: Configuration{
			Threshold: 8.0,
			ExcludePatterns: defaultExcludePatterns,
			PillarWeights: PillarWeights{
				CodeQuality:   0.5,
				TestCoverage:  0.3,
				Deployability: 0.2,
			},
		},
	},
	{
		Name:        "balanced",
		Description: "Default balanced approach — suitable for most teams. Moderate threshold (6/10) with even pillar distribution.",
		Config: Configuration{
			Threshold: 6.0,
			ExcludePatterns: defaultExcludePatterns,
			PillarWeights: PillarWeights{
				CodeQuality:   0.4,
				TestCoverage:  0.35,
				Deployability: 0.25,
			},
		},
	},
	{
		Name:        "permissive",
		Description: "Lightweight gatekeeping — for early-stage projects or teams new to automated quality gates. Low threshold (4/10) with deployability priority.",
		Config: Configuration{
			Threshold: 4.0,
			ExcludePatterns: defaultExcludePatterns,
			PillarWeights: PillarWeights{
				CodeQuality:   0.25,
				TestCoverage:  0.25,
				Deployability: 0.5,
			},
		},
	},
}

var defaultExcludePatterns = []string{
	"*.lock",
	"*.min.*",
	"*.map",
	"*.pb.go",
	"*_gen.go",
	"*_mock.go",
	"vendor/*",
	"node_modules/*",
	".git/*",
	"__pycache__/*",
}

// GetPreset returns a preset by name. Returns an error if not found.
func GetPreset(name string) (Preset, error) {
	for _, p := range AvailablePresets {
		if p.Name == name {
			return p, nil
		}
	}
	return Preset{}, fmt.Errorf("unknown preset %q; available: %s", name, presetNames())
}

// ListPresets returns formatted preset descriptions for help output.
func ListPresets() string {
	var lines []string
	for _, p := range AvailablePresets {
		lines = append(lines, fmt.Sprintf("  %-12s %s", p.Name+":", p.Description))
	}
	return strings.Join(lines, "\n")
}

func presetNames() string {
	var names []string
	for _, p := range AvailablePresets {
		names = append(names, p.Name)
	}
	return strings.Join(names, ", ")
}

// ApplyPreset applies a preset's configuration, then overlays any non-zero
// values from the user config on top (allowing overrides).
func ApplyPreset(presetName string, userCfg Configuration) (Configuration, error) {
	preset, err := GetPreset(presetName)
	if err != nil {
		return Configuration{}, err
	}

	result := preset.Config
	// Override with user values where they differ from zero
	if userCfg.Threshold > 0 {
		result.Threshold = userCfg.Threshold
	}
	if len(userCfg.ExcludePatterns) > 0 {
		result.ExcludePatterns = userCfg.ExcludePatterns
	}
	if userCfg.PillarWeights.CodeQuality > 0 {
		result.PillarWeights.CodeQuality = userCfg.PillarWeights.CodeQuality
	}
	if userCfg.PillarWeights.TestCoverage > 0 {
		result.PillarWeights.TestCoverage = userCfg.PillarWeights.TestCoverage
	}
	if userCfg.PillarWeights.Deployability > 0 {
		result.PillarWeights.Deployability = userCfg.PillarWeights.Deployability
	}

	return result, nil
}
