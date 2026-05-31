package config

// Configuration holds all quality gate settings.
type Configuration struct {
	Threshold       float64
	ExcludePatterns []string
	PillarWeights   PillarWeights
	Retry           RetryConfig
	Cost            CostConfig
}

// PillarWeights defines the weight of each evaluation pillar.
type PillarWeights struct {
	CodeQuality   float64
	TestCoverage  float64
	Deployability float64
}

// RetryConfig defines retry behavior for transient API failures.
type RetryConfig struct {
	MaxRetries        int
	InitialDelayMs    int
	MaxDelayMs        int
	BackoffMultiplier float64
}

// CostConfig defines cost monitoring settings.
type CostConfig struct {
	MaxCostPerEvaluation float64
	Currency             string
}

// Default returns the default configuration values (balanced preset).
func Default() Configuration {
	preset, _ := GetPreset("balanced")
	cfg := preset.Config
	cfg.Retry = RetryConfig{
		MaxRetries:        3,
		InitialDelayMs:    1000,
		MaxDelayMs:        10000,
		BackoffMultiplier: 2.0,
	}
	cfg.Cost = CostConfig{
		Currency: "USD",
	}
	return cfg
}
