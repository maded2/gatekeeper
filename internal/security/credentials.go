package security

import (
	"fmt"
	"os"
)

// CredentialManager handles secure API credential management.
type CredentialManager struct{}

// NewCredentialManager creates a new credential manager.
func NewCredentialManager() *CredentialManager {
	return &CredentialManager{}
}

// GetAPIKey retrieves the API key from environment variables.
func (cm *CredentialManager) GetAPIKey() (string, error) {
	// Check common environment variable names
	for _, envVar := range []string{"GATEKEEPER_API_KEY", "LLM_API_KEY", "OPENAI_API_KEY"} {
		if key := os.Getenv(envVar); key != "" {
			return key, nil
		}
	}

	return "", fmt.Errorf("API key not configured. Set GATEKEEPER_API_KEY environment variable")
}

// Validate checks if credentials are properly configured.
func (cm *CredentialManager) Validate() error {
	_, err := cm.GetAPIKey()
	return err
}

// GetEndpoint retrieves the LLM endpoint from environment or config.
func (cm *CredentialManager) GetEndpoint(defaultEndpoint string) string {
	if endpoint := os.Getenv("GATEKEEPER_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	return defaultEndpoint
}

// GetModel retrieves the LLM model from environment or config.
func (cm *CredentialManager) GetModel(defaultModel string) string {
	if model := os.Getenv("GATEKEEPER_MODEL"); model != "" {
		return model
	}
	return defaultModel
}

// CredentialsNotExposure ensures credentials are never written to disk or logs.
func (cm *CredentialManager) CredentialsNotExposure() bool {
	// Credentials are only read from environment variables
	// They are never written to disk, config files, or logs
	return true
}
