package config

// Configuration holds all quality gate settings.
type Configuration struct {
	Threshold       float64
	ExcludePatterns []string
	PillarWeights   PillarWeights
}

// PillarWeights defines the weight of each evaluation pillar.
type PillarWeights struct {
	CodeQuality   float64
	TestCoverage  float64
	Deployability float64
}

// Default returns the default configuration values (balanced preset).
func Default() Configuration {
	preset, _ := GetPreset("balanced")
	return preset.Config
}
