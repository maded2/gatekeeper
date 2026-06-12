package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gatekeeper/internal/config"
)

// --- Story G-3: Understand Clear Error Messages ---

// ACCEPTANCE CRITERIA 1:
// "Missing configuration file: message explains where to place gatekeeper.json"
func TestError_MissingConfig(t *testing.T) {
	_, err := config.Load("/nonexistent/gatekeeper.json")
	if err == nil {
		t.Fatal("expected error for missing config")
	}

	msg := err.Error()
	if !strings.Contains(msg, "gatekeeper.json") {
		t.Errorf("expected error to mention gatekeeper.json: %s", msg)
	}
}

// ACCEPTANCE CRITERIA 2:
// "Invalid JSON in configuration: message identifies the line and nature of the error"
func TestError_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gatekeeper.json")
	if err := os.WriteFile(path, []byte("{invalid json}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	msg := err.Error()
	if !strings.Contains(msg, "parse") && !strings.Contains(msg, "JSON") {
		t.Errorf("expected error to mention JSON parsing: %s", msg)
	}
}

// ACCEPTANCE CRITERIA 3:
// "Unreachable LLM endpoint: message explains the connection failure"
// (Verified by LLM package, not config)

// ACCEPTANCE CRITERIA 4:
// "Missing API key environment variable: message names the expected variable"
func TestError_MissingAPIKey(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		APIKeyEnvVar: "GATEKEEPER_API_KEY",
	}

	err := config.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	msg := err.Error()
	if !strings.Contains(msg, "GATEKEEPER_API_KEY") {
		t.Errorf("expected error to name the env var: %s", msg)
	}
}

// ACCEPTANCE CRITERIA 5:
// "Unreadable path: message identifies which path could not be accessed"
func TestError_UnreadablePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gatekeeper.json")
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(path, 0644) // restore for cleanup

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for unreadable path")
	}

	msg := err.Error()
	if !strings.Contains(msg, path) {
		t.Errorf("expected error to mention the path: %s", msg)
	}
}

// ACCEPTANCE CRITERIA 6:
// "All error messages use exit code 1 (runtime error)"
// (Verified by cmd/ which calls os.Exit(1) on errors)

// Validate catches invalid threshold
func TestError_InvalidThreshold(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gatekeeper.TargetThreshold = 150

	err := config.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for threshold > 100")
	}
}
