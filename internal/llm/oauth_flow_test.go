package llm_test

// Acceptance tests for the browser-based OAuth sign-in flow (Gemini OAuth).
//
// Story: As a developer using an OAuth-authenticated LLM provider, when no
// valid cached token exists, then Gatekeeper completes the browser sign-in
// and returns an access token — without hanging.
//
// The printed sign-in URL is the user-facing contract for completing login
// (it is what a human visits when the browser does not open automatically),
// so these tests drive the flow through that observable surface plus the
// local callback endpoint.

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gatekeeper/internal/llm"
)

// --- helpers ---

type tokenResult struct {
	token string
	err   error
}

// startGetToken runs GetToken in the background with os.Stdout captured, so
// the test can observe the printed sign-in URL. Returns the result channel
// and a reader over the captured stdout.
func startGetToken(t *testing.T, mgr *llm.BrowserOAuthManager, timeout time.Duration) (<-chan tokenResult, *bufio.Reader) {
	t.Helper()

	oldStdout := os.Stdout
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = pw

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resCh := make(chan tokenResult, 1)
	go func() {
		tok, err := mgr.GetToken(ctx)
		resCh <- tokenResult{token: tok, err: err}
	}()

	t.Cleanup(func() {
		cancel()
		os.Stdout = oldStdout
		pw.Close()
	})

	return resCh, bufio.NewReader(pr)
}

// waitForSignInURL reads captured stdout until a line parses as an HTTP(S)
// URL carrying a "state" query parameter (the printed sign-in URL).
func waitForSignInURL(t *testing.T, r *bufio.Reader, timeout time.Duration) *url.URL {
	t.Helper()

	lineCh := make(chan string, 32)
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			lineCh <- scanner.Text()
		}
		close(lineCh)
	}()

	deadline := time.After(timeout)
	for {
		select {
		case line, ok := <-lineCh:
			if !ok {
				t.Fatal("stdout closed before the sign-in URL was printed")
			}
			u, err := url.Parse(strings.TrimSpace(line))
			if err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Query().Get("state") != "" {
				return u
			}
		case <-deadline:
			t.Fatalf("timed out after %s waiting for the sign-in URL on stdout — the sign-in flow hung", timeout)
		}
	}
}

// waitHTTPUp polls the given URL until it answers with any HTTP response.
func waitHTTPUp(t *testing.T, target string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(target) //nolint:gosec // test-only, local URL
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("local callback server at %s did not come up within %s", target, timeout)
}

// mockTokenServer stands in for the provider's OAuth token endpoint.
type mockTokenServer struct {
	*httptest.Server
	form url.Values
}

func newMockTokenServer(t *testing.T, accessToken string) *mockTokenServer {
	t.Helper()
	m := &mockTokenServer{}
	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse token request form: %v", err)
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		m.form = r.Form
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	t.Cleanup(m.Server.Close)
	return m
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func newFlowManager(t *testing.T, mock *mockTokenServer, redirectURL, cacheFile string) *llm.BrowserOAuthManager {
	t.Helper()
	return llm.NewBrowserOAuthManager(llm.OAuthConfig{
		AuthURL:        "https://accounts.provider.example/o/oauth2/v2/auth", // never fetched by the client; the user's browser would visit it
		TokenURL:       mock.URL,
		ClientID:       "test-client-id",
		ClientSecret:   "test-client-secret",
		Scopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
		RedirectURL:    redirectURL,
		TokenCacheFile: cacheFile,
	})
}

// --- acceptance tests ---

// Given no cached token and a dynamic (port 0) redirect URL,
// when the user completes sign-in in their browser (callback with code+state),
// then Gatekeeper exchanges the code at the token endpoint and returns the
// access token, using the actually-bound callback port in the sign-in URL.
func TestBrowserAuth_CompletesSignInAndReturnsToken(t *testing.T) {
	mock := newMockTokenServer(t, "gemini-oauth-access-token")
	cacheFile := filepath.Join(t.TempDir(), "token.json")

	mgr := newFlowManager(t, mock, "http://127.0.0.1:0/callback", cacheFile)
	resCh, stdout := startGetToken(t, mgr, 10*time.Second)

	signInURL := waitForSignInURL(t, stdout, 5*time.Second)
	state := signInURL.Query().Get("state")
	redirectURI := signInURL.Query().Get("redirect_uri")
	if redirectURI == "" {
		t.Fatal("sign-in URL is missing the redirect_uri parameter")
	}

	// The advertised callback must use the real bound port, not "0".
	cb, err := url.Parse(redirectURI)
	if err != nil {
		t.Fatalf("parse redirect_uri %q: %v", redirectURI, err)
	}
	if cb.Port() == "" || cb.Port() == "0" {
		t.Fatalf("redirect_uri %q does not carry the actually-bound port", redirectURI)
	}

	// Simulate the user's browser completing sign-in at Google:
	// Google redirects to the local callback with code and state.
	resp, err := http.Get(redirectURI + "?code=auth-code-123&state=" + url.QueryEscape(state)) //nolint:gosec // local test URL
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("callback returned status %d, body: %s", resp.StatusCode, body)
	}

	select {
	case res := <-resCh:
		if res.err != nil {
			t.Fatalf("GetToken failed after successful sign-in: %v", res.err)
		}
		if res.token != "gemini-oauth-access-token" {
			t.Errorf("expected gemini-oauth-access-token, got %q", res.token)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("GetToken did not return within 5s after the callback arrived")
	}

	// The authorization code must have been exchanged at the token endpoint.
	if mock.form == nil {
		t.Fatal("token endpoint was never called")
	}
	if got := mock.form.Get("grant_type"); got != "authorization_code" {
		t.Errorf("expected grant_type authorization_code, got %q", got)
	}
	if got := mock.form.Get("code"); got != "auth-code-123" {
		t.Errorf("expected code auth-code-123, got %q", got)
	}
	if got := mock.form.Get("client_id"); got != "test-client-id" {
		t.Errorf("expected client_id test-client-id, got %q", got)
	}

	// The token must be cached for subsequent non-interactive runs.
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("token cache file not written: %v", err)
	}
	var cached llm.OAuthToken
	if err := json.Unmarshal(data, &cached); err != nil {
		t.Fatalf("parse token cache: %v", err)
	}
	if cached.AccessToken != "gemini-oauth-access-token" {
		t.Errorf("cached access token = %q, want gemini-oauth-access-token", cached.AccessToken)
	}
}

// Given no cached token,
// when the callback arrives with a state that does not match (CSRF protection),
// then the sign-in fails fast with an error instead of hanging until timeout.
func TestBrowserAuth_StateMismatchFailsFast(t *testing.T) {
	mock := newMockTokenServer(t, "should-not-be-used")
	port := freePort(t)
	mgr := newFlowManager(t, mock, fmt.Sprintf("http://127.0.0.1:%d/callback", port), filepath.Join(t.TempDir(), "token.json"))

	resCh, _ := startGetToken(t, mgr, 10*time.Second)

	waitHTTPUp(t, fmt.Sprintf("http://127.0.0.1:%d/", port), 5*time.Second)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/callback?code=abc&state=WRONG-STATE", port)) //nolint:gosec // local test URL
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	resp.Body.Close()

	select {
	case res := <-resCh:
		if res.err == nil {
			t.Fatal("expected an error for state mismatch, got a token")
		}
		if res.token != "" {
			t.Errorf("expected no token on failure, got %q", res.token)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("state mismatch did not fail fast — the flow hung instead of reporting the invalid callback")
	}

	if mock.form != nil {
		t.Error("token endpoint must not be called when the state does not match")
	}
}

// Given no cached token,
// when the user cancels at the provider's consent screen (error=access_denied),
// then the sign-in fails promptly with a clear error and no token request is made.
func TestBrowserAuth_UserCancelAtConsentFailsCleanly(t *testing.T) {
	mock := newMockTokenServer(t, "should-not-be-used")
	cacheFile := filepath.Join(t.TempDir(), "token.json")

	mgr := newFlowManager(t, mock, "http://127.0.0.1:0/callback", cacheFile)
	resCh, stdout := startGetToken(t, mgr, 10*time.Second)

	signInURL := waitForSignInURL(t, stdout, 5*time.Second)
	state := signInURL.Query().Get("state")
	redirectURI := signInURL.Query().Get("redirect_uri")
	if redirectURI == "" {
		t.Fatal("sign-in URL is missing the redirect_uri parameter")
	}

	// Simulate Google redirecting back with an error (user clicked "Cancel").
	resp, err := http.Get(redirectURI + "?error=access_denied&error_description=The+user+cancelled&state=" + url.QueryEscape(state)) //nolint:gosec // local test URL
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	resp.Body.Close()

	select {
	case res := <-resCh:
		if res.err == nil {
			t.Fatal("expected an error when the user cancels consent, got a token")
		}
		if !strings.Contains(strings.ToLower(res.err.Error()), "cancel") {
			t.Errorf("error should tell the user sign-in was cancelled, got: %v", res.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("user cancellation did not fail fast — the flow hung")
	}

	if mock.form != nil {
		t.Error("token endpoint must not be called when the user cancels")
	}
	if _, err := os.Stat(cacheFile); !os.IsNotExist(err) {
		t.Error("no token cache file should be written after a cancelled sign-in")
	}
}

// Given a redirect URL whose port is already in use,
// when Gatekeeper starts the sign-in flow,
// then it returns a clear startup error promptly instead of hanging.
func TestBrowserAuth_PortInUseReturnsClearError(t *testing.T) {
	mock := newMockTokenServer(t, "should-not-be-used")

	// Occupy the port so the callback server cannot bind it.
	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer blocker.Close()
	port := blocker.Addr().(*net.TCPAddr).Port

	mgr := newFlowManager(t, mock, fmt.Sprintf("http://127.0.0.1:%d/callback", port), filepath.Join(t.TempDir(), "token.json"))
	resCh, _ := startGetToken(t, mgr, 5*time.Second)

	select {
	case res := <-resCh:
		if res.err == nil {
			t.Fatal("expected an error when the callback port is in use")
		}
		if !strings.Contains(res.err.Error(), "start browser login server") {
			t.Errorf("error should identify the local server startup failure, got: %v", res.err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("port-in-use condition did not surface promptly — the flow hung")
	}
}
