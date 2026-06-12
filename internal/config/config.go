package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// GatekeeperConfig represents the full gatekeeper.json configuration.
type GatekeeperConfig struct {
	Gatekeeper  GatekeeperSection `json:"gatekeeper"`
	Pillars     PillarsSection    `json:"pillars"`
	Exclusions  ExclusionsSection `json:"exclusions"`
}

// GatekeeperSection holds top-level gatekeeper settings.
type GatekeeperSection struct {
	TargetThreshold       float64   `json:"target_threshold"`
	FailOnCriticalSecurity *bool     `json:"fail_on_critical_security,omitempty"`
	LLM                   *LLMConfig `json:"llm,omitempty"`
	Privacy               *PrivacyConfig `json:"privacy,omitempty"`
}

// LLMConfig holds LLM provider settings.
type LLMConfig struct {
	Provider      string  `json:"provider,omitempty"`
	BaseURL       string  `json:"base_url,omitempty"`
	ModelName     string  `json:"model_name,omitempty"`
	APIKeyEnvVar  string  `json:"api_key_env_var,omitempty"`
	TimeoutMS     int     `json:"timeout_ms,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
}

// PrivacyConfig holds privacy / air-gapped settings.
type PrivacyConfig struct {
	AllowPublicCloudTransmission *bool `json:"allow_public_cloud_transmission,omitempty"`
	DataScrubbing                *bool `json:"data_scrubbing,omitempty"`
}

// PillarsSection holds pillar-specific tuning.
type PillarsSection struct {
	Static       StaticPillar       `json:"static,omitempty"`
	Verification VerificationPillar `json:"verification,omitempty"`
}

// StaticPillar holds static analysis settings.
type StaticPillar struct {
	MaxCyclomaticComplexity int  `json:"max_cyclomatic_complexity,omitempty"`
	AllowExperimentalModules *bool `json:"allow_experimental_modules,omitempty"`
}

// VerificationPillar holds verification / testing settings.
type VerificationPillar struct {
	RequireMutationTesting    *bool  `json:"require_mutation_testing,omitempty"`
	MutationCoverageFloor     float64 `json:"mutation_coverage_floor,omitempty"`
	HistoricalPipelineURL     string  `json:"historical_pipeline_telemetry_url,omitempty"`
}

// ExclusionsSection holds file exclusion patterns.
type ExclusionsSection struct {
	Paths []string `json:"paths"`
}

// Default exclusion paths included in every generated config.
var defaultExclusions = []string{
	"**/vendor/**",
	"**/node_modules/**",
	"**/dist/**",
}

// DefaultExclusions returns a copy of the default exclusion paths.
func DefaultExclusions() []string {
	result := make([]string, len(defaultExclusions))
	copy(result, defaultExclusions)
	return result
}

// DefaultConfig returns a GatekeeperConfig with safe, conservative defaults.
func DefaultConfig() GatekeeperConfig {
	return GatekeeperConfig{
		Gatekeeper: GatekeeperSection{
			TargetThreshold: 75,
		},
		Pillars: PillarsSection{
			Verification: VerificationPillar{
				MutationCoverageFloor: 80,
			},
		},
		Exclusions: ExclusionsSection{
			Paths: DefaultExclusions(),
		},
	}
}

// GenerateDefault writes a default gatekeeper.json to the given path.
func GenerateDefault(path string) error {
	cfg := DefaultConfig()
	return writeJSON(path, cfg)
}

// Load reads and parses a gatekeeper.json from the given path.
func Load(path string) (*GatekeeperConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found at %s: run 'gatekeeper init' to create a default configuration", path)
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg GatekeeperConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config JSON: %w", err)
	}

	return &cfg, nil
}

// writeJSON marshals the value to pretty JSON and writes it to the given path.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

// Save writes the config back to the given path.
func Save(cfg *GatekeeperConfig, path string) error {
	return writeJSON(path, cfg)
}

// Validate checks the config for required constraints.
func Validate(cfg *GatekeeperConfig) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	if cfg.Gatekeeper.TargetThreshold < 0 || cfg.Gatekeeper.TargetThreshold > 100 {
		return fmt.Errorf("target_threshold must be between 0 and 100, got %f", cfg.Gatekeeper.TargetThreshold)
	}

	if cfg.Gatekeeper.LLM != nil && cfg.Gatekeeper.LLM.APIKeyEnvVar != "" {
		if os.Getenv(cfg.Gatekeeper.LLM.APIKeyEnvVar) == "" {
			return fmt.Errorf("environment variable %s is not set; required for LLM authentication", cfg.Gatekeeper.LLM.APIKeyEnvVar)
		}
	}

	return nil
}
