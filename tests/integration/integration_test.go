// Package integration tests the Gatekeeper CLI end-to-end.
//
// These tests build the binary and run it against test workspaces
// to verify the full command flow including config loading,
// evaluation, reporting, and exit codes.
package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath returns the path to the built gatekeeper binary.
func binaryPath(t *testing.T) string {
	t.Helper()

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "gatekeeper-test-bin", ".")
	cmd.Dir = workspaceRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build gatekeeper: %s (%v)", out, err)
	}

	return filepath.Join(workspaceRoot(t), "gatekeeper-test-bin")
}

// workspaceRoot returns the project root directory by finding go.mod.
func workspaceRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Walk up until we find go.mod
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("could not find project root (go.mod not found)")
	return ""
}

// runGatekeeper runs the gatekeeper binary with the given arguments
// in the specified directory and returns stdout, stderr, and exit code.
func runGatekeeper(t *testing.T, bin, dir string, args ...string) (string, string, int) {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run gatekeeper: %v", err)
		}
	}

	return stdout.String(), stderr.String(), exitCode
}

// setupTestWorkspace creates a temporary workspace with sample Go files.
func setupTestWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create sample Go files
	mustWrite(t, dir, "main.go", `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`)
	mustWrite(t, dir, "auth/auth.go", `package auth

func Authenticate(user, pass string) bool {
	if user == "" || pass == "" {
		return false
	}
	return user == "admin" && pass == "secret"
}
`)
	mustWrite(t, dir, "README.md", "# Test Project")

	return dir
}

// setupGitWorkspace creates a git repo with branches for diff testing.
func setupGitWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Initialize git repo
	mustRun(t, dir, "git", "init")
	mustRun(t, dir, "git", "config", "user.email", "test@test.com")
	mustRun(t, dir, "git", "config", "user.name", "Test")

	// Create initial file on main
	mustWrite(t, dir, "main.go", `package main

func main() {}
`)
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "initial")

	// Create feature branch
	mustRun(t, dir, "git", "checkout", "-b", "feature")

	// Modify file on feature branch
	mustWrite(t, dir, "main.go", `package main

func main() {
	hello()
}

func hello() {
	println("hello")
}
`)
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "add hello function")

	return dir
}

func mustWrite(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustRun(t *testing.T, dir, cmd string, args ...string) {
	t.Helper()
	c := exec.Command(cmd, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "HOME="+dir)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %s (%v)", cmd, args, out, err)
	}
}

// --- Integration Tests ---

// TestInit_Command tests the 'gatekeeper init' command.
func TestInit_Command(t *testing.T) {
	dir := t.TempDir()
	bin := binaryPath(t)

	_, stderr, exitCode := runGatekeeper(t, bin, dir, "init")
	if exitCode != 0 {
		t.Fatalf("init failed with exit code %d: %s", exitCode, stderr)
	}

	// Verify gatekeeper.json was created
	cfgPath := filepath.Join(dir, "gatekeeper.json")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("gatekeeper.json not created: %v", err)
	}

	// Verify config is valid JSON
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read gatekeeper.json: %v", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("gatekeeper.json is not valid JSON: %v", err)
	}

	// Verify default threshold
	gk := cfg["gatekeeper"].(map[string]interface{})
	if gk["target_threshold"].(float64) != 75 {
		t.Errorf("expected default threshold of 75, got %v", gk["target_threshold"])
	}
}

// TestCheck_Command tests the 'gatekeeper check' command.
func TestCheck_Command(t *testing.T) {
	dir := setupTestWorkspace(t)
	bin := binaryPath(t)

	// First create config
	_, _, exitCode := runGatekeeper(t, bin, dir, "init")
	if exitCode != 0 {
		t.Fatalf("init failed: %d", exitCode)
	}

	// Run check
	stdout, stderr, exitCode := runGatekeeper(t, bin, dir, "check")
	if exitCode != 0 && exitCode != 2 {
		t.Fatalf("check failed with exit code %d: %s", exitCode, stderr)
	}

	// Verify output contains score
	if !strings.Contains(stdout, "Quality Score") {
		t.Errorf("expected 'Quality Score' in output: %s", stdout)
	}

	// Verify output contains pillar breakdown
	if !strings.Contains(stdout, "Pillar Breakdown") {
		t.Errorf("expected 'Pillar Breakdown' in output: %s", stdout)
	}
}

// TestCheck_JSON_Output tests the 'gatekeeper check --format=json' command.
func TestCheck_JSON_Output(t *testing.T) {
	dir := setupTestWorkspace(t)
	bin := binaryPath(t)

	// Create config
	runGatekeeper(t, bin, dir, "init")

	// Run check with JSON output
	stdout, stderr, exitCode := runGatekeeper(t, bin, dir, "check", "--format=json")
	if exitCode != 0 && exitCode != 2 {
		t.Fatalf("check failed with exit code %d: %s", exitCode, stderr)
	}

	// Verify output is valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, stdout)
	}

	// Verify required fields
	if _, ok := result["total"]; !ok {
		t.Error("expected 'total' field in JSON output")
	}
	if _, ok := result["pillars"]; !ok {
		t.Error("expected 'pillars' field in JSON output")
	}
	if _, ok := result["findings"]; !ok {
		t.Error("expected 'findings' field in JSON output")
	}
	if _, ok := result["timestamp"]; !ok {
		t.Error("expected 'timestamp' field in JSON output")
	}
}

// TestCheck_Markdown_Output tests the 'gatekeeper check --format=markdown' command.
func TestCheck_Markdown_Output(t *testing.T) {
	dir := setupTestWorkspace(t)
	bin := binaryPath(t)

	// Create config
	runGatekeeper(t, bin, dir, "init")

	// Run check with markdown output
	stdout, stderr, exitCode := runGatekeeper(t, bin, dir, "check", "--format=markdown")
	if exitCode != 0 && exitCode != 2 {
		t.Fatalf("check failed with exit code %d: %s", exitCode, stderr)
	}

	// Verify markdown formatting
	if !strings.Contains(stdout, "##") {
		t.Error("expected markdown headers in output")
	}
	if !strings.Contains(stdout, "|") {
		t.Error("expected markdown tables in output")
	}
}

// TestCheck_Exit_Code_Pass tests that a passing score returns exit code 0.
func TestCheck_Exit_Code_Pass(t *testing.T) {
	dir := setupTestWorkspace(t)
	bin := binaryPath(t)

	// Create config with low threshold
	runGatekeeper(t, bin, dir, "init")

	// Modify config to have low threshold
	cfgPath := filepath.Join(dir, "gatekeeper.json")
	cfg, _ := os.ReadFile(cfgPath)
	cfg = []byte(strings.Replace(string(cfg), `"target_threshold": 75`, `"target_threshold": 0`, 1))
	os.WriteFile(cfgPath, cfg, 0644)

	// Run check
	_, _, exitCode := runGatekeeper(t, bin, dir, "check")
	if exitCode != 0 {
		t.Errorf("expected exit code 0 (pass), got %d", exitCode)
	}
}

// TestCheck_Exit_Code_Fail tests that a failing score returns exit code 2.
func TestCheck_Exit_Code_Fail(t *testing.T) {
	dir := setupTestWorkspace(t)
	bin := binaryPath(t)

	// Create config
	runGatekeeper(t, bin, dir, "init")

	// Modify config to have threshold above 100 (will fail since max score is 100)
	cfgPath := filepath.Join(dir, "gatekeeper.json")
	cfg, _ := os.ReadFile(cfgPath)
	cfg = []byte(strings.Replace(string(cfg), `"target_threshold": 75`, `"target_threshold": 99`, 1))
	os.WriteFile(cfgPath, cfg, 0644)

	// Run check - should pass since workspace has perfect score
	// Instead, create a workspace with issues
	mustWrite(t, dir, "bad_code.go", `package bad

var secret = "sk-1234567890abcdef"`) // triggers security finding

	_, _, exitCode := runGatekeeper(t, bin, dir, "check")
	// With security finding, score drops below 99
	if exitCode != 2 {
		t.Errorf("expected exit code 2 (fail), got %d", exitCode)
	}
}

// TestCheck_Exit_Code_Error tests that a missing config returns exit code 1.
func TestCheck_Exit_Code_Error(t *testing.T) {
	dir := setupTestWorkspace(t)
	bin := binaryPath(t)

	// Run check without config
	_, stderr, exitCode := runGatekeeper(t, bin, dir, "check")
	if exitCode != 1 {
		t.Errorf("expected exit code 1 (error), got %d", exitCode)
	}
	if !strings.Contains(stderr, "gatekeeper.json") {
		t.Errorf("expected error message to mention gatekeeper.json: %s", stderr)
	}
}

// TestDiff_Command tests the 'gatekeeper diff' command.
func TestDiff_Command(t *testing.T) {
	dir := setupGitWorkspace(t)
	bin := binaryPath(t)

	// Create config
	runGatekeeper(t, bin, dir, "init")

	// Run diff
	stdout, stderr, exitCode := runGatekeeper(t, bin, dir, "diff", "--base=main", "--target=feature")
	if exitCode != 0 && exitCode != 2 {
		t.Fatalf("diff failed with exit code %d: %s", exitCode, stderr)
	}

	// Verify output contains score
	if !strings.Contains(stdout, "Quality Score") {
		t.Errorf("expected 'Quality Score' in output: %s", stdout)
	}
}

// TestCommitRange_Command tests the 'gatekeeper commit-range' command.
func TestCommitRange_Command(t *testing.T) {
	dir := setupGitWorkspace(t)
	bin := binaryPath(t)

	// Create config
	runGatekeeper(t, bin, dir, "init")

	// Run commit-range
	stdout, stderr, exitCode := runGatekeeper(t, bin, dir, "commit-range", "--range=HEAD~1..HEAD")
	if exitCode != 0 && exitCode != 2 {
		t.Fatalf("commit-range failed with exit code %d: %s", exitCode, stderr)
	}

	// Verify output contains score
	if !strings.Contains(stdout, "Quality Score") {
		t.Errorf("expected 'Quality Score' in output: %s", stdout)
	}
}

// TestTrivial_Change_Skip tests that trivial changes are skipped.
func TestTrivial_Change_Skip(t *testing.T) {
	dir := setupGitWorkspace(t)
	bin := binaryPath(t)

	// Create config
	runGatekeeper(t, bin, dir, "init")

	// Create a new branch with only README changes
	mustRun(t, dir, "git", "checkout", "main")
	mustRun(t, dir, "git", "checkout", "-b", "docs-only")

	mustWrite(t, dir, "README.md", "# Updated README")
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "update readme")

	// Run diff - should skip trivial changes
	stdout, _, exitCode := runGatekeeper(t, bin, dir, "diff", "--base=main", "--target=docs-only")
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for trivial changes, got %d", exitCode)
	}
	if !strings.Contains(stdout, "trivial") && !strings.Contains(stdout, "No changes") {
		t.Errorf("expected 'trivial' or 'No changes' in output: %s", stdout)
	}
}

// TestCheck_Output_File tests writing JSON output to a file.
func TestCheck_Output_File(t *testing.T) {
	dir := setupTestWorkspace(t)
	bin := binaryPath(t)

	// Create config
	runGatekeeper(t, bin, dir, "init")

	// Run check with output file
	outputFile := filepath.Join(dir, "report.json")
	_, stderr, exitCode := runGatekeeper(t, bin, dir, "check", "--format=json", "--output="+outputFile)
	if exitCode != 0 && exitCode != 2 {
		t.Fatalf("check failed with exit code %d: %s", exitCode, stderr)
	}

	// Verify output file exists
	if _, err := os.Stat(outputFile); err != nil {
		t.Fatalf("output file not created: %v", err)
	}

	// Verify output file is valid JSON
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output file is not valid JSON: %v", err)
	}
}
