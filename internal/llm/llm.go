package llm

import (
	"context"
	"os"

	"gatekeeper/internal/config"
)

// Config holds LLM configuration for making requests.
type Config struct {
	BaseURL      string
	ModelName    string
	APIKey       string
	TimeoutMS    int
	Temperature  float64
	AuthType     config.AuthType
	oauthManager *BrowserOAuthManager
}

// FromConfig creates an LLM Config from the application config.
func FromConfig(cfg config.GatekeeperConfig) (*Config, error) {
	if cfg.Gatekeeper.LLM == nil {
		return nil, nil
	}

	llmCfg := cfg.Gatekeeper.LLM

	// Determine auth type (default to api_key for backward compatibility)
	authType := llmCfg.AuthType
	if authType == "" {
		authType = config.AuthAPIKey
	}

	result := &Config{
		BaseURL:     llmCfg.BaseURL,
		ModelName:   llmCfg.ModelName,
		TimeoutMS:   llmCfg.TimeoutMS,
		Temperature: llmCfg.Temperature,
		AuthType:    authType,
	}

	// Set up API key or OAuth based on auth type
	switch authType {
	case config.AuthAPIKey:
		if llmCfg.APIKeyEnvVar != "" {
			result.APIKey = os.Getenv(llmCfg.APIKeyEnvVar)
		}
	case config.AuthOAuthBrowser:
		oauthCfg := OAuthConfig{
			AuthURL:        llmCfg.OAuthAuthURL,
			TokenURL:       llmCfg.OAuthTokenURL,
			ClientID:       os.Getenv(llmCfg.OAuthClientIDEnvVar),
			ClientSecret:   os.Getenv(llmCfg.OAuthClientSecretEnvVar),
			Scopes:         llmCfg.OAuthScopes,
			RedirectURL:    llmCfg.OAuthRedirectURL,
			TokenCacheFile: llmCfg.OAuthTokenCacheFile,
		}
		result.oauthManager = NewBrowserOAuthManager(oauthCfg)
	}

	return result, nil
}

// GetAPIKey returns the API key for the configured auth type.
// For api_key auth, returns the static API key.
// For oauth_browser auth, returns a valid OAuth access token (refreshing if needed).
func (c *Config) GetAPIKey(ctx context.Context) (string, error) {
	if c.AuthType == config.AuthOAuthBrowser && c.oauthManager != nil {
		return c.oauthManager.GetToken(ctx)
	}
	return c.APIKey, nil
}

// IsConfigured returns true if LLM is configured.
func IsConfigured(cfg config.GatekeeperConfig) bool {
	return cfg.Gatekeeper.LLM != nil && cfg.Gatekeeper.LLM.BaseURL != ""
}

// MaxRetries returns the maximum number of retry attempts for LLM requests.
func MaxRetries() int {
	return 2
}

// RuleBasedPillars returns the set of pillars that can be computed without LLM.
func RuleBasedPillars() []string {
	return []string{"static", "architecture", "verification", "security"}
}
