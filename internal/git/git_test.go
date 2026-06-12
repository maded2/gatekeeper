package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gatekeeper/internal/git"
)

// --- Story C-1: Check Quality of Changes Between Two Branches ---

// Helper: create a git repo with two branches and different files.
func setupGitRepo(t *testing.T) string {
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

func mustRun(t *testing.T, dir, cmd string, args ...string) {
	t.Helper()
	c := exec.Command(cmd, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "HOME="+dir)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %s (%s)", cmd, args, out, err)
	}
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

// ACCEPTANCE CRITERIA 1:
// "I can specify a base branch and a target branch to compare"
func TestDiff_GetChangedFiles(t *testing.T) {
	dir := setupGitRepo(t)

	changed, err := git.GetChangedFiles(dir, "main", "feature")
	if err != nil {
		t.Fatalf("GetChangedFiles returned error: %v", err)
	}

	if len(changed) == 0 {
		t.Error("expected changed files between main and feature")
	}

	found := false
	for _, f := range changed {
		if f == "main.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected main.go in changed files, got: %v", changed)
	}
}

// ACCEPTANCE CRITERIA 2:
// "Only the files and lines that differ between the two branches are analyzed"
func TestDiff_GetDiffContent(t *testing.T) {
	dir := setupGitRepo(t)

	diff, err := git.GetDiff(dir, "main", "feature")
	if err != nil {
		t.Fatalf("GetDiff returned error: %v", err)
	}

	if diff == "" {
		t.Error("expected non-empty diff between branches")
	}
}

// ACCEPTANCE CRITERIA 3:
// "The output shows how the changes affect each quality pillar"
// (Verified by evaluator integration, not git package directly)

// Verify error on non-existent branch
func TestDiff_NonExistentBranch(t *testing.T) {
	dir := setupGitRepo(t)

	_, err := git.GetChangedFiles(dir, "main", "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent branch")
	}
}
