package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gatekeeper/internal/config"
	"gatekeeper/internal/llm"
)

// --- oauth_browser Configuration Tests ---

func TestOAuthBrowserConfig_RequiresAuthURL(t *testing.T) {
	os.Setenv("TEST_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_OAUTH_CLIENT_SECRET", "test-client-secret")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_ID")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_SECRET")

	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		AuthType:               config.AuthOAuthBrowser,
		OAuthTokenURL:          "https://oauth2.googleapis.com/token",
		OAuthClientIDEnvVar:    "TEST_OAUTH_CLIENT_ID",
		OAuthClientSecretEnvVar: "TEST_OAUTH_CLIENT_SECRET",
		OAuthRedirectURL:       "http://localhost:8080/callback",
	}

	err := config.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for missing oauth_auth_url")
	}
	if !strings.Contains(err.Error(), "oauth_auth_url") {
		t.Errorf("expected error to mention oauth_auth_url: %v", err)
	}
}

func TestOAuthBrowserConfig_RequiresRedirectURL(t *testing.T) {
	os.Setenv("TEST_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_OAUTH_CLIENT_SECRET", "test-client-secret")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_ID")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_SECRET")

	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		AuthType:               config.AuthOAuthBrowser,
		OAuthTokenURL:          "https://oauth2.googleapis.com/token",
		OAuthAuthURL:           "https://accounts.google.com/o/oauth2/v2/auth",
		OAuthClientIDEnvVar:    "TEST_OAUTH_CLIENT_ID",
		OAuthClientSecretEnvVar: "TEST_OAUTH_CLIENT_SECRET",
	}

	err := config.Validate(&cfg)
	if err == nil {
		t.Fatal("expected error for missing oauth_redirect_url")
	}
	if !strings.Contains(err.Error(), "oauth_redirect_url") {
		t.Errorf("expected error to mention oauth_redirect_url: %v", err)
	}
}

func TestOAuthBrowserConfig_Valid(t *testing.T) {
	os.Setenv("TEST_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_OAUTH_CLIENT_SECRET", "test-client-secret")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_ID")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_SECRET")

	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		AuthType:               config.AuthOAuthBrowser,
		OAuthTokenURL:          "https://oauth2.googleapis.com/token",
		OAuthAuthURL:           "https://accounts.google.com/o/oauth2/v2/auth",
		OAuthClientIDEnvVar:    "TEST_OAUTH_CLIENT_ID",
		OAuthClientSecretEnvVar: "TEST_OAUTH_CLIENT_SECRET",
		OAuthScopes:            []string{"https://www.googleapis.com/auth/cloud-platform"},
		OAuthRedirectURL:       "http://localhost:8080/callback",
		OAuthTokenCacheFile:    "~/.cache/gatekeeper/oauth_token.json",
	}

	err := config.Validate(&cfg)
	if err != nil {
		t.Fatalf("expected no error for valid oauth_browser config, got: %v", err)
	}
}

func TestOAuthBrowserConfig_JSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gatekeeper.json")

	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		BaseURL:               "https://generativelanguage.googleapis.com/v1beta",
		ModelName:             "gemini-pro",
		AuthType:              config.AuthOAuthBrowser,
		OAuthTokenURL:         "https://oauth2.googleapis.com/token",
		OAuthAuthURL:          "https://accounts.google.com/o/oauth2/v2/auth",
		OAuthClientIDEnvVar:   "GOOGLE_CLIENT_ID",
		OAuthClientSecretEnvVar: "GOOGLE_CLIENT_SECRET",
		OAuthScopes:           []string{"https://www.googleapis.com/auth/cloud-platform"},
		OAuthRedirectURL:      "http://localhost:8080/callback",
		OAuthTokenCacheFile:   "~/.cache/gatekeeper/oauth_token.json",
		TimeoutMS:             5000,
		Temperature:           0,
	}

	err := config.Save(&cfg, path)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Gatekeeper.LLM.AuthType != config.AuthOAuthBrowser {
		t.Errorf("expected auth_type oauth_browser, got %s", loaded.Gatekeeper.LLM.AuthType)
	}
	if loaded.Gatekeeper.LLM.OAuthAuthURL != "https://accounts.google.com/o/oauth2/v2/auth" {
		t.Errorf("expected oauth_auth_url, got %s", loaded.Gatekeeper.LLM.OAuthAuthURL)
	}
	if loaded.Gatekeeper.LLM.OAuthRedirectURL != "http://localhost:8080/callback" {
		t.Errorf("expected redirect URL, got %s", loaded.Gatekeeper.LLM.OAuthRedirectURL)
	}
	if loaded.Gatekeeper.LLM.OAuthTokenCacheFile != "~/.cache/gatekeeper/oauth_token.json" {
		t.Errorf("expected token cache file, got %s", loaded.Gatekeeper.LLM.OAuthTokenCacheFile)
	}
}

// --- Token Persistence Tests ---

func TestTokenPersistence_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token.json")

	token := llm.OAuthToken{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "test-refresh-token",
	}

	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("failed to marshal token: %v", err)
	}
	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	manager := llm.NewBrowserOAuthManager(llm.OAuthConfig{
		TokenCacheFile: tokenFile,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	accessToken, err := manager.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessToken != "test-access-token" {
		t.Errorf("expected test-access-token, got %s", accessToken)
	}
}

func TestTokenPersistence_ExpiresAtPreserved(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token.json")

	// Token with a future expiry time stored as RFC3339
	futureExpiry := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	token := llm.OAuthToken{
		AccessToken:  "test-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "test-refresh",
		ExpiresAt:    futureExpiry,
	}

	data, _ := json.Marshal(token)
	os.WriteFile(tokenFile, data, 0600)

	manager := llm.NewBrowserOAuthManager(llm.OAuthConfig{
		TokenCacheFile: tokenFile,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	accessToken, err := manager.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessToken != "test-token" {
		t.Errorf("expected test-token, got %s", accessToken)
	}
}

func TestTokenRefresh_WithRefreshToken(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token.json")

	expiredToken := llm.OAuthToken{
		AccessToken:  "expired-access-token",
		TokenType:    "Bearer",
		ExpiresIn:    0,
		RefreshToken: "valid-refresh-token",
	}
	data, _ := json.Marshal(expiredToken)
	os.WriteFile(tokenFile, data, 0600)

	refreshCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Form.Get("grant_type") == "refresh_token" {
			refreshCalled = true
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "new-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	manager := llm.NewBrowserOAuthManager(llm.OAuthConfig{
		TokenURL:       server.URL,
		ClientID:       "test-client-id",
		ClientSecret:   "test-client-secret",
		TokenCacheFile: tokenFile,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accessToken, err := manager.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessToken != "new-access-token" {
		t.Errorf("expected new-access-token, got %s", accessToken)
	}
	if !refreshCalled {
		t.Error("expected refresh token flow to be called")
	}
}

func TestTokenCache_ReturnsCachedToken(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token.json")

	validToken := llm.OAuthToken{
		AccessToken: "cached-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}
	data, _ := json.Marshal(validToken)
	os.WriteFile(tokenFile, data, 0600)

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "new-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	manager := llm.NewBrowserOAuthManager(llm.OAuthConfig{
		TokenURL:       server.URL,
		ClientID:       "test-client-id",
		ClientSecret:   "test-client-secret",
		TokenCacheFile: tokenFile,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	accessToken, err := manager.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessToken != "cached-token" {
		t.Errorf("expected cached-token, got %s", accessToken)
	}
	if requestCount != 0 {
		t.Errorf("expected 0 server requests, got %d", requestCount)
	}
}

// --- FromConfig Tests ---

func TestFromConfig_OAuthBrowser(t *testing.T) {
	os.Setenv("TEST_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_OAUTH_CLIENT_SECRET", "test-client-secret")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_ID")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_SECRET")

	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		BaseURL:               "https://generativelanguage.googleapis.com/v1beta",
		ModelName:             "gemini-pro",
		AuthType:              config.AuthOAuthBrowser,
		OAuthTokenURL:         "https://oauth2.googleapis.com/token",
		OAuthAuthURL:          "https://accounts.google.com/o/oauth2/v2/auth",
		OAuthClientIDEnvVar:   "TEST_OAUTH_CLIENT_ID",
		OAuthClientSecretEnvVar: "TEST_OAUTH_CLIENT_SECRET",
		OAuthScopes:           []string{"https://www.googleapis.com/auth/cloud-platform"},
		OAuthRedirectURL:      "http://localhost:8080/callback",
		OAuthTokenCacheFile:   "~/.cache/gatekeeper/oauth_token.json",
		TimeoutMS:             5000,
		Temperature:           0,
	}

	llmCfg, err := llm.FromConfig(cfg)
	if err != nil {
		t.Fatalf("FromConfig returned error: %v", err)
	}

	if llmCfg.BaseURL != "https://generativelanguage.googleapis.com/v1beta" {
		t.Errorf("expected base URL, got %s", llmCfg.BaseURL)
	}
	if llmCfg.ModelName != "gemini-pro" {
		t.Errorf("expected model gemini-pro, got %s", llmCfg.ModelName)
	}
	if llmCfg.AuthType != config.AuthOAuthBrowser {
		t.Errorf("expected auth type oauth_browser, got %s", llmCfg.AuthType)
	}
}

func TestFromConfig_DefaultAuthType(t *testing.T) {
	os.Setenv("TEST_API_KEY", "test-api-key")
	defer os.Unsetenv("TEST_API_KEY")

	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		BaseURL:      "http://localhost:11434/v1",
		ModelName:    "llama3",
		APIKeyEnvVar: "TEST_API_KEY",
		TimeoutMS:    5000,
		Temperature:  0,
	}

	llmCfg, err := llm.FromConfig(cfg)
	if err != nil {
		t.Fatalf("FromConfig returned error: %v", err)
	}

	if llmCfg.AuthType != config.AuthAPIKey {
		t.Errorf("expected default auth type api_key, got %s", llmCfg.AuthType)
	}
	if llmCfg.APIKey != "test-api-key" {
		t.Errorf("expected API key test-api-key, got %s", llmCfg.APIKey)
	}
}

func TestGetAPIKey_APIKey(t *testing.T) {
	cfg := &llm.Config{
		APIKey:   "static-api-key",
		AuthType: config.AuthAPIKey,
	}

	key, err := cfg.GetAPIKey(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "static-api-key" {
		t.Errorf("expected static-api-key, got %s", key)
	}
}

func TestGetAPIKey_OAuthBrowser(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token.json")

	validToken := llm.OAuthToken{
		AccessToken: "cached-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}
	data, _ := json.Marshal(validToken)
	os.WriteFile(tokenFile, data, 0600)

	os.Setenv("TEST_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_OAUTH_CLIENT_SECRET", "test-client-secret")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_ID")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_SECRET")

	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		AuthType:              config.AuthOAuthBrowser,
		OAuthTokenURL:         "https://oauth2.googleapis.com/token",
		OAuthAuthURL:          "https://accounts.google.com/o/oauth2/v2/auth",
		OAuthClientIDEnvVar:   "TEST_OAUTH_CLIENT_ID",
		OAuthClientSecretEnvVar: "TEST_OAUTH_CLIENT_SECRET",
		OAuthRedirectURL:      "http://localhost:8080/callback",
		OAuthTokenCacheFile:   tokenFile,
	}

	llmCfg, err := llm.FromConfig(cfg)
	if err != nil {
		t.Fatalf("FromConfig returned error: %v", err)
	}

	key, err := llmCfg.GetAPIKey(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "cached-token" {
		t.Errorf("expected cached-token, got %s", key)
	}
}
