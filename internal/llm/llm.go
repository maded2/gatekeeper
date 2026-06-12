package llm

import (
	"gatekeeper/internal/config"
)

// Config holds LLM configuration for making requests.
type Config struct {
	BaseURL       string
	ModelName     string
	APIKey        string
	TimeoutMS     int
	Temperature   float64
}

// FromConfig creates an LLM Config from the application config.
func FromConfig(cfg config.GatekeeperConfig) (*Config, error) {
	if cfg.Gatekeeper.LLM == nil {
		return nil, nil
	}

	llmCfg := cfg.Gatekeeper.LLM
	return &Config{
		BaseURL:     llmCfg.BaseURL,
		ModelName:   llmCfg.ModelName,
		TimeoutMS:   llmCfg.TimeoutMS,
		Temperature: llmCfg.Temperature,
	}, nil
}

// IsConfigured returns true if LLM is configured.
func IsConfigured(cfg config.GatekeeperConfig) bool {
	return cfg.Gatekeeper.LLM != nil && cfg.Gatekeeper.LLM.BaseURL != ""
}
