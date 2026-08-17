package llm_test

// Acceptance tests for the LLM chat client (OpenAI-compatible endpoint).
//
// Story: As a developer with a configured LLM provider (e.g. Google Gemini
// via OAuth), when an evaluation runs, then Gatekeeper transmits scrubbed
// code to the provider and applies the returned pillar adjustments and
// remediations; on failure it retries and then reports an error so callers
// can fall back to rule-based scoring (Story G-2).

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gatekeeper/internal/config"
	"gatekeeper/internal/llm"
)

// mockOpenAI stands in for an OpenAI-compatible /chat/completions endpoint.
type mockOpenAI struct {
	*httptest.Server
	requests  atomic.Int64
	lastAuth  string
	lastPath  string
	lastBody  []byte
	content   string // assistant message content returned on success
	failFirst int    // number of initial requests to fail with 500
}

func newMockOpenAI(t *testing.T, content string, failFirst int) *mockOpenAI {
	t.Helper()
	m := &mockOpenAI{content: content, failFirst: failFirst}
	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.requests.Add(1)
		m.lastAuth = r.Header.Get("Authorization")
		m.lastPath = r.URL.Path
		m.lastBody, _ = io.ReadAll(r.Body)

		if int(m.requests.Load()) <= m.failFirst {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 0,
			"model":   "mock-model",
			"choices": []map[string]interface{}{
				{
					"index":         0,
					"message":       map[string]string{"role": "assistant", "content": m.content},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
		})
	}))
	t.Cleanup(m.Server.Close)
	return m
}

func validEvaluationJSON() string {
	return `{"pillar_adjustments":{"static_health_deduction":5.0,"architecture_deduction":2.5},"remediations":[{"priority":"HIGH","pillar":"static","location":"main.go:lines 3-9","finding":"Poor function naming","actionable_fix":"Rename to descriptive names"}]}`
}

func apiKeyConfig(mock *mockOpenAI) *llm.Config {
	return &llm.Config{
		BaseURL:     mock.URL,
		ModelName:   "gemini-2.0-flash",
		APIKey:      "test-key",
		TimeoutMS:   5000,
		Temperature: 0,
		AuthType:    config.AuthAPIKey,
	}
}

func allowPrivacy() config.PrivacyConfig {
	allow := true
	return config.PrivacyConfig{AllowPublicCloudTransmission: &allow, DataScrubbing: nil}
}

// --- acceptance tests ---

// Given a configured LLM with API key auth,
// when Evaluate is called,
// then an authenticated chat completion request is sent to {base_url}/chat/completions
// carrying the configured model and temperature, and the structured JSON
// response is parsed into pillar adjustments and remediations.
func TestClient_Evaluate_SendsAuthenticatedRequestAndParsesEvaluation(t *testing.T) {
	mock := newMockOpenAI(t, validEvaluationJSON(), 0)
	client := llm.NewClient(apiKeyConfig(mock))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	eval, err := client.Evaluate(ctx, allowPrivacy(), "package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if got := mock.lastAuth; got != "Bearer test-key" {
		t.Errorf("Authorization header = %q, want %q", got, "Bearer test-key")
	}
	if got := mock.lastPath; got != "/chat/completions" {
		t.Errorf("request path = %q, want /chat/completions", got)
	}

	body := string(mock.lastBody)
	if !strings.Contains(body, `"model":"gemini-2.0-flash"`) {
		t.Errorf("request body missing configured model: %s", body)
	}
	if !strings.Contains(body, "func main()") {
		t.Error("request body does not contain the code under evaluation")
	}

	if eval.PillarAdjustments.StaticHealthDeduction != 5.0 {
		t.Errorf("static_health_deduction = %v, want 5.0", eval.PillarAdjustments.StaticHealthDeduction)
	}
	if eval.PillarAdjustments.ArchitectureDeduction != 2.5 {
		t.Errorf("architecture_deduction = %v, want 2.5", eval.PillarAdjustments.ArchitectureDeduction)
	}
	if len(eval.Remediations) != 1 {
		t.Fatalf("expected 1 remediation, got %d", len(eval.Remediations))
	}
	r := eval.Remediations[0]
	if r.Priority != "HIGH" || r.Pillar != "static" || r.Finding != "Poor function naming" || r.ActionableFix != "Rename to descriptive names" {
		t.Errorf("unexpected remediation: %+v", r)
	}
}

// Given a configured LLM with oauth_browser auth and a valid cached token,
// when Evaluate is called,
// then the OAuth access token (not an API key) is used as the bearer credential.
func TestClient_Evaluate_OAuthTokenUsedForAuth(t *testing.T) {
	mock := newMockOpenAI(t, validEvaluationJSON(), 0)

	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token.json")
	validToken := llm.OAuthToken{AccessToken: "gemini-oauth-token", TokenType: "Bearer", ExpiresIn: 3600}
	data, _ := json.Marshal(validToken)
	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	os.Setenv("TEST_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_OAUTH_CLIENT_SECRET", "test-client-secret")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_ID")
	defer os.Unsetenv("TEST_OAUTH_CLIENT_SECRET")

	cfg := config.DefaultConfig()
	cfg.Gatekeeper.LLM = &config.LLMConfig{
		BaseURL:                 mock.URL,
		ModelName:               "gemini-2.0-flash",
		AuthType:                config.AuthOAuthBrowser,
		OAuthTokenURL:           mock.URL + "/token",
		OAuthAuthURL:            "https://accounts.provider.example/auth",
		OAuthClientIDEnvVar:     "TEST_OAUTH_CLIENT_ID",
		OAuthClientSecretEnvVar: "TEST_OAUTH_CLIENT_SECRET",
		OAuthRedirectURL:        "http://127.0.0.1:9/callback",
		OAuthTokenCacheFile:     tokenFile,
		TimeoutMS:               5000,
	}

	llmCfg, err := llm.FromConfig(cfg)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := llm.NewClient(llmCfg).Evaluate(ctx, allowPrivacy(), "func main() {}"); err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if got := mock.lastAuth; got != "Bearer gemini-oauth-token" {
		t.Errorf("Authorization header = %q, want %q (OAuth access token)", got, "Bearer gemini-oauth-token")
	}
}

// Given data scrubbing is enabled (the default),
// when Evaluate transmits code containing a secret pattern,
// then the secret is redacted in the request body.
func TestClient_Evaluate_ScrubsSecretsBeforeTransmission(t *testing.T) {
	mock := newMockOpenAI(t, validEvaluationJSON(), 0)
	client := llm.NewClient(apiKeyConfig(mock))

	secretCode := "package main\n\nvar apiKey = \"sk-abcdefgh1234567890\"\n"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := client.Evaluate(ctx, allowPrivacy(), secretCode); err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	body := string(mock.lastBody)
	if strings.Contains(body, "sk-abcdefgh1234567890") {
		t.Error("raw secret was transmitted — scrubbing did not run")
	}
	if !strings.Contains(body, "[REDACTED]") {
		t.Error("expected [REDACTED] marker in transmitted code")
	}
}

// Given allow_public_cloud_transmission is false (air-gapped mode),
// when Evaluate is called,
// then no network request is made and an air-gap error is returned.
func TestClient_Evaluate_AirGapBlocksTransmission(t *testing.T) {
	mock := newMockOpenAI(t, validEvaluationJSON(), 0)
	client := llm.NewClient(apiKeyConfig(mock))

	disallow := false
	privacy := config.PrivacyConfig{AllowPublicCloudTransmission: &disallow}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Evaluate(ctx, privacy, "func main() {}")
	if err == nil {
		t.Fatal("expected an error when cloud transmission is disallowed")
	}
	if !errors.Is(err, llm.ErrAirGapped) {
		t.Errorf("expected ErrAirGapped, got: %v", err)
	}
	if n := mock.requests.Load(); n != 0 {
		t.Errorf("expected 0 requests in air-gapped mode, got %d", n)
	}
}

// Given the provider fails transiently (HTTP 500),
// when Evaluate is called,
// then it retries up to MaxRetries() additional times and succeeds once the provider recovers.
func TestClient_Evaluate_RetriesThenSucceeds(t *testing.T) {
	mock := newMockOpenAI(t, validEvaluationJSON(), 2) // fail first 2, succeed on 3rd
	client := llm.NewClient(apiKeyConfig(mock))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	eval, err := client.Evaluate(ctx, allowPrivacy(), "func main() {}")
	if err != nil {
		t.Fatalf("Evaluate should have succeeded after retries: %v", err)
	}
	if eval.PillarAdjustments.StaticHealthDeduction != 5.0 {
		t.Errorf("unexpected evaluation after retry: %+v", eval)
	}
	if n := mock.requests.Load(); n != int64(1+llm.MaxRetries()) {
		t.Errorf("expected %d attempts, got %d", 1+llm.MaxRetries(), n)
	}
}

// Given the provider keeps failing,
// when all retries are exhausted,
// then Evaluate returns an error after exactly 1 + MaxRetries() attempts.
func TestClient_Evaluate_AllRetriesFailReturnsError(t *testing.T) {
	mock := newMockOpenAI(t, validEvaluationJSON(), 99)
	client := llm.NewClient(apiKeyConfig(mock))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Evaluate(ctx, allowPrivacy(), "func main() {}")
	if err == nil {
		t.Fatal("expected an error after all retries failed")
	}
	if n := mock.requests.Load(); n != int64(1+llm.MaxRetries()) {
		t.Errorf("expected %d attempts, got %d", 1+llm.MaxRetries(), n)
	}
}

// Given the provider returns a non-JSON assistant message,
// when Evaluate parses the response,
// then it returns a parse error (never panics).
func TestClient_Evaluate_MalformedJSONReturnsError(t *testing.T) {
	mock := newMockOpenAI(t, "I cannot evaluate this code, sorry.", 0)
	client := llm.NewClient(apiKeyConfig(mock))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Evaluate(ctx, allowPrivacy(), "func main() {}")
	if err == nil {
		t.Fatal("expected a parse error for non-JSON model output")
	}
}

// Given the provider times out on every attempt,
// when Evaluate is called with a per-request timeout,
// then it returns an error within a bounded time.
func TestClient_Evaluate_TimeoutBounded(t *testing.T) {
	// A server that never responds in time. The handler exits when the client
	// gives up (or after a safety bound) so Server.Close() in cleanup cannot hang.
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	t.Cleanup(slow.Close)

	cfg := apiKeyConfig(&mockOpenAI{Server: slow})
	cfg.TimeoutMS = 300
	client := llm.NewClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	start := time.Now()
	_, err := client.Evaluate(ctx, allowPrivacy(), "func main() {}")
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if elapsed > 10*time.Second {
		t.Errorf("Evaluate took %s; expected bounded retries with %dms per-request timeout", elapsed, cfg.TimeoutMS)
	}
}
