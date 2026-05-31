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

// Default returns the default configuration values.
func Default() Configuration {
	return Configuration{
		Threshold: 6.0,
		ExcludePatterns: []string{
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
		},
		PillarWeights: PillarWeights{
			CodeQuality:   0.4,
			TestCoverage:  0.35,
			Deployability: 0.25,
		},
	}
}
