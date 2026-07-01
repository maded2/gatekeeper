package llm

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// OAuthConfig holds OAuth2 authorization code grant configuration.
type OAuthConfig struct {
	AuthURL        string
	TokenURL       string
	ClientID       string
	ClientSecret   string
	Scopes         []string
	RedirectURL    string
	TokenCacheFile string
}

// OAuthToken represents an OAuth2 access token with expiry information.
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
	expiresAt    time.Time
}

// IsExpired returns true if the token has expired or will expire within the grace period.
func (t *OAuthToken) IsExpired() bool {
	gracePeriod := 30 * time.Second
	return time.Now().After(t.expiresAt.Add(-gracePeriod))
}

// BrowserOAuthManager handles the browser-based OAuth2 authorization code flow.
type BrowserOAuthManager struct {
	mu       sync.Mutex
	token    *OAuthToken
	config   OAuthConfig
	client   *http.Client
}

// NewBrowserOAuthManager creates a new BrowserOAuthManager with the given OAuth configuration.
func NewBrowserOAuthManager(cfg OAuthConfig) *BrowserOAuthManager {
	return &BrowserOAuthManager{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken returns a valid (non-expired) access token.
// It loads cached tokens from file, refreshes if expired, or initiates a new browser auth flow.
func (m *BrowserOAuthManager) GetToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load cached token from file if not already loaded
	if m.token == nil {
		if err := m.loadToken(); err != nil {
			return "", fmt.Errorf("load cached token: %w", err)
		}
	}

	// Return cached token if still valid
	if m.token != nil && !m.token.IsExpired() {
		return m.token.AccessToken, nil
	}

	// Try to refresh the token if we have a refresh token
	if m.token != nil && m.token.RefreshToken != "" {
		newToken, err := m.refreshToken(ctx)
		if err == nil {
			m.token = newToken
			m.saveToken()
			return newToken.AccessToken, nil
		}
		// If refresh fails, fall through to full browser auth
	}

	// Full browser auth flow
	token, err := m.browserAuth(ctx)
	if err != nil {
		// If we have a cached token (even if expired), return it as a fallback
		if m.token != nil {
			return m.token.AccessToken, fmt.Errorf("token refresh failed (using expired token): %w", err)
		}
		return "", err
	}

	m.token = token
	m.saveToken()
	return token.AccessToken, nil
}

// buildAuthURL constructs the OAuth2 authorization URL with state parameter.
func (m *BrowserOAuthManager) buildAuthURL(state string) string {
	u, _ := url.Parse(m.config.AuthURL)
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", m.config.ClientID)
	q.Set("redirect_uri", m.config.RedirectURL)
	q.Set("scope", strings.Join(m.config.Scopes, " "))
	q.Set("state", state)
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")
	u.RawQuery = q.Encode()
	return u.String()
}

// oauthServer wraps an HTTP server for the browser login callback flow.
type oauthServer struct {
	server  *http.Server
	state   string
	done    chan string
}

// waitingPage is the HTML shown while the user completes login in their browser.
const waitingPage = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Gatekeeper — Signing In</title>
<style>body{font-family:system-ui,sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;background:#f5f5f5}
.card{background:#fff;padding:2rem;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,.1);text-align:center;max-width:360px}
h1{font-size:1.25rem;margin-bottom:.5rem}p{color:#555;font-size:.9rem}</style></head>
<body><div class="card"><h1>Waiting for sign-in</h1>
<p>Please complete authentication in your browser.</p></div></body></html>`

// successPage is the HTML shown after successful authentication.
const successPage = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Gatekeeper — Signed In</title>
<style>body{font-family:system-ui,sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;background:#f0fff4}
.card{background:#fff;padding:2rem;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,.1);text-align:center;max-width:360px}
h1{color:#16a34a;font-size:1.25rem;margin-bottom:.5rem}p{color:#555;font-size:.9rem}</style></head>
<body><div class="card"><h1>Signed in successfully</h1>
<p>You may close this window.</p></div></body></html>`

// errorPage is the HTML shown on authentication failure.
func errorPage(msg string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Gatekeeper — Sign-in Failed</title>
<style>body{font-family:system-ui,sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;background:#fef2f2}
.card{background:#fff;padding:2rem;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,.1);text-align:center;max-width:360px}
h1{color:#dc2626;font-size:1.25rem;margin-bottom:.5rem}p{color:#555;font-size:.9rem;word-break:break-word}</style></head>
<body><div class="card"><h1>Sign-in failed</h1><p>%s</p></div></body></html>`, msg)
}

// newOAuthServer creates a new oauthServer with the given state.
func newOAuthServer(state string) *oauthServer {
	return &oauthServer{
		state: state,
		done:  make(chan string, 1),
	}
}

// Start launches the HTTP server on the given address.
func (s *oauthServer) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.Handle)
	mux.HandleFunc("/callback", s.Handle)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return s.server.ListenAndServe()
}

// Handle processes incoming HTTP requests for the browser login flow.
func (s *oauthServer) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/callback":
		s.handleCallback(w, r)
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(waitingPage))
	}
}

// handleCallback processes the OAuth callback and extracts the authorization code.
func (s *oauthServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := q.Get("code")
	state := q.Get("state")
	error_ := q.Get("error")

	if error_ != "" {
		msg := q.Get("error_description")
		if msg == "" {
			msg = error_
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(errorPage(msg)))
		s.done <- ""
		return
	}

	if code == "" || state != s.state {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(errorPage("Invalid or missing authorization code.")))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(successPage))
	s.done <- code
}

// Stop gracefully shuts down the HTTP server.
func (s *oauthServer) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}

// browserAuth initiates the browser-based OAuth2 authorization code flow.
func (m *BrowserOAuthManager) browserAuth(ctx context.Context) (*OAuthToken, error) {
	// Generate a random state parameter
	state := generateState()

	// Build the authorization URL
	authURL := m.buildAuthURL(state)

	// Extract address from redirect URL
	u, err := url.Parse(m.config.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("parse redirect URL: %w", err)
	}

	addr := u.Host
	if addr == "" {
		addr = "127.0.0.1:0"
	}

	// Create and start the browser login server
	server := newOAuthServer(state)
	if err := server.Start(addr); err != nil {
		return nil, fmt.Errorf("start browser login server: %w", err)
	}

	// Update redirect URL with actual port if dynamic
	if u.Port() == "0" && server.server != nil {
		actualAddr := server.server.Addr
		u.Host = actualAddr
		m.config.RedirectURL = u.String()
		addr = actualAddr
	}

	// Rebuild auth URL with the correct redirect URL (after port resolution)
	authURL = m.buildAuthURL(state)

	// Print instructions
	fmt.Printf("Opening browser for sign-in...\n")
	fmt.Printf("If the browser does not open, visit:\n%s\n", authURL)

	// Open the browser
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser. Please visit:\n%s\n", authURL)
	}

	// Show local server status
	fmt.Printf("Local server running at http://%s — awaiting sign-in...\n", addr)

	// Wait for the callback
	select {
	case <-ctx.Done():
		_ = server.Stop()
		return nil, fmt.Errorf("sign-in timed out: %w", ctx.Err())
	case code := <-server.done:
		_ = server.Stop()
		if code == "" {
			return nil, fmt.Errorf("sign-in was cancelled or failed")
		}
		// Exchange the code for a token
		token, err := m.exchangeCode(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("exchange authorization code: %w", err)
		}
		return token, nil
	}
}

// exchangeCode exchanges an authorization code for an access token.
func (m *BrowserOAuthManager) exchangeCode(ctx context.Context, code string) (*OAuthToken, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", m.config.ClientID)
	form.Set("client_secret", m.config.ClientSecret)
	form.Set("redirect_uri", m.config.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.config.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	expiresIn := tokenResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3600
	}

	exp := time.Now().Add(time.Duration(expiresIn) * time.Second)
	return &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    expiresIn,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    exp.Format(time.RFC3339),
		expiresAt:    exp,
	}, nil
}

// refreshToken refreshes an expired token using the refresh token.
func (m *BrowserOAuthManager) refreshToken(ctx context.Context) (*OAuthToken, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", m.token.RefreshToken)
	form.Set("client_id", m.config.ClientID)
	form.Set("client_secret", m.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.config.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("refresh request returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}

	expiresIn := tokenResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3600
	}

	// Preserve existing refresh token if not provided
	refreshToken := tokenResp.RefreshToken
	if refreshToken == "" {
		refreshToken = m.token.RefreshToken
	}

	exp := time.Now().Add(time.Duration(expiresIn) * time.Second)
	return &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
		ExpiresAt:    exp.Format(time.RFC3339),
		expiresAt:    exp,
	}, nil
}

// loadToken loads a cached token from the token cache file.
func (m *BrowserOAuthManager) loadToken() error {
	if m.config.TokenCacheFile == "" {
		return nil
	}

	cacheFile, err := expandPath(m.config.TokenCacheFile)
	if err != nil {
		return fmt.Errorf("expand token cache path: %w", err)
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read token cache: %w", err)
	}

	var token OAuthToken
	if err := json.Unmarshal(data, &token); err != nil {
		return fmt.Errorf("parse token cache: %w", err)
	}

	// Restore expiry time from stored timestamp, falling back to ExpiresIn
	if token.ExpiresAt != "" {
		if parsed, err := time.Parse(time.RFC3339, token.ExpiresAt); err == nil {
			token.expiresAt = parsed
		}
	}
	if token.expiresAt.IsZero() && token.ExpiresIn > 0 {
		token.expiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	m.token = &token
	return nil
}

// saveToken saves the current token to the token cache file.
func (m *BrowserOAuthManager) saveToken() {
	if m.config.TokenCacheFile == "" || m.token == nil {
		return
	}

	cacheFile, err := expandPath(m.config.TokenCacheFile)
	if err != nil {
		return
	}

	// Ensure directory exists
	dir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}

	data, err := json.Marshal(m.token)
	if err != nil {
		return
	}

	if err := os.WriteFile(cacheFile, data, 0600); err != nil {
		// Non-fatal: token caching is best-effort
	}
}

// generateState generates a random state parameter for CSRF protection.
func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// expandPath expands a ~ to the user's home directory.
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home directory: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

// openBrowser opens the default browser to the given URL.
func openBrowser(u string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{u}
	case "darwin":
		cmd = "open"
		args = []string{u}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", u}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}
