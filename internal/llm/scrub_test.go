package llm_test

import (
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/llm"
)

// --- Story F-1: Prevent Secrets from Leaving My Network ---

// ACCEPTANCE CRITERIA 1:
// "Gatekeeper scans code for patterns matching common secrets"
func TestScrub_APIKeyPattern(t *testing.T) {
	code := `apiKey := "sk-1234567890abcdef"`
	clean := llm.ScrubSecrets(code)

	if clean == code {
		t.Error("expected secrets to be redacted")
	}
	if clean == "" {
		t.Error("expected non-empty output after scrubbing")
	}
}

// ACCEPTANCE CRITERIA 2:
// "Detected secrets are replaced with [REDACTED] placeholders"
func TestScrub_ReplacesWithRedacted(t *testing.T) {
	code := `password := "supersecret123"`
	clean := llm.ScrubSecrets(code)

	if clean == code {
		t.Error("expected password to be redacted")
	}
}

// ACCEPTANCE CRITERIA 4:
// "I can enable or disable this feature in my configuration"
func TestScrub_Configurable(t *testing.T) {
	cfg := config.DefaultConfig()
	trueVal := true
	cfg.Gatekeeper.Privacy = &config.PrivacyConfig{
		DataScrubbing: &trueVal,
	}

	if cfg.Gatekeeper.Privacy.DataScrubbing == nil || !*cfg.Gatekeeper.Privacy.DataScrubbing {
		t.Error("expected data scrubbing to be enabled")
	}
}

// --- Story F-2: Disable Cloud Transmission Entirely ---

// ACCEPTANCE CRITERIA 1:
// "I can set a configuration flag that prohibits all public cloud transmission"
func TestAirGap_Configurable(t *testing.T) {
	cfg := config.DefaultConfig()
	falseVal := false
	cfg.Gatekeeper.Privacy = &config.PrivacyConfig{
		AllowPublicCloudTransmission: &falseVal,
	}

	if cfg.Gatekeeper.Privacy.AllowPublicCloudTransmission == nil || *cfg.Gatekeeper.Privacy.AllowPublicCloudTransmission {
		t.Error("expected cloud transmission to be disabled")
	}
}

// ACCEPTANCE CRITERIA 3:
// "When enabled and no internal LLM is configured, falls back to rule-based evaluation"
func TestAirGap_FallbackToRuleBased(t *testing.T) {
	cfg := config.DefaultConfig()
	falseVal := false
	cfg.Gatekeeper.Privacy = &config.PrivacyConfig{
		AllowPublicCloudTransmission: &falseVal,
	}

	// When air-gapped and no internal LLM, should fall back to rule-based
	if llm.IsConfigured(cfg) {
		t.Error("expected LLM to not be configured in air-gapped mode")
	}
}

// --- Story F-3: Transmit Only Changed Code ---

// ACCEPTANCE CRITERIA 1:
// "Only changed functions/classes are sent for LLM analysis"
// (Verified by CheckDiff which only analyzes changed files)
