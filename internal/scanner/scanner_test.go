package scanner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gatekeeper/internal/config"
	"gatekeeper/internal/scanner"
)

// --- Story A-2: Exclude Unwanted Paths from Analysis ---

// Helper: create a temp directory with sample files for scanning.
func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create source files
	mustWrite(t, dir, "src/main.go", "package main")
	mustWrite(t, dir, "src/auth.go", "package auth")
	mustWrite(t, dir, "vendor/lib.go", "package vendor")
	mustWrite(t, dir, "node_modules/pkg/index.js", "module.exports={}")
	mustWrite(t, dir, "dist/bundle.js", "/* bundled */")
	mustWrite(t, dir, "internal/generated/code.go", "package generated")
	mustWrite(t, dir, "README.md", "# docs")

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

// ACCEPTANCE CRITERIA 1:
// "I can list path patterns in my configuration"
func TestScanner_UsesConfigExclusions(t *testing.T) {
	dir := setupTestDir(t)

	cfg := config.DefaultConfig()
	cfg.Exclusions.Paths = append(cfg.Exclusions.Paths, "**/generated/*.go")

	s := scanner.New(cfg.Exclusions.Paths)
	files, _ := s.Scan(dir)

	// vendor, node_modules, dist, and generated should all be excluded
	for _, f := range files {
		rel := strings.TrimPrefix(f, dir+string(filepath.Separator))
		if strings.HasPrefix(rel, "vendor/") ||
			strings.HasPrefix(rel, "node_modules/") ||
			strings.HasPrefix(rel, "dist/") ||
			strings.HasPrefix(rel, "internal/generated/") {
			t.Errorf("expected file to be excluded: %s", rel)
		}
	}
}

// ACCEPTANCE CRITERIA 2:
// "Files matching exclusion patterns are silently skipped during analysis"
func TestScanner_ExcludedFilesNotReturned(t *testing.T) {
	dir := setupTestDir(t)

	cfg := config.DefaultConfig()
	s := scanner.New(cfg.Exclusions.Paths)
	files, _ := s.Scan(dir)

	for _, f := range files {
		rel := strings.TrimPrefix(f, dir+string(filepath.Separator))
		// Default exclusions: vendor/, node_modules/, dist/
		if strings.HasPrefix(rel, "vendor/") {
			t.Errorf("vendor file should be excluded: %s", rel)
		}
		if strings.HasPrefix(rel, "node_modules/") {
			t.Errorf("node_modules file should be excluded: %s", rel)
		}
		if strings.HasPrefix(rel, "dist/") {
			t.Errorf("dist file should be excluded: %s", rel)
		}
	}
}

// ACCEPTANCE CRITERIA 3:
// "I receive a warning if I exclude a path that contains no files (misconfigured pattern)"
func TestScanner_WarnsOnEmptyExclusion(t *testing.T) {
	dir := setupTestDir(t)

	patterns := []string{"**/nonexistent/**"}
	s := scanner.New(patterns)
	_, warnings := s.Scan(dir)

	if len(warnings) == 0 {
		t.Error("expected warning for exclusion pattern that matches no files")
	}

	found := false
	for _, w := range warnings {
		if strings.Contains(w, "nonexistent") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning mentioning 'nonexistent', got: %v", warnings)
	}
}

// ACCEPTANCE CRITERIA 4:
// "Glob patterns support wildcard matching for nested directories"
func TestScanner_GlobPatternsMatchNestedDirs(t *testing.T) {
	dir := setupTestDir(t)

	// Create nested vendor structure
	mustWrite(t, dir, "a/b/c/vendor/deep.go", "package deep")

	patterns := []string{"**/vendor/**"}
	s := scanner.New(patterns)
	files, _ := s.Scan(dir)

	for _, f := range files {
		if strings.Contains(f, "/vendor/") {
			t.Errorf("expected nested vendor file to be excluded: %s", f)
		}
	}
}
