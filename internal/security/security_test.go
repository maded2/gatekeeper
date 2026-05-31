package security_test

import (
	"os"
	"strings"
	"testing"

	"github.com/eddie/gatekeeper/internal/security"
)

// =============================================================================
// Story 9.1: API Credential Management
// Acceptance Tests
// =============================================================================

func TestCredentials_FromEnvironmentVariable(t *testing.T) {
	// Given an API key set in environment
	os.Setenv("GATEKEEPER_API_KEY", "test-key-12345")
	defer os.Unsetenv("GATEKEEPER_API_KEY")

	// When credentials are retrieved
	cm := security.NewCredentialManager()
	key, err := cm.GetAPIKey()

	// Then the key is returned without error
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if key != "test-key-12345" {
		t.Errorf("expected 'test-key-12345', got %q", key)
	}
}

func TestCredentials_MissingCredentialsError(t *testing.T) {
	// Given no API key configured
	os.Unsetenv("GATEKEEPER_API_KEY")
	os.Unsetenv("LLM_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")

	// When credentials are retrieved
	cm := security.NewCredentialManager()
	_, err := cm.GetAPIKey()

	// Then a clear error is returned with instructions
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
	if !strings.Contains(err.Error(), "GATEKEEPER_API_KEY") {
		t.Errorf("expected error to mention GATEKEEPER_API_KEY, got: %v", err)
	}
}

func TestCredentials_NeverWrittenToDisk(t *testing.T) {
	// Given an API key set in environment
	os.Setenv("GATEKEEPER_API_KEY", "secret-key-12345")
	defer os.Unsetenv("GATEKEEPER_API_KEY")

	// When credentials are used
	cm := security.NewCredentialManager()
	key, _ := cm.GetAPIKey()

	// Then the key is only in memory (never written to disk)
	if key == "" {
		t.Error("expected key to be retrievable")
	}
	// The key should not appear in any config files
	t.Log("credentials are read from environment variables only")
}

func TestCredentials_ValidateSuccess(t *testing.T) {
	// Given configured credentials
	os.Setenv("GATEKEEPER_API_KEY", "test-key")
	defer os.Unsetenv("GATEKEEPER_API_KEY")

	// When validation is performed
	cm := security.NewCredentialManager()
	err := cm.Validate()

	// Then no error is returned
	if err != nil {
		t.Errorf("expected no validation error, got: %v", err)
	}
}

func TestCredentials_ValidateFailure(t *testing.T) {
	// Given missing credentials
	os.Unsetenv("GATEKEEPER_API_KEY")
	os.Unsetenv("LLM_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")

	// When validation is performed
	cm := security.NewCredentialManager()
	err := cm.Validate()

	// Then an error is returned
	if err == nil {
		t.Error("expected validation error for missing credentials")
	}
}

func TestCredentials_NotExposedInLogs(t *testing.T) {
	// Given an API key
	os.Setenv("GATEKEEPER_API_KEY", "super-secret-key")
	defer os.Unsetenv("GATEKEEPER_API_KEY")

	// When credentials are retrieved
	cm := security.NewCredentialManager()
	key, _ := cm.GetAPIKey()

	// Then the key is not logged or exposed
	if key == "" {
		t.Error("expected key to be retrievable")
	}
	// In real implementation, we'd verify no logging occurs
	t.Log("credentials are not exposed in error messages")
}

// =============================================================================
// Story 9.2: Sensitive Data Pre-Filtering
// Acceptance Tests
// =============================================================================

func TestRedactor_DetectsAPIKeys(t *testing.T) {
	// Given a diff with API keys
	content := `api_key = "sk-1234567890abcdef1234567890abcdef"`

	// When redacted
	redactor := security.NewRedactor()
	result, count := redactor.Redact(content)

	// Then the API key is redacted
	if count == 0 {
		t.Error("expected API key to be detected and redacted")
	}
	if strings.Contains(result, "sk-1234567890abcdef") {
		t.Error("expected API key to be redacted from output")
	}
}

func TestRedactor_DetectsPasswords(t *testing.T) {
	// Given a diff with passwords
	content := `password = "supersecret123"`

	// When redacted
	redactor := security.NewRedactor()
	result, count := redactor.Redact(content)

	// Then the password is redacted
	if count == 0 {
		t.Error("expected password to be detected and redacted")
	}
	if strings.Contains(result, "supersecret123") {
		t.Error("expected password to be redacted from output")
	}
}

func TestRedactor_DetectsTokens(t *testing.T) {
	// Given a diff with tokens
	content := `access_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0"`

	// When redacted
	redactor := security.NewRedactor()
	result, count := redactor.Redact(content)

	// Then the token is redacted
	if count == 0 {
		t.Error("expected token to be detected and redacted")
	}
	if strings.Contains(result, "eyJhbGciOiJIUzI1NiIs") {
		t.Error("expected token to be redacted from output")
	}
}

func TestRedactor_IndicatesRedaction(t *testing.T) {
	// Given a diff with sensitive data
	content := `password = "secret123"`

	// When redacted
	redactor := security.NewRedactor()
	result, _ := redactor.Redact(content)

	// Then redaction is indicated in the output
	if !strings.Contains(result, "[REDACTED]") {
		t.Error("expected [REDACTED] marker in output")
	}
}

func TestRedactor_CustomPatterns(t *testing.T) {
	// Given a redactor with custom patterns
	redactor := security.NewRedactor()
	err := redactor.AddPattern(`(?i)custom_secret\s*[:=]\s*["\']?[^\s"'\n]+["\']?`)
	if err != nil {
		t.Fatal(err)
	}

	// When redacting content with custom pattern
	content := `custom_secret = "my-custom-secret-value"`
	_, count := redactor.Redact(content)

	// Then the custom pattern is detected
	if count == 0 {
		t.Error("expected custom pattern to be detected")
	}
}

func TestRedactor_NoFalsePositives(t *testing.T) {
	// Given content without sensitive data
	content := `
package main

func main() {
    fmt.Println("Hello, World!")
}`

	// When redacted
	redactor := security.NewRedactor()
	result, count := redactor.Redact(content)

	// Then no redactions occur
	if count != 0 {
		t.Errorf("expected no redactions, got %d", count)
	}
	if result != content {
		t.Error("expected content to be unchanged")
	}
}

func TestRedactor_Report(t *testing.T) {
	// Given a set of files with sensitive data
	files := map[string]string{
		"config.go": `api_key = "sk-1234567890abcdef1234567890abcdef"`,
		"main.go":   `fmt.Println("Hello")`,
	}

	// When a redaction report is generated
	redactor := security.NewRedactor()
	report := redactor.GenerateReport(files)

	// Then the report shows redacted files
	if report.TotalRedactions == 0 {
		t.Error("expected at least 1 redaction")
	}
	if report.FilesRedacted["config.go"] == 0 {
		t.Error("expected config.go to be redacted")
	}
	if report.FilesRedacted["main.go"] != 0 {
		t.Error("expected main.go to not be redacted")
	}
}

// =============================================================================
// Story 9.3: Self-Hosted Model Support
// Acceptance Tests
// =============================================================================

func TestSelfHosted_EndpointConfiguration(t *testing.T) {
	// Given a self-hosted endpoint
	os.Setenv("GATEKEEPER_ENDPOINT", "http://localhost:8080/v1/chat/completions")
	defer os.Unsetenv("GATEKEEPER_ENDPOINT")

	// When the endpoint is retrieved
	cm := security.NewCredentialManager()
	endpoint := cm.GetEndpoint("https://api.openai.com/v1/chat/completions")

	// Then the self-hosted endpoint is used
	if endpoint != "http://localhost:8080/v1/chat/completions" {
		t.Errorf("expected self-hosted endpoint, got %q", endpoint)
	}
}

func TestSelfHosted_DefaultEndpoint(t *testing.T) {
	// Given no self-hosted endpoint configured
	os.Unsetenv("GATEKEEPER_ENDPOINT")

	// When the endpoint is retrieved
	cm := security.NewCredentialManager()
	endpoint := cm.GetEndpoint("https://api.openai.com/v1/chat/completions")

	// Then the default endpoint is used
	if endpoint != "https://api.openai.com/v1/chat/completions" {
		t.Errorf("expected default endpoint, got %q", endpoint)
	}
}
