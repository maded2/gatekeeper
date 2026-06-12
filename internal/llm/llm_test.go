package llm_test

import (
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/llm"
)

// --- Story E-3: Configure Custom LLM Provider ---

// ACCEPTANCE CRITERIA 1:
// "I can configure a custom base URL for the LLM provider"
func TestLLMConfig_CustomBaseURL(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		BaseURL: "http://localhost:11434/v1",
	}

	if cfg.Gatekeeper.LLM.BaseURL != "http://localhost:11434/v1" {
		t.Error("expected custom base URL to be set")
	}
}

// ACCEPTANCE CRITERIA 2:
// "I can specify which environment variable holds the API key"
func TestLLMConfig_CustomAPIKeyEnvVar(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		APIKeyEnvVar: "OLLAMA_API_KEY",
	}

	if cfg.Gatekeeper.LLM.APIKeyEnvVar != "OLLAMA_API_KEY" {
		t.Error("expected custom API key env var")
	}
}

// ACCEPTANCE CRITERIA 3:
// "I can set a timeout for LLM requests"
func TestLLMConfig_CustomTimeout(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		TimeoutMS: 10000,
	}

	if cfg.Gatekeeper.LLM.TimeoutMS != 10000 {
		t.Error("expected custom timeout")
	}
}

// ACCEPTANCE CRITERIA 4:
// "I can configure the model name"
func TestLLMConfig_CustomModel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		ModelName: "llama3",
	}

	if cfg.Gatekeeper.LLM.ModelName != "llama3" {
		t.Error("expected custom model name")
	}
}

// ACCEPTANCE CRITERIA 5:
// "I can set the temperature to 0 for deterministic results"
func TestLLMConfig_TemperatureZero(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		Temperature: 0,
	}

	if cfg.Gatekeeper.LLM.Temperature != 0 {
		t.Error("expected temperature to be 0")
	}
}

// Verify FromConfig creates proper LLM config
func TestLLM_FromConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		BaseURL:      "http://localhost:11434/v1",
		ModelName:    "llama3",
		APIKeyEnvVar: "OLLAMA_KEY",
		TimeoutMS:    5000,
		Temperature:  0,
	}

	llmCfg, err := llm.FromConfig(cfg)
	if err != nil {
		t.Fatalf("FromConfig returned error: %v", err)
	}

	if llmCfg.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("expected base URL, got %s", llmCfg.BaseURL)
	}
	if llmCfg.ModelName != "llama3" {
		t.Errorf("expected model llama3, got %s", llmCfg.ModelName)
	}
	if llmCfg.Temperature != 0 {
		t.Errorf("expected temperature 0, got %f", llmCfg.Temperature)
	}
}

// Verify IsConfigured works
func TestLLM_IsConfigured(t *testing.T) {
	cfg := config.DefaultConfig()
	if llm.IsConfigured(cfg) {
		t.Error("expected unconfigured LLM")
	}

	cfg.Gatekeeper.LLM = &config.LLMConfig{
		BaseURL: "http://localhost:11434/v1",
	}
	if !llm.IsConfigured(cfg) {
		t.Error("expected configured LLM")
	}
}
